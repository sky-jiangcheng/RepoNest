package memsrc

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestClipToBytes(t *testing.T) {
	if got := ClipToBytes("abc", 10); got != "abc" {
		t.Errorf("short passthrough = %q", got)
	}
	if got := ClipToBytes("abcdef", 3); got != "abc" {
		t.Errorf("ascii clip = %q", got)
	}
	// Each CJK rune is 3 bytes; a 4-byte cap must keep exactly one whole rune.
	cjk := strings.Repeat("漢", 5)
	got := ClipToBytes(cjk, 4)
	if !utf8.ValidString(got) {
		t.Fatalf("clip split a rune (invalid UTF-8): %q", got)
	}
	if len(got) != 3 {
		t.Fatalf("clip length = %d bytes, want 3 (one rune)", len(got))
	}
	if got := ClipToBytes("", 5); got != "" {
		t.Errorf("empty = %q", got)
	}
	if got := ClipToBytes("abc", 0); got != "" {
		t.Errorf("max<=0 should be empty, got %q", got)
	}
}
