package service

import (
	"fmt"
	"log"
	"strings"

	"reponest/internal/db"
	"reponest/internal/domain"
)

// Write-size guards for notes. Every writer (desktop UI, MCP tools, plugin
// imports) goes through the service layer, so validating here covers all
// entry paths at once. A single unbounded write from an agent would
// otherwise bloat the context document (reponest_context embeds notes) and
// the SQLite row itself.
const (
	maxNoteContentLen = 100_000 // ~100 KB of Markdown per note
	maxNoteTitleLen   = 200
	maxNoteTagsLen    = 500
	maxNoteTagCount   = 20
)

// validateNoteBounds rejects oversized note fields before they reach the
// database, with limits generous enough for real Markdown notes.
func validateNoteBounds(title, content, tags string) error {
	if len(content) > maxNoteContentLen {
		return fmt.Errorf("content too long: %d bytes (max %d)", len(content), maxNoteContentLen)
	}
	if len(title) > maxNoteTitleLen {
		return fmt.Errorf("title too long: %d bytes (max %d)", len(title), maxNoteTitleLen)
	}
	if len(tags) > maxNoteTagsLen {
		return fmt.Errorf("tags too long: %d bytes (max %d)", len(tags), maxNoteTagsLen)
	}
	if n := countTags(tags); n > maxNoteTagCount {
		return fmt.Errorf("too many tags: %d (max %d)", n, maxNoteTagCount)
	}
	return nil
}

// countTags counts comma-separated tag entries, ignoring blanks so ",a,,b,"
// counts as two tags.
func countTags(tags string) int {
	n := 0
	for _, t := range strings.Split(tags, ",") {
		if strings.TrimSpace(t) != "" {
			n++
		}
	}
	return n
}

// ListNotes returns all notes for a project.
func (s *Service) ListNotes(projectID int64) []domain.Note {
	notes, err := db.ListNotes(s.db, projectID)
	if err != nil {
		log.Printf("list notes error: %v", err)
		return nil
	}
	if notes == nil {
		notes = []domain.Note{}
	}
	return notes
}

// CreateNote creates a new note for a project.
func (s *Service) CreateNote(projectID int64, content string) (*domain.Note, error) {
	if strings.TrimSpace(content) == "" {
		return nil, fmt.Errorf("content is required")
	}
	if err := validateNoteBounds("", content, ""); err != nil {
		return nil, err
	}
	note, err := db.CreateNote(s.db, projectID, content)
	if err == nil && s.rt != nil {
		s.rt.Emit("note.created", map[string]any{
			"id": note.ID, "project_id": projectID, "content": content,
		})
	}
	return note, err
}

// CreateNoteWithMeta creates a note with explicit title, tags, kind and source.
func (s *Service) CreateNoteWithMeta(projectID int64, title, content, tags, kind, source string) (*domain.Note, error) {
	if strings.TrimSpace(content) == "" {
		return nil, fmt.Errorf("content is required")
	}
	if err := validateNoteBounds(title, content, tags); err != nil {
		return nil, err
	}
	note, err := db.CreateNoteEx(s.db, projectID, title, content, tags, kind, source)
	if err == nil && s.rt != nil {
		s.rt.Emit("note.created", map[string]any{
			"id": note.ID, "project_id": projectID, "title": title, "content": content, "tags": tags, "kind": kind,
		})
	}
	return note, err
}

// UpdateNote updates the content of a note.
func (s *Service) UpdateNote(noteID int64, content string) error {
	if strings.TrimSpace(content) == "" {
		return fmt.Errorf("content is required")
	}
	if err := validateNoteBounds("", content, ""); err != nil {
		return err
	}
	return db.UpdateNote(s.db, noteID, content)
}

// UpdateNoteFull updates both content and metadata in a single transaction,
// avoiding the version-snapshot inconsistency that occurs when content and
// metadata are updated in separate calls.
func (s *Service) UpdateNoteFull(noteID int64, content, title, tags, kind string, pinned bool) error {
	if strings.TrimSpace(content) == "" {
		return fmt.Errorf("content is required")
	}
	if err := validateNoteBounds(title, content, tags); err != nil {
		return err
	}
	return db.UpdateNoteFull(s.db, noteID, content, title, tags, kind, pinned)
}

// DeleteNote removes a note.
func (s *Service) DeleteNote(noteID int64) error {
	return db.DeleteNote(s.db, noteID)
}

// UpdateNoteMeta updates a note's editable metadata (title, tags, kind, pinned).
func (s *Service) UpdateNoteMeta(noteID int64, title, tags, kind string, pinned bool) error {
	if err := validateNoteBounds(title, "", tags); err != nil {
		return err
	}
	return db.UpdateNoteMeta(s.db, noteID, title, tags, kind, pinned)
}

// PinNote sets or clears the pinned flag on a note.
func (s *Service) PinNote(noteID int64, pinned bool) error {
	return db.PinNote(s.db, noteID, pinned)
}

// MoveNote reassigns a note to a different project.
func (s *Service) MoveNote(noteID, projectID int64) error {
	return db.MoveNote(s.db, noteID, projectID)
}

// ListAllNotes returns every note across all projects, joined with project
// info, ordered pinned first then most recently updated. Unbounded: this is
// the desktop UI's full-list path. Agent-facing callers must use
// ListAllNotesLimited so a large knowledge base is never fully loaded.
func (s *Service) ListAllNotes() []domain.NoteWithProject {
	return s.listAllNotes(0)
}

// ListAllNotesLimited returns the most recent notes across all projects,
// capped at limit (<= 0 falls back to the default). Used by MCP tools and
// llms.txt generation, whose output is bounded anyway: pulling every note
// with its full content to display a few dozen is pure waste.
func (s *Service) ListAllNotesLimited(limit int) []domain.NoteWithProject {
	if limit <= 0 {
		limit = defaultAllNotesLimit
	}
	return s.listAllNotes(limit)
}

// defaultAllNotesLimit matches reponest_notes_list's documented default.
const defaultAllNotesLimit = 50

func (s *Service) listAllNotes(limit int) []domain.NoteWithProject {
	notes, err := db.ListAllNotes(s.db, limit, "")
	if err != nil {
		log.Printf("list all notes error: %v", err)
		return nil
	}
	if notes == nil {
		notes = []domain.NoteWithProject{}
	}
	return notes
}

// CountNotes returns the total number of notes across all projects without
// loading them, for health/score checks that only need the number.
func (s *Service) CountNotes() int {
	n, err := db.CountNotes(s.db)
	if err != nil {
		log.Printf("count notes error: %v", err)
		return 0
	}
	return n
}

// ListAllTags returns the distinct set of tags used across all notes.
func (s *Service) ListAllTags() []string {
	tags, err := db.ListAllTags(s.db)
	if err != nil {
		log.Printf("list all tags error: %v", err)
		return nil
	}
	if tags == nil {
		tags = []string{}
	}
	return tags
}

// ListNoteVersions returns the recent version history for a note.
func (s *Service) ListNoteVersions(noteID int64) []domain.NoteVersion {
	versions, err := db.ListNoteVersions(s.db, noteID)
	if err != nil {
		log.Printf("list note versions error: %v", err)
		return nil
	}
	if versions == nil {
		return []domain.NoteVersion{}
	}
	return versions
}

// RestoreNoteVersion restores a note to the content of a previous version.
func (s *Service) RestoreNoteVersion(noteID, versionID int64) error {
	return db.RestoreNoteVersion(s.db, noteID, versionID)
}

// DiffNoteVersions returns a line-based diff between a version and the
// current note.
func (s *Service) DiffNoteVersions(noteID, versionID int64) (string, error) {
	return db.DiffNoteVersions(s.db, noteID, versionID)
}

// GetNote returns a single note by ID for read-only consumers (CLI / MCP).
func (s *Service) GetNote(noteID int64) (*domain.Note, error) {
	return db.GetNoteByID(s.db, noteID)
}
