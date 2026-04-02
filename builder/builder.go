package builder

import (
	"strings"

	"github.com/kotofurumiya/sqbl/dialect"
)

// SqlBuilder is implemented by all SQL statement builders.
// It provides a common interface for rendering SQL statements.
type SqlBuilder interface {
	// ToSql renders the statement with a trailing semicolon using the builder's dialect.
	ToSql() string
	// ToSqlWithDialect renders the statement using the given dialect, without a trailing semicolon.
	ToSqlWithDialect(d dialect.SqlDialect) string
}

var _ SqlBuilder = (*SqlSelectBuilder)(nil)

// mapJoin transforms each element of s with fn and joins the results with sep.
// Returns "" when s is empty, so callers can use a single if-check.
//
//	mapJoin([]int{1, 2, 3}, ", ", strconv.Itoa)  →  "1, 2, 3"
//	mapJoin([]int{},         ", ", strconv.Itoa)  →  ""
func mapJoin[T any](s []T, sep string, fn func(T) string) string {
	if len(s) == 0 {
		return ""
	}
	strs := make([]string, len(s))
	for i, v := range s {
		strs[i] = fn(v)
	}
	return strings.Join(strs, sep)
}

// quoteIdentifiers quotes every name with the dialect and joins them with ", ".
// Convenience wrapper around mapJoin for the most common identifier-list pattern.
//
//	quoteIdentifiers(d, []string{"id", "name"})  →  `"id", "name"`
//	quoteIdentifiers(d, nil)                     →  ""
func quoteIdentifiers(d dialect.SqlDialect, names []string) string {
	return mapJoin(names, ", ", d.QuoteIdentifier)
}
