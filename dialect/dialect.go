// Package dialect defines the SqlDialect interface and common implementations
// that handle database-specific SQL differences such as quoting and placeholders.
package dialect

// SqlDialect abstracts database-specific SQL differences such as identifier quoting,
// bind parameter placeholders, and boolean literal syntax.
type SqlDialect interface {
	// Quote wraps a single identifier part in the dialect-specific quote character.
	// It also escapes any existing quote characters within the identifier.
	//
	// Examples (MySQL):
	//   users      -> `users`
	//   my`table   -> `my``table`
	//
	// Examples (PostgreSQL):
	//   users      -> "users"
	//   my"table   -> "my""table"
	Quote(str string) string

	// QuoteIdentifier quotes a full identifier (possibly dot-separated) by
	// splitting it and quoting each part individually.
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
	QuoteIdentifier(str string) string

	// PlaceholderPositional returns the positional bind parameter string.
	// Used for P() (no index, no name).
	//
	// Examples (PostgreSQL): ?
	// Examples (MySQL):      ?
	PlaceholderPositional() string

	// PlaceholderIndexed returns the indexed bind parameter string for the given index.
	// Index starts from 1.
	//
	// Examples (PostgreSQL): 1 -> $1, 2 -> $2
	// Examples (MySQL):      1 -> ?,  2 -> ?
	PlaceholderIndexed(index int) string

	// Bool converts a boolean value to its dialect-specific SQL representation.
	//
	// Examples (PostgreSQL): true -> TRUE,  false -> FALSE
	// Examples (SQLite):     true -> 1,     false -> 0
	Bool(b bool) string
}
