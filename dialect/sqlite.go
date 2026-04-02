package dialect

import "strings"

// SqliteDialect implements the Dialect interface for SQLite.
type SqliteDialect struct{}

var _ SqlDialect = &SqliteDialect{}

// Quote wraps a single identifier part in double quotes.
// It also escapes any existing double quotes within the identifier.
//
// Examples (SQLite):
//
//	users      -> "users"
//	my"table   -> "my""table"
func (s *SqliteDialect) Quote(str string) string {
	escaped := strings.ReplaceAll(str, "\"", "\"\"")
	return "\"" + escaped + "\""
}

// QuoteIdentifier quotes a full identifier (possibly dot-separated) by
// splitting it and quoting each part individually.
//
// Examples (SQLite):
//
//	users          -> "users"
//	main.users     -> "main"."users"
//	db.sch.table   -> "db"."sch"."table"
func (s *SqliteDialect) QuoteIdentifier(name string) string {
	parts := strings.Split(name, ".")
	for i, part := range parts {
		parts[i] = s.Quote(part)
	}
	return strings.Join(parts, ".")
}

// PlaceholderPositional returns the positional bind parameter string.
//
// Examples:
//
//	-> ?
func (s *SqliteDialect) PlaceholderPositional() string {
	return "?"
}

// PlaceholderIndexed returns the indexed bind parameter string for the given index.
// SQLite uses ? for all indexed placeholders, ignoring the index.
//
// Examples:
//
//	1 -> ?, 2 -> ?
func (s *SqliteDialect) PlaceholderIndexed(_ int) string {
	return "?"
}

// Bool converts a boolean value to its SQLite SQL representation.
// SQLite has no native boolean type; 1 represents true and 0 represents false.
//
// Examples:
//
//	true -> 1, false -> 0
func (s *SqliteDialect) Bool(b bool) string {
	if b {
		return "1"
	}
	return "0"
}
