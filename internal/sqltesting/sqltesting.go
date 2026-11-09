// Package sqltesting provides helpers for SQL string comparison in tests.
package sqltesting

import (
	"regexp"
	"strings"
	"testing"
)

var whitespacePattern = regexp.MustCompile(`\s+`)

// Compact collapses consecutive whitespace and newlines into a single space,
// making multi-line SQL strings easy to compare against generated output.
//
//	sqltesting.Compact(`
//	    SELECT id
//	    FROM users
//	`) // → "SELECT id FROM users"
func Compact(sql string) string {
	return whitespacePattern.ReplaceAllString(strings.TrimSpace(sql), " ")
}

// AssertSQL reports a test error when got differs from want.
// want is compacted first, so it may be written across multiple indented lines;
// got is compared as-is, so stray whitespace in generated SQL is still caught.
// Break lines in want only where the expected SQL has a single space.
func AssertSQL(t testing.TB, got, want string) {
	t.Helper()
	if w := Compact(want); got != w {
		t.Errorf("SQL mismatch\n got: %s\nwant: %s", got, w)
	}
}
