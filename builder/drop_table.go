package builder

import (
	"bytes"

	"github.com/kotofurumiya/sqbl/dialect"
)

// DropTableBuilder constructs a DROP TABLE statement using a fluent method chain.
// Use a dialect-specific constructor (e.g. sqblpg.DropTable) to create an instance.
//
//	sqblpg.DropTable("users").IfExists()
//	sqblpg.DropTable("users").IfExists().Cascade()
type DropTableBuilder struct {
	dialect  dialect.Dialect
	table    string
	ifExists bool
	cascade  bool // PostgreSQL only
}

var _ Statement = DropTableBuilder{}

// ToSQL renders the statement with the builder's dialect and a trailing semicolon.
func (b DropTableBuilder) ToSQL() (string, error) { return render(b, b.dialect, true) }

// ToSQLWithDialect renders the statement with d and no trailing semicolon.
func (b DropTableBuilder) ToSQLWithDialect(d dialect.Dialect) (string, error) {
	return render(b, d, false)
}

// MustSQL is like ToSQL but panics on error.
func (b DropTableBuilder) MustSQL() string { return must(b.ToSQL()) }

// MustSQLWithDialect is like ToSQLWithDialect but panics on error.
func (b DropTableBuilder) MustSQLWithDialect(d dialect.Dialect) string {
	return must(b.ToSQLWithDialect(d))
}

func (b DropTableBuilder) renderSQL(buf *bytes.Buffer, d dialect.Dialect) {
	buf.WriteString("DROP TABLE")
	if b.ifExists {
		buf.WriteString(" IF EXISTS")
	}
	if b.table != "" {
		buf.WriteByte(' ')
		d.QuoteIdentifier(buf, b.table)
	}
	if b.cascade {
		buf.WriteString(" CASCADE")
	}
}

// Dialect sets the SQL dialect used when rendering the query.
func (b DropTableBuilder) Dialect(d dialect.Dialect) DropTableBuilder {
	b.dialect = d
	return b
}

// Table sets the target table name.
func (b DropTableBuilder) Table(table string) DropTableBuilder {
	b.table = table
	return b
}

// IfExists adds IF EXISTS to the statement, preventing an error if the table does not exist.
//
//	sqblpg.DropTable("users").IfExists()
//	// DROP TABLE IF EXISTS "users"
func (b DropTableBuilder) IfExists() DropTableBuilder {
	b.ifExists = true
	return b
}

// Cascade adds CASCADE to the statement. Supported by PostgreSQL.
// Automatically drops objects that depend on the table.
//
//	sqblpg.DropTable("users").Cascade()
//	// DROP TABLE "users" CASCADE
func (b DropTableBuilder) Cascade() DropTableBuilder {
	b.cascade = true
	return b
}
