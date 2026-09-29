package scanner

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// RepoInfo holds information about a discovered git repository.
type RepoInfo struct {
	Path  string // absolute path to the repository
	Depth int    // depth from scan root
}

// MaxEntries is the maximum number of directories to scan before stopping.
const MaxEntries = 10000

// ScanRepositories recursively searches for .git directories within the given roots,
// up to the specified max depth. Returns a list of discovered repositories.
// The walk honours ctx: a cancelled scan (agent client disconnect, desktop
// "stop" button) aborts promptly instead of finishing a multi-minute walk
// for nobody — the returned error is then ctx.Err(), never a partial result
// presented as a complete one.
func ScanRepositories(ctx context.Context, roots []string, maxDepth int) ([]RepoInfo, error) {
	var repos []RepoInfo
	for _, root := range roots {
		if err := ctx.Err(); err != nil {
			return repos, err
		}
		found, err := scanRoot(ctx, root, maxDepth)
		if err != nil {
			if ctxErr := ctx.Err(); ctxErr != nil {
				return repos, ctxErr
			}
			// skip inaccessible roots
			continue
		}
		repos = append(repos, found...)
	}
	return repos, nil
}

// scanRoot scans a single root directory for git repositories.
func scanRoot(ctx context.Context, root string, maxDepth int) ([]RepoInfo, error) {
	var repos []RepoInfo
	entriesCount := 0

	// Check if root itself is a git repo
	rootGitDir := filepath.Join(root, ".git")
	if info, err := os.Stat(rootGitDir); err == nil && info.IsDir() {
		absPath, _ := filepath.Abs(root)
		repos = append(repos, RepoInfo{
			Path:  absPath,
			Depth: 0,
		})
	}

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			// permission denied, skip this entry
			if os.IsPermission(err) {
				if d != nil && d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			return nil
		}

		// Cancelled mid-walk: stop descending, surface the cancellation.
		if ctxErr := ctx.Err(); ctxErr != nil {
			return filepath.SkipAll
		}

		if !d.IsDir() {
			return nil
		}

		// Calculate depth relative to root
		relPath, _ := filepath.Rel(root, path)
		if relPath == "." {
			return nil
		}

		depth := len(strings.Split(filepath.ToSlash(relPath), "/"))

		// Stop if exceeding max depth
		if depth > maxDepth {
			return filepath.SkipDir
		}

		entriesCount++
		if entriesCount > MaxEntries {
			return filepath.SkipAll
		}

		// Check for .git directory
		gitDir := filepath.Join(path, ".git")
		if info, err := os.Stat(gitDir); err == nil && info.IsDir() {
			absPath, _ := filepath.Abs(path)
			repos = append(repos, RepoInfo{
				Path:  absPath,
				Depth: depth,
			})
			// Don't descend into git repositories
			return filepath.SkipDir
		}

		return nil
	})

	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			// Cancellation won the race: never present a partial walk as a
			// complete result.
			return repos, ctxErr
		}
		if !errors.Is(err, filepath.SkipAll) {
			return repos, err
		}
		// MaxEntries reached: keep the repos already found on this root
		// instead of silently dropping the whole root's results.
	}
	return repos, nil
}
