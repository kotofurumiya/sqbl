package builder

import (
	"bytes"

	"github.com/kotofurumiya/sqbl/dialect"
	"github.com/kotofurumiya/sqbl/internal/sqlbuf"
)

// SqlDropTableBuilder constructs a DROP TABLE statement using a fluent method chain.
// Use a dialect-specific constructor (e.g. sqblpg.DropTable) to create an instance.
//
//	sqblpg.DropTable("users").IfExists()
//	sqblpg.DropTable("users").IfExists().Cascade()
type SqlDropTableBuilder struct {
	dialect  dialect.SqlDialect
	table    string
	ifExists bool
	cascade  bool // PostgreSQL only
}

var _ SqlBuilder = SqlDropTableBuilder{}

// ToSql renders the DROP TABLE statement with a trailing semicolon.
func (b SqlDropTableBuilder) ToSql() string {
	buf := sqlbuf.GetStringBuffer()
	b.renderSQL(buf, b.dialect)
	buf.WriteByte(';')
	s := buf.String()
	sqlbuf.PutStringBuffer(buf)
	return s
}

// ToSqlWithDialect renders the DROP TABLE statement using the given dialect, without a trailing semicolon.
func (b SqlDropTableBuilder) ToSqlWithDialect(d dialect.SqlDialect) string {
	buf := sqlbuf.GetStringBuffer()
	b.renderSQL(buf, d)
	s := buf.String()
	sqlbuf.PutStringBuffer(buf)
	return s
}

func (b SqlDropTableBuilder) renderSQL(buf *bytes.Buffer, d dialect.SqlDialect) {
	buf.WriteString("DROP TABLE")
	if b.ifExists {
		buf.WriteString(" IF EXISTS")
	}
	if b.table != "" {
		buf.WriteByte(' ')
		d.QuoteIdentifier(buf, b.table)
	}
	// CASCADE is PostgreSQL-only; ignored on other dialects.
	if b.cascade {
		if _, isPg := d.(*dialect.PostgresDialect); isPg {
			buf.WriteString(" CASCADE")
		}
	}
}

// Dialect sets the SQL dialect used when rendering the query.
func (b SqlDropTableBuilder) Dialect(d dialect.SqlDialect) SqlDropTableBuilder {
	b.dialect = d
	return b
}

// Table sets the target table name.
func (b SqlDropTableBuilder) Table(table string) SqlDropTableBuilder {
	b.table = table
	return b
}

// IfExists adds IF EXISTS to the statement, preventing an error if the table does not exist.
//
//	sqblpg.DropTable("users").IfExists()
//	// DROP TABLE IF EXISTS "users"
func (b SqlDropTableBuilder) IfExists() SqlDropTableBuilder {
	b.ifExists = true
	return b
}

// Cascade adds CASCADE to the statement (PostgreSQL only).
// Automatically drops objects that depend on the table.
//
//	sqblpg.DropTable("users").Cascade()
//	// DROP TABLE "users" CASCADE
func (b SqlDropTableBuilder) Cascade() SqlDropTableBuilder {
	b.cascade = true
	return b
}
