package service

import (
	"strings"
	"testing"
	"unicode/utf8"

	"repo-nest/internal/importers/claude"
)

func TestHandoffFromSessionFillsSections(t *testing.T) {
	sess := claude.Session{
		SessionID:         "abcdef123456789",
		GitBranch:         "main",
		LastAssistantText: "Refactored the parser; tests green.",
		ToolsUsed:         []string{"Edit", "Bash"},
	}
	in := handoffFromSession(sess)
	if in.Agent != "claude-code" {
		t.Errorf("agent = %q", in.Agent)
	}
	if in.Tags != "auto-captured" {
		t.Errorf("tags = %q", in.Tags)
	}
	if in.Summary != "Refactored the parser; tests green." {
		t.Errorf("summary = %q", in.Summary)
	}
	// Must satisfy CreateHandoffNote's "at least one section" rule.
	if len(in.Changes) == 0 {
		t.Fatal("Changes empty; handoff would be rejected")
	}
	joined := strings.Join(in.Changes, "\n")
	if !strings.Contains(joined, "abcdef12") || !strings.Contains(joined, "main") {
		t.Errorf("changes header missing id/branch: %q", joined)
	}
	if !strings.Contains(joined, "使用工具 Edit") || !strings.Contains(joined, "使用工具 Bash") {
		t.Errorf("changes missing tools: %q", joined)
	}
}

func TestHandoffFromSessionFallsBackToPrompt(t *testing.T) {
	// No assistant text: summary falls back to the first prompt, and Changes
	// still carries the header line so the handoff is valid.
	in := handoffFromSession(claude.Session{FirstUserPrompt: "Build the cache", SessionID: "sid"})
	if in.Summary != "Build the cache" {
		t.Errorf("summary fallback = %q", in.Summary)
	}
	if len(in.Changes) == 0 {
		t.Error("expected a header change line even with no tools")
	}
}

func TestHandoffFromSessionCapsToolItems(t *testing.T) {
	tools := make([]string, 50)
	for i := range tools {
		tools[i] = "t" + string(rune('a'+i%26)) + strings.Repeat("x", i)
	}
	in := handoffFromSession(claude.Session{LastAssistantText: "s", ToolsUsed: tools})
	if len(in.Changes) > maxHandoffItems {
		t.Fatalf("changes %d exceeds max %d", len(in.Changes), maxHandoffItems)
	}
	if !strings.Contains(strings.Join(in.Changes, "\n"), "共 50 种工具") {
		t.Error("expected overflow summary bullet")
	}
}

func TestTruncateBytesRuneBoundary(t *testing.T) {
	cjk := strings.Repeat("汉", 100) // 3 bytes each = 300 bytes
	got := truncateBytes(cjk, 10)   // 10 is not a rune boundary
	if len(got) > 10 {
		t.Fatalf("len = %d > max 10", len(got))
	}
	if !utf8.ValidString(got) {
		t.Fatalf("cut produced invalid UTF-8: %q", got)
	}
	if len(got)%3 != 0 {
		t.Errorf("expected whole runes, got %d bytes", len(got))
	}
	// Under max is a no-op.
	if truncateBytes("abc", 10) != "abc" {
		t.Error("short string should be unchanged")
	}
}
