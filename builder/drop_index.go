package builder

import (
	"bytes"

	"github.com/kotofurumiya/sqbl/dialect"
	"github.com/kotofurumiya/sqbl/internal/sqlbuf"
)

// SqlDropIndexBuilder constructs a DROP INDEX statement using a fluent method chain.
// Use a dialect-specific constructor (e.g. sqblpg.DropIndex) to create an instance.
//
//	sqblpg.DropIndex("idx_users_email").IfExists()
//	sqblmysql.DropIndex("idx_users_email").On("users")
type SqlDropIndexBuilder struct {
	dialect  dialect.SqlDialect
	name     string
	ifExists bool
	on       string // required for MySQL: DROP INDEX name ON table
}

var _ SqlBuilder = SqlDropIndexBuilder{}

// ToSql renders the DROP INDEX statement with a trailing semicolon.
func (b SqlDropIndexBuilder) ToSql() string {
	buf := sqlbuf.GetStringBuffer()
	b.renderSQL(buf, b.dialect)
	buf.WriteByte(';')
	s := buf.String()
	sqlbuf.PutStringBuffer(buf)
	return s
}

// ToSqlWithDialect renders the DROP INDEX statement using the given dialect, without a trailing semicolon.
func (b SqlDropIndexBuilder) ToSqlWithDialect(d dialect.SqlDialect) string {
	buf := sqlbuf.GetStringBuffer()
	b.renderSQL(buf, d)
	s := buf.String()
	sqlbuf.PutStringBuffer(buf)
	return s
}

func (b SqlDropIndexBuilder) renderSQL(buf *bytes.Buffer, d dialect.SqlDialect) {
	buf.WriteString("DROP INDEX")
	if b.ifExists {
		buf.WriteString(" IF EXISTS")
	}
	if b.name != "" {
		buf.WriteByte(' ')
		d.QuoteIdentifier(buf, b.name)
	}
	if b.on != "" {
		buf.WriteString(" ON ")
		d.QuoteIdentifier(buf, b.on)
	}
}

// Dialect sets the SQL dialect used when rendering the query.
func (b SqlDropIndexBuilder) Dialect(d dialect.SqlDialect) SqlDropIndexBuilder {
	b.dialect = d
	return b
}

// Name sets the index name.
func (b SqlDropIndexBuilder) Name(name string) SqlDropIndexBuilder {
	b.name = name
	return b
}

// IfExists adds IF EXISTS to the statement.
// Prevents an error if the index does not exist.
// Supported by PostgreSQL, SQLite, and MariaDB; MySQL does not support it.
//
//	sqblpg.DropIndex("idx_users_email").IfExists()
//	// DROP INDEX IF EXISTS "idx_users_email"
func (b SqlDropIndexBuilder) IfExists() SqlDropIndexBuilder {
	b.ifExists = true
	return b
}

// On adds ON table to the statement.
// Required by MySQL and MariaDB; PostgreSQL and SQLite reject it.
//
//	sqblmysql.DropIndex("idx_users_email").On("users")
//	// DROP INDEX `idx_users_email` ON `users`
func (b SqlDropIndexBuilder) On(table string) SqlDropIndexBuilder {
	b.on = table
	return b
}
