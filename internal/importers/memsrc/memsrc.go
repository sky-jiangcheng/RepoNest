// Package memsrc holds the small, source-agnostic helpers shared by the
// built-in memory importers (claude, codex, ...). Extracted when the second
// importer landed (ADR-0011 决策 2): each importer supplies its own file
// discovery and document shaping, while project matching and bounded reads
// live here so they cannot drift apart across sources.
package memsrc

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"repo-nest/internal/db"
)

// MatchProject finds the RepoNest project id for a display name, preferring an
// exact name, then a repository path ending with /name (or /name.git), then a
// project-name containment. Returns 0 when nothing matches; importers set
// ImportDoc.ProjectID = 0 in that case and the runtime counts it as skipped.
func MatchProject(displayName string, projects []db.Project, repos []db.Repository) int64 {
	lower := strings.ToLower(displayName)
	if lower == "" {
		return 0
	}
	// 1. exact project name
	for _, p := range projects {
		if p.Name == displayName {
			return p.ID
		}
	}
	// 2. repository path ending with /displayName
	for _, r := range repos {
		rp := strings.ToLower(r.Path)
		if strings.HasSuffix(rp, "/"+lower) || strings.HasSuffix(rp, "/"+lower+".git") {
			if r.ProjectID != nil {
				return *r.ProjectID
			}
		}
	}
	// 3. project name containment
	for _, p := range projects {
		if strings.Contains(strings.ToLower(p.Name), lower) {
			return p.ID
		}
	}
	return 0
}

// ReadCapped reads at most limit bytes from path. A file larger than the limit
// returns exactly limit bytes with no error, so the caller can distinguish
// "oversized" from "unreadable" and skip the file with a clear message rather
// than loading a multi-GB transcript fully into memory.
func ReadCapped(path string, limit int) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(io.LimitReader(f, int64(limit)))
}

// LastPathSegment returns the final element of a filesystem path:
// "/Users/me/Workspace/Foo" -> "Foo". Trailing slashes are trimmed. Returns ""
// for empty input.
func LastPathSegment(p string) string {
	p = strings.TrimRight(p, "/")
	if p == "" {
		return ""
	}
	return filepath.Base(p)
}
