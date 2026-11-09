package builder

import (
	"bytes"

	"github.com/kotofurumiya/sqbl/dialect"
	"github.com/kotofurumiya/sqbl/internal/sqlbuf"
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

// ToSQL renders the DROP TABLE statement with a trailing semicolon.
func (b DropTableBuilder) ToSQL() string {
	d := requireDialect(b.dialect)
	buf := sqlbuf.GetStringBuffer()
	b.renderSQL(buf, d)
	buf.WriteByte(';')
	s := buf.String()
	sqlbuf.PutStringBuffer(buf)
	return s
}

// ToSQLWithDialect renders the DROP TABLE statement using the given dialect, without a trailing semicolon.
func (b DropTableBuilder) ToSQLWithDialect(d dialect.Dialect) string {
	requireDialect(d)
	buf := sqlbuf.GetStringBuffer()
	b.renderSQL(buf, d)
	s := buf.String()
	sqlbuf.PutStringBuffer(buf)
	return s
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
