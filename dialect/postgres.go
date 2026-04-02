package dialect

import (
	"fmt"
	"strings"
)

// PostgresDialect implements the Dialect interface for PostgreSQL.
type PostgresDialect struct{}

var _ SqlDialect = &PostgresDialect{}

// Quote wraps a single identifier part in the dialect-specific quote character.
// It also escapes any existing quote characters within the identifier.
//
// Examples (PostgreSQL):
//
//	users      -> "users"
//	my"table   -> "my""table"
func (p *PostgresDialect) Quote(s string) string {
	escaped := strings.ReplaceAll(s, "\"", "\"\"")
	return "\"" + escaped + "\""
}

// QuoteIdentifier quotes a full identifier (possibly dot-separated) by
// splitting it and quoting each part individually.
//
// Examples (PostgreSQL):
//
//	users          -> "users"
//	public.users   -> "public"."users"
//	db.sch.table   -> "db"."sch"."table"
func (p *PostgresDialect) QuoteIdentifier(name string) string {
	parts := strings.Split(name, ".")
	for i, part := range parts {
		// Quote each part (e.g., schema, table) individually to maintain structure.
		parts[i] = p.Quote(part)
	}
	return strings.Join(parts, ".")
}

// PlaceholderPositional returns the positional bind parameter string.
//
// Examples:
//
//	-> ?
func (p *PostgresDialect) PlaceholderPositional() string {
	return "?"
}

// PlaceholderIndexed returns the indexed bind parameter string for the given index.
// Index starts from 1.
//
// Examples:
//
//	1 -> $1, 2 -> $2
func (p *PostgresDialect) PlaceholderIndexed(index int) string {
	return fmt.Sprintf("$%d", index)
}

// Bool converts a boolean value to its dialect-specific SQL representation.
//
// Examples:
//
//	true -> TRUE, false -> FALSE
func (p *PostgresDialect) Bool(b bool) string {
	if b {
		return "TRUE"
	}
	return "FALSE"
}
