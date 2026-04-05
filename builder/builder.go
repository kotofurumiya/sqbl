package builder

import (
	"bytes"

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

var _ SqlBuilder = SqlSelectBuilder{}

// quoteIdentifiers writes each name quoted by the dialect into buf, separated by ", ".
// Used by table-constraint closures in create_table.go.
//
//	quoteIdentifiers(buf, d, []string{"id", "name"})  →  "id", "name"
//	quoteIdentifiers(buf, d, nil)                     →  (nothing written)
func quoteIdentifiers(buf *bytes.Buffer, d dialect.SqlDialect, names []string) {
	for i, name := range names {
		if i > 0 {
			buf.WriteString(", ")
		}
		d.QuoteIdentifier(buf, name)
	}
}
