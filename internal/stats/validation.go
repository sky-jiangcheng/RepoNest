package stats

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// datePattern validates YYYY-MM-DD format.
var datePattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

// spacesRe collapses runs of whitespace, used to normalise git's space-padded
// default author-date layout before parsing.
var spacesRe = regexp.MustCompile(`\s+`)

// safeAuthorPattern allows only safe characters for git author matching.
// Allowed: letters, digits, space, dot, underscore, hyphen, at-sign.
var safeAuthorPattern = regexp.MustCompile(`^[a-zA-Z0-9 ._\-@]+$`)

// ValidateDate checks that the date string matches YYYY-MM-DD format.
func ValidateDate(date string) error {
	if !datePattern.MatchString(date) {
		return fmt.Errorf("invalid date format: %s (expected YYYY-MM-DD)", date)
	}
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return fmt.Errorf("invalid date: %s", date)
	}
	// Ensure the parsed date matches the input (catches things like "0000-00-00")
	if t.Format("2006-01-02") != date {
		return fmt.Errorf("invalid date: %s", date)
	}
	return nil
}

// ValidateAuthor checks that the author string contains only safe characters.
func ValidateAuthor(author string) error {
	if author == "" {
		return nil
	}
	if !safeAuthorPattern.MatchString(author) {
		return fmt.Errorf("invalid author name: contains unsafe characters")
	}
	return nil
}

func isHex(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}
	return true
}

// parseTimestamp converts a commit-time string into unix seconds. It is used
// only for latest-commit comparison, so an unrecognised value returns 0 rather
// than surfacing an error (preserving the previous lenient behaviour).
//
// Accepted formats (P29): bare unix epoch seconds (what git %at yields), the
// app's internal "2006-01-02 15:04:05", RFC 3339 / strict ISO 8601 (git %aI /
// %cI, with or without a zone offset), git's "%ai" form ("2006-01-02 15:04:05
// -0700"), date-only "2006-01-02" (git %ad --date=short), and git's default
// author-date layout ("Mon Jan _2 15:04:05 2006 -0700", git %ad without --date).
func parseTimestamp(s string) int64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	// A bare integer is already unix seconds — no re-parsing needed.
	if isAllDigits(s) {
		if n, err := strconv.ParseInt(s, 10, 64); err == nil {
			return n
		}
		return 0
	}
	layouts := []string{
		time.RFC3339,                    // 2006-01-02T15:04:05Z07:00 (git %aI/%cI, ISO 8601)
		"2006-01-02T15:04:05",           // ISO 8601 without a zone
		"2006-01-02 15:04:05 -0700",     // git %ai (space separator, zone offset)
		"2006-01-02 15:04:05",           // app internal format / git %ai without zone
		"2006-01-02",                    // date-only / git %ad --date=short
		"Mon Jan 2 15:04:05 2006 -0700", // git default author date %ad (with zone)
		"Mon Jan 2 15:04:05 2006",       // git default author date, no zone
	}
	// git's default layout space-pads the day ("Jan  3"), which time.Parse's
	// "Jan 2" pattern does not match; collapse runs of spaces before trying.
	collapsed := spacesRe.ReplaceAllString(s, " ")
	for _, candidate := range []string{s, collapsed} {
		for _, l := range layouts {
			if t, err := time.Parse(l, candidate); err == nil {
				return t.Unix()
			}
		}
	}
	return 0
}

// isAllDigits reports whether s is a non-empty run of ASCII digits.
func isAllDigits(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return s != ""
}

func extractBranch(refString string) string {
	if strings.Contains(refString, "HEAD -> ") {
		parts := strings.Split(refString, "HEAD -> ")
		if len(parts) > 1 {
			b := strings.Split(parts[1], ",")[0]
			return strings.TrimSpace(b)
		}
	}
	return ""
}
