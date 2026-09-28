package db

import (
	"strings"
	"testing"
)

// The bounds are enforced at the DB layer because this package is the
// chokepoint every writer converges on — the service layer (desktop UI, MCP
// tools) AND the plugin runtime's import upserts, which call these functions
// directly. These tests prove the second path is bounded too.
func TestNoteWriteBoundsAtDBLayer(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	pid := createTestProject(t, db, "proj-a")

	// Sanity: normal-size writes must succeed, or the bounds are too tight.
	note, err := CreateNoteEx(db, pid, "title", "body", "tag", "other", "manual")
	if err != nil {
		t.Fatalf("CreateNoteEx: %v", err)
	}

	cases := []struct {
		name string
		fn   func() error
	}{
		{"CreateNoteEx oversized content", func() error {
			_, err := CreateNoteEx(db, pid, "t", strings.Repeat("x", MaxNoteContentLen+1), "", "knowledge", "manual")
			return err
		}},
		{"CreateNoteEx oversized title", func() error {
			_, err := CreateNoteEx(db, pid, strings.Repeat("t", MaxNoteTitleLen+1), "body", "", "knowledge", "manual")
			return err
		}},
		{"CreateNoteEx too many tags", func() error {
			_, err := CreateNoteEx(db, pid, "t", "body", strings.Repeat("a,", MaxNoteTagCount+1), "knowledge", "manual")
			return err
		}},
		{"UpdateNote oversized content", func() error {
			return UpdateNote(db, note.ID, strings.Repeat("x", MaxNoteContentLen+1))
		}},
		{"UpdateNoteFull oversized title", func() error {
			return UpdateNoteFull(db, note.ID, "body", strings.Repeat("t", MaxNoteTitleLen+1), "tag", "knowledge", false)
		}},
		{"UpdateNoteMeta oversized title", func() error {
			return UpdateNoteMeta(db, note.ID, strings.Repeat("t", MaxNoteTitleLen+1), "tag", "knowledge", false)
		}},
		{"UpdateNoteMeta too many tags", func() error {
			return UpdateNoteMeta(db, note.ID, "title", strings.Repeat("a,", MaxNoteTagCount+1), "knowledge", false)
		}},
	}
	for _, c := range cases {
		if err := c.fn(); err == nil {
			t.Errorf("%s: expected rejection", c.name)
		}
	}

	// The note must be untouched by the rejected writes.
	got, err := GetNoteByID(db, note.ID)
	if err != nil {
		t.Fatalf("GetNoteByID: %v", err)
	}
	if got.Title != "title" || got.Content != "body" || got.Tags != "tag" {
		t.Errorf("note mutated by rejected writes: %+v", got)
	}
}
