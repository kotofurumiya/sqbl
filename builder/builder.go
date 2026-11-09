package builder

import (
	"bytes"

	"github.com/kotofurumiya/sqbl/dialect"
)

// Statement is implemented by all SQL statement builders.
// It provides a common interface for rendering SQL statements.
type Statement interface {
	// ToSQL renders the statement with a trailing semicolon using the builder's dialect.
	// It panics if the builder has no dialect.
	ToSQL() string
	// ToSQLWithDialect renders the statement using the given dialect, without a trailing semicolon.
	// It panics if d is nil.
	ToSQLWithDialect(d dialect.Dialect) string
}

var _ Statement = SelectBuilder{}

// requireDialect returns d, or panics with a hint when no dialect is available.
// A missing dialect is a programming error that shows up on the first render,
// so it is reported by panic rather than by an error return on every call.
func requireDialect(d dialect.Dialect) dialect.Dialect {
	if d == nil {
		panic("sqbl: no dialect to render with; use ToSQLWithDialect, " +
			"or start from a dialect package such as sqblpg")
	}
	return d
}

// quoteIdentifiers writes each name quoted by the dialect into buf, separated by ", ".
// Used by table-constraint closures in create_table.go.
//
//	quoteIdentifiers(buf, d, []string{"id", "name"})  →  "id", "name"
//	quoteIdentifiers(buf, d, nil)                     →  (nothing written)
func quoteIdentifiers(buf *bytes.Buffer, d dialect.Dialect, names []string) {
	for i, name := range names {
		if i > 0 {
			buf.WriteString(", ")
		}
		d.QuoteIdentifier(buf, name)
	}
}
