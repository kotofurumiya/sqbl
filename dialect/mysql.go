package dialect

import "strings"

// MysqlDialect implements the Dialect interface for MySQL.
type MysqlDialect struct{}

var _ SqlDialect = &MysqlDialect{}

// Quote wraps a single identifier part in backticks.
// It also escapes any existing backticks within the identifier.
//
// Examples (MySQL):
//
//	users      -> `users`
//	my`table   -> `my``table`
func (m *MysqlDialect) Quote(str string) string {
	escaped := strings.ReplaceAll(str, "`", "``")
	return "`" + escaped + "`"
}

// QuoteIdentifier quotes a full identifier (possibly dot-separated) by
// splitting it and quoting each part individually.
//
// Examples (MySQL):
//
//	users          -> `users`
//	public.users   -> `public`.`users`
//	db.sch.table   -> `db`.`sch`.`table`
func (m *MysqlDialect) QuoteIdentifier(name string) string {
	parts := strings.Split(name, ".")
	for i, part := range parts {
		parts[i] = m.Quote(part)
	}
	return strings.Join(parts, ".")
}

// PlaceholderPositional returns the positional bind parameter string.
//
// Examples:
//
//	-> ?
func (m *MysqlDialect) PlaceholderPositional() string {
	return "?"
}

// PlaceholderIndexed returns the indexed bind parameter string for the given index.
// MySQL uses ? for all indexed placeholders, ignoring the index.
//
// Examples:
//
//	1 -> ?, 2 -> ?
func (m *MysqlDialect) PlaceholderIndexed(_ int) string {
	return "?"
}

// Bool converts a boolean value to its MySQL SQL representation.
//
// Examples:
//
//	true -> TRUE, false -> FALSE
func (m *MysqlDialect) Bool(b bool) string {
	if b {
		return "TRUE"
	}
	return "FALSE"
}
