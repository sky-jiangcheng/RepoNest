package integrity

import (
	"strings"
	"testing"
)

// Regression tests for two defects that each needed a COMBINED database state
// to surface: an unassigned repository *and* real orphans in the same database,
// and a version stamp behind the build *and* a matching shape gap. Every
// earlier test covered each half separately, so both bugs passed CI.
//
// The pattern they share: a benign signal and a serious one present at the
// same time, where a short-circuit on the benign one hid the serious one. That
// is why each test asserts on the human-readable Detail as well as the status
// - a receipt whose prose contradicts its own metrics is worse than no
// receipt, because it is believed.
func TestOrphanNotDowngradedWhenReposUnassigned(t *testing.T) {
	database := setupDB(t)
	// One repository, never attached to a project (a legitimate transient
	// state: indexed before the grouper runs).
	mustExec(t, database, "INSERT INTO repositories (id, path) VALUES (1,'/tmp/r')")
	// Real referential corruption alongside it: notes/todos pointing at a
	// project that does not exist.
	mustExec(t, database, "PRAGMA foreign_keys = OFF")
	mustExec(t, database, "INSERT INTO project_notes (id, project_id, title, content) VALUES (1, 999, 'a', 'a')")
	mustExec(t, database, "INSERT INTO project_todos (id, project_id, title) VALUES (1, 999, 'b')")
	mustExec(t, database, "PRAGMA foreign_keys = ON")

	c := CheckOrphanRows(database)
	t.Logf("status=%v", c.Status)
	t.Logf("detail=%s", c.Detail)
	t.Logf("metrics=%v", c.Metrics)
	t.Logf("offenders=%v", c.Offenders)

	if c.Status != StatusFail {
		t.Fatalf("real corruption downgraded: want StatusFail, got %v", c.Status)
	}
	if len(c.Offenders) == 0 {
		t.Fatal("offenders were dropped — the receipt loses its evidence")
	}
	// The headline sentence must not deny what the metrics report.
	for _, bad := range []string{"no dangling project_id", "every row resolves"} {
		if strings.Contains(c.Detail, bad) {
			t.Fatalf("detail contradicts its own metrics: %q", c.Detail)
		}
	}
	if !strings.Contains(c.Detail, "dangling project_id") {
		t.Fatalf("detail should name the corruption, got %q", c.Detail)
	}
}

func TestBehindDBNotAccusedOfLying(t *testing.T) {
	database := setupDB(t)
	// An honestly-behind database: the stamp is v3 and v8/v10 shape is really
	// absent. That is a normal backup, not a corrupted file.
	mustExec(t, database, "UPDATE app_config SET value = '3' WHERE key = 'schema_version'")
	// note_versions.note_id cannot be dropped (SQLite refuses a column that a
	// foreign key references), so v10's shape gap stands in for the pending
	// migrations' missing columns.
	mustExec(t, database, "ALTER TABLE repo_meta DROP COLUMN dependencies")

	c := CheckSchemaVersion(database)
	t.Logf("status=%v", c.Status)
	t.Logf("detail=%s", c.Detail)
	t.Logf("metrics=%v", c.Metrics)

	for _, bad := range []string{"recorded without being applied", "does not match the actual tables"} {
		if strings.Contains(c.Detail, bad) {
			t.Fatalf("honest lag misdiagnosed as corruption: %q", c.Detail)
		}
	}
	pending, ok := c.Metrics["pending"]
	if !ok {
		t.Fatal("pending metric missing — the user cannot see how far behind they are")
	}
	if pending != int64(ExpectedSchemaVersion)-3 {
		t.Fatalf("pending=%d, want %d", pending, ExpectedSchemaVersion-3)
	}
	if !strings.Contains(c.Detail, "migration(s) have not run") {
		t.Fatalf("detail should say how far behind, got %q", c.Detail)
	}
	// The concrete gaps are still worth naming, framed as consequences.
	if !strings.Contains(c.Detail, "confirmed missing") {
		t.Fatalf("detail should still name the concrete gaps, got %q", c.Detail)
	}
}

func TestStampAheadWithMissingShapeStillAccused(t *testing.T) {
	database := setupDB(t)
	// The accusation is still correct when the stamp is NOT behind: here it
	// claims to be current while the shape disagrees.
	mustExec(t, database, "ALTER TABLE repo_meta DROP COLUMN dependencies")

	c := CheckSchemaVersion(database)
	t.Logf("status=%v", c.Status)
	t.Logf("detail=%s", c.Detail)
	if !strings.Contains(c.Detail, "recorded without being applied") {
		t.Fatalf("the genuine mismatch diagnosis was lost: %q", c.Detail)
	}
}
