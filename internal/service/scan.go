package service

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"repo-nest/internal/db"
	"repo-nest/internal/grouper"
	"repo-nest/internal/platform"
	"repo-nest/internal/scanner"
)

// ScanResult holds the result of a scan operation.
type ScanResult struct {
	Success    bool   `json:"success"`
	ReposFound int    `json:"repos_found"`
	Projects   int    `json:"projects"`
	// SyncErrors counts project groups whose DB sync failed while others
	// succeeded. A partial sync used to report plain success, leaving the
	// agent to believe the knowledge base was complete.
	SyncErrors int    `json:"sync_errors,omitempty"`
	TaskID     string `json:"task_id,omitempty"`
}

// ScanStatus holds the current scanning progress.
type ScanStatus struct {
	Running     bool   `json:"running"`
	Backfilling bool   `json:"backfilling"`
	Message     string `json:"message"`
	Progress    int    `json:"progress"`
	Total       int    `json:"total"`
}

// EnsureDefaultScanRoots seeds the platform default scan roots on first run,
// tracked by the scan_roots_seeded config flag so a user who deliberately
// removes every root is not re-seeded on the next launch.
//
// Shared by the desktop app and the MCP server: a headless (MCP-only) install
// must be able to discover repositories without ever opening the GUI, which is
// what turns the cold start from install→scan→star→refresh into install→scan.
func (s *Service) EnsureDefaultScanRoots() {
	if seeded, _ := db.GetConfig(s.db, "scan_roots_seeded"); seeded != "" {
		return
	}
	// Never clobber roots that are already configured: seeding is a first-run
	// convenience, not a reset. A database with roots but no flag (set by an
	// older build, or by configuration) keeps what it has.
	if existing, _ := db.GetScanRoots(s.db); len(existing) > 0 {
		_ = db.SetConfig(s.db, "scan_roots_seeded", "1")
		return
	}
	if defaults := platform.DefaultScanRoots(); len(defaults) > 0 {
		if err := db.ReplaceScanRoots(s.db, defaults); err != nil {
			log.Printf("seed default scan roots error: %v", err)
		} else {
			log.Printf("seeded %d default scan root(s)", len(defaults))
		}
	}
	_ = db.SetConfig(s.db, "scan_roots_seeded", "1")
}

// ScanNow runs the full scan pipeline synchronously and reports how much it
// found. Unlike TriggerScan (fire-and-forget, for the desktop UI's background
// job) this blocks until discovery and stats refresh finish, so an AI agent
// gets a deterministic answer from a single tool call.
//
// Concurrency is guarded by the same flag TriggerScan uses: a headless scan
// must not race an in-flight desktop scan. ctx is the caller's (MCP request)
// context, so a client disconnect cancels the scan instead of burning CPU on
// work nobody is waiting for.
func (s *Service) ScanNow(ctx context.Context) (*ScanResult, error) {
	s.scanMu.Lock()
	if s.scanning {
		s.scanMu.Unlock()
		return nil, fmt.Errorf("scan already in progress")
	}
	s.scanning = true
	s.scanMu.Unlock()

	defer func() {
		s.scanMu.Lock()
		s.scanning = false
		s.scanProgress = 0
		s.scanTotal = 0
		s.scanMu.Unlock()
	}()

	res, err := s.runCollectedScan(ctx)
	if err != nil {
		return nil, err
	}
	res.Success = true
	return &res, nil
}

// TriggerScan starts an async full repository scan and returns immediately.
// Stats are refreshed for every collected project, including the ones this
// scan marks as collected.
func (s *Service) TriggerScan() (*ScanResult, error) {
	s.scanMu.Lock()
	if s.scanning {
		s.scanMu.Unlock()
		return nil, fmt.Errorf("scan already in progress")
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.scanCancel = cancel
	s.scanning = true
	s.scanMu.Unlock()

	taskID := fmt.Sprintf("%d", time.Now().UnixNano())

	go func() {
		if _, err := s.runCollectedScan(ctx); err != nil {
			// Cancellation via CancelScan is the expected path: the desktop
			// user pressed "stop". Anything else is a real scan failure that
			// used to vanish into runCollectedScan's log-only error handling.
			if ctx.Err() == nil {
				log.Printf("async scan failed: %v", err)
			}
		}
		s.scanMu.Lock()
		s.scanning = false
		s.scanProgress = 0
		s.scanTotal = 0
		s.scanCancel = nil
		s.scanMu.Unlock()
	}()

	return &ScanResult{Success: true, TaskID: taskID}, nil
}

// GetScanStatus returns the current scan progress.
func (s *Service) GetScanStatus() *ScanStatus {
	s.scanMu.Lock()
	running := s.scanning
	backfilling := s.scanBackfilling
	progress := s.scanProgress
	total := s.scanTotal
	s.scanMu.Unlock()
	msg := ""
	if running {
		if total > 0 {
			msg = fmt.Sprintf("Scanning %d/%d...", progress, total)
		} else {
			msg = "Scanning..."
		}
	}
	return &ScanStatus{Running: running, Backfilling: backfilling, Message: msg, Progress: progress, Total: total}
}

// runCollectedScan is the single scan pipeline: filesystem scan → grouping →
// transactional sync of projects/repos → stale cleanup → stats refresh for
// collected projects. Collected ids are re-read after the transaction commits
// so projects discovered by this scan are included. It returns how many
// repositories were found and how many projects they grouped into, so a
// synchronous caller (ScanNow) can report the outcome.
//
// Hard failures (filesystem scan, transaction lifecycle) are returned as
// errors instead of being logged and swallowed: ScanNow must be able to tell
// an agent "the scan failed" rather than reporting success with 0 repos.
// Per-group sync errors stay non-fatal (one bad path must not abort the whole
// scan) but are counted into ScanResult.SyncErrors, so a partial sync never
// masquerades as a clean one; a panic anywhere in the pipeline is converted
// into an error too, because a recovered panic that reports success would be
// a lie.
func (s *Service) runCollectedScan(ctx context.Context) (out ScanResult, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("panic in collected scan: %v", r)
			out = ScanResult{}
			err = fmt.Errorf("scan panicked: %v", r)
		}
	}()

	depthStr, _ := db.GetConfig(s.db, "scan_depth")
	maxDepth, _ := strconv.Atoi(depthStr)
	if maxDepth <= 0 || maxDepth > 2 {
		maxDepth = 2
	}

	roots, _ := db.GetScanRoots(s.db)
	repos, err := scanner.ScanRepositories(ctx, roots, maxDepth)
	if err != nil {
		// Cancellation or a genuine walk failure — never present a partial
		// filesystem result as a complete scan.
		return ScanResult{ReposFound: len(repos)}, fmt.Errorf("scan repositories: %w", err)
	}

	// Group discovered repositories into projects. Repositories returned by
	// the scanner are filesystem paths without a DB id, so they are all
	// grouped and then synced.
	groups := grouper.GroupRepositories(repos)

	s.scanMu.Lock()
	s.scanTotal = len(groups)
	s.scanProgress = 0
	s.scanMu.Unlock()

	scannedPaths := make([]string, 0, len(repos))
	for _, r := range repos {
		scannedPaths = append(scannedPaths, r.Path)
	}

	tx, err := s.db.Begin()
	if err != nil {
		return ScanResult{ReposFound: len(repos)}, fmt.Errorf("begin scan transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	syncErrs := 0
	for i, group := range groups {
		select {
		case <-ctx.Done():
			return ScanResult{ReposFound: len(repos)}, ctx.Err()
		default:
		}
		s.scanMu.Lock()
		s.scanProgress = i + 1
		s.scanMu.Unlock()

		projectID, err := db.SyncProjectTx(tx, group.Name, group.RootPath, 0, group.IsAutoGrouped)
		if err != nil {
			log.Printf("sync project error: %v", err)
			syncErrs++
			continue
		}
		// Participating in a controlled scan makes the project "collected":
		// it becomes eligible for the history backfill after the scan.
		if err := db.MarkProjectCollectedTx(tx, projectID); err != nil {
			log.Printf("mark project collected error: %v", err)
		}
		for _, repo := range group.Repos {
			if err := db.UpsertRepositoryTx(tx, repo.Path, projectID); err != nil {
				log.Printf("upsert repo error: %v", err)
			}
		}
	}
	// Every group failing means the sync did nothing useful; reporting
	// "found N projects" would tell the agent the knowledge base is ready
	// when it is empty.
	if syncErrs > 0 && syncErrs == len(groups) {
		return ScanResult{}, fmt.Errorf("scan sync failed for all %d discovered project group(s)", len(groups))
	}

	if err := db.CleanupStaleDataTx(tx, scannedPaths); err != nil {
		return ScanResult{ReposFound: len(repos)}, fmt.Errorf("cleanup stale data: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return ScanResult{ReposFound: len(repos)}, fmt.Errorf("commit scan transaction: %w", err)
	}

	// Re-read after commit: projects found by this scan were just marked
	// collected, so they now show up and get their history backfilled in the
	// same run instead of waiting for a second scan.
	collectedIDs, err := db.GetCollectedProjectIDs(ctx, s.db)
	if err != nil {
		// The scan itself is committed at this point; only the stats refresh
		// is skipped, so report the real counts with a warning-level log
		// instead of failing the whole call.
		log.Printf("load collected projects error: %v", err)
		return ScanResult{ReposFound: len(repos), Projects: len(groups), SyncErrors: syncErrs}, nil
	}
	s.refreshCollectedStats(ctx, collectedIDs)
	_ = db.SetConfig(s.db, "last_stats_backfill", s.git.GetTodayDate())
	log.Printf("scan complete: %d repos, %d projects (%d sync errors)", len(repos), len(groups), syncErrs)
	if s.rt != nil {
		s.rt.Emit("project.scanned", map[string]any{
			"repos_found": len(repos), "projects": len(groups),
		})
	}
	return ScanResult{ReposFound: len(repos), Projects: len(groups), SyncErrors: syncErrs}, nil
}
