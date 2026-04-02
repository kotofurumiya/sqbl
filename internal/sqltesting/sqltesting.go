// Package sqltesting provides helpers for SQL string comparison in tests.
package sqltesting

import (
	"regexp"
	"strings"
)

var whitespacePattern = regexp.MustCompile(`[\s\n]+`)

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
