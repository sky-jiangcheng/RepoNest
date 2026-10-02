package claude

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

const sessionFixture = `{"type":"mode","mode":"normal","sessionId":"sess-abc"}
{"type":"user","sessionId":"sess-abc","timestamp":"2026-09-23T08:00:00.000Z","cwd":"/Users/u/Workspace/Proj","gitBranch":"main","message":{"role":"user","content":[{"type":"text","text":"Refactor the parser"}]}}
{"type":"assistant","sessionId":"sess-abc","timestamp":"2026-09-23T08:01:00.000Z","cwd":"/Users/u/Workspace/Proj","gitBranch":"main","message":{"role":"assistant","content":[{"type":"thinking","thinking":"plan"},{"type":"tool_use","name":"Edit"}]}}
{"type":"assistant","sessionId":"sess-abc","timestamp":"2026-09-23T08:02:00.000Z","cwd":"/Users/u/Workspace/Proj","gitBranch":"main","isSidechain":true,"message":{"role":"assistant","content":[{"type":"text","text":"SUBAGENT SHOULD BE IGNORED"}]}}
{"type":"cost-state","sessionId":"sess-abc"}
not json at all, skipped leniently
{"type":"assistant","sessionId":"sess-abc","timestamp":"2026-09-23T08:03:00.000Z","cwd":"/Users/u/Workspace/Proj","gitBranch":"main","message":{"role":"assistant","content":[{"type":"tool_use","name":"Bash"},{"type":"tool_use","name":"Edit"},{"type":"text","text":"Parser refactored; tests green."}]}}
`

func TestParseSessionSummary(t *testing.T) {
	s, err := ParseSession(strings.NewReader(sessionFixture))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if s.SessionID != "sess-abc" {
		t.Errorf("sessionID = %q", s.SessionID)
	}
	if s.Cwd != "/Users/u/Workspace/Proj" {
		t.Errorf("cwd = %q", s.Cwd)
	}
	if s.GitBranch != "main" {
		t.Errorf("gitBranch = %q", s.GitBranch)
	}
	if s.FirstUserPrompt != "Refactor the parser" {
		t.Errorf("firstUserPrompt = %q", s.FirstUserPrompt)
	}
	if s.LastAssistantText != "Parser refactored; tests green." {
		t.Errorf("lastAssistantText = %q (sidechain leaked?)", s.LastAssistantText)
	}
	if s.LastTimestamp != "2026-09-23T08:03:00.000Z" {
		t.Errorf("lastTimestamp = %q", s.LastTimestamp)
	}
	// Edit appears twice -> deduped; order preserved; sidechain line ignored.
	want := []string{"Edit", "Bash"}
	if !reflect.DeepEqual(s.ToolsUsed, want) {
		t.Errorf("toolsUsed = %v, want %v", s.ToolsUsed, want)
	}
	if !s.HasContent() {
		t.Error("HasContent should be true")
	}
}

func TestParseSessionNoContent(t *testing.T) {
	s, err := ParseSession(strings.NewReader(`{"type":"mode","sessionId":"x"}` + "\n" + `{"type":"cost-state"}` + "\n"))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if s.HasContent() {
		t.Errorf("expected no content, got %+v", s)
	}
	if s.MessageCount != 0 {
		t.Errorf("MessageCount = %d, want 0", s.MessageCount)
	}
}

func TestParseSessionTruncatesLongText(t *testing.T) {
	long := strings.Repeat("字", maxRetainedRunes+500)
	fixture := `{"type":"assistant","sessionId":"s","timestamp":"2026-09-23T08:00:00.000Z","message":{"role":"assistant","content":[{"type":"text","text":"` + long + `"}]}}`
	s, err := ParseSession(strings.NewReader(fixture))
	if err != nil {
		t.Fatal(err)
	}
	if got := len([]rune(s.LastAssistantText)); got > maxRetainedRunes {
		t.Errorf("retained text = %d runes, exceeds cap %d", got, maxRetainedRunes)
	}
}

func TestLatestSessionFile(t *testing.T) {
	dir := t.TempDir()
	older := filepath.Join(dir, "019d0000-older.jsonl")
	newer := filepath.Join(dir, "019d9999-newer.jsonl")
	for _, p := range []string{older, newer} {
		if err := os.WriteFile(p, []byte("{}\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// A non-transcript file and a subdir must be ignored.
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "sub.jsonl"), 0o755); err != nil {
		t.Fatal(err)
	}
	base := time.Now()
	if err := os.Chtimes(older, base.Add(-2*time.Hour), base.Add(-2*time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(newer, base, base); err != nil {
		t.Fatal(err)
	}
	got, err := LatestSessionFile(dir)
	if err != nil {
		t.Fatalf("latest: %v", err)
	}
	if got != newer {
		t.Errorf("latest = %q, want %q", got, newer)
	}
}

func TestLatestSessionFileEmpty(t *testing.T) {
	dir := t.TempDir()
	got, err := LatestSessionFile(dir)
	if err != nil {
		t.Fatalf("err on empty dir: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}
