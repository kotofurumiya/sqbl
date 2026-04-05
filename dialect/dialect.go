// Package dialect defines the SqlDialect interface and common implementations
// that handle database-specific SQL differences such as quoting and placeholders.
package dialect

import "bytes"

// SqlDialect abstracts database-specific SQL differences such as identifier quoting,
// bind parameter placeholders, and boolean literal syntax.
type SqlDialect interface {
	// Quote writes a single identifier part wrapped in the dialect-specific quote
	// character directly into buf. Any existing quote characters within the identifier
	// are escaped.
	//
	// Examples (MySQL):
	//   users      -> `users`
	//   my`table   -> `my``table`
	//
	// Examples (PostgreSQL):
	//   users      -> "users"
	//   my"table   -> "my""table"
	Quote(buf *bytes.Buffer, s string)

	// QuoteIdentifier writes a full identifier (possibly dot-separated) into buf,
	// quoting each part individually.
	//
	// Examples (MySQL):
	//   users          -> `users`
	//   public.users   -> `public`.`users`
	//   db.sch.table   -> `db`.`sch`.`table`
	//
	// Examples (PostgreSQL):
	//   users          -> "users"
	//   public.users   -> "public"."users"
	//   db.sch.table   -> "db"."sch"."table"
	QuoteIdentifier(buf *bytes.Buffer, name string)

	// PlaceholderPositional writes the positional bind parameter into buf.
	// Used for P() (no index, no name).
	//
	// Examples (PostgreSQL): ?
	// Examples (MySQL):      ?
	PlaceholderPositional(buf *bytes.Buffer)

	// PlaceholderIndexed writes the indexed bind parameter into buf.
	// Index starts from 1.
	//
	// Examples (PostgreSQL): 1 -> $1, 2 -> $2
	// Examples (MySQL):      1 -> ?,  2 -> ?
	PlaceholderIndexed(buf *bytes.Buffer, index int)

	// Bool writes the dialect-specific SQL representation of a boolean value into buf.
	//
	// Examples (PostgreSQL): true -> TRUE,  false -> FALSE
	// Examples (SQLite):     true -> 1,     false -> 0
	Bool(buf *bytes.Buffer, b bool)
}
