package builder

import (
	"bytes"

	"github.com/kotofurumiya/sqbl/dialect"
	"github.com/kotofurumiya/sqbl/internal/sqlbuf"
)

// DropIndexBuilder constructs a DROP INDEX statement using a fluent method chain.
// Use a dialect-specific constructor (e.g. sqblpg.DropIndex) to create an instance.
//
//	sqblpg.DropIndex("idx_users_email").IfExists()
//	sqblmysql.DropIndex("idx_users_email").On("users")
type DropIndexBuilder struct {
	dialect  dialect.Dialect
	name     string
	ifExists bool
	on       string // required for MySQL: DROP INDEX name ON table
}

var _ Statement = DropIndexBuilder{}

// ToSQL renders the DROP INDEX statement with a trailing semicolon.
func (b DropIndexBuilder) ToSQL() string {
	d := requireDialect(b.dialect)
	buf := sqlbuf.GetStringBuffer()
	b.renderSQL(buf, d)
	buf.WriteByte(';')
	s := buf.String()
	sqlbuf.PutStringBuffer(buf)
	return s
}

// ToSQLWithDialect renders the DROP INDEX statement using the given dialect, without a trailing semicolon.
func (b DropIndexBuilder) ToSQLWithDialect(d dialect.Dialect) string {
	requireDialect(d)
	buf := sqlbuf.GetStringBuffer()
	b.renderSQL(buf, d)
	s := buf.String()
	sqlbuf.PutStringBuffer(buf)
	return s
}

func (b DropIndexBuilder) renderSQL(buf *bytes.Buffer, d dialect.Dialect) {
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
func (b DropIndexBuilder) Dialect(d dialect.Dialect) DropIndexBuilder {
	b.dialect = d
	return b
}

// Name sets the index name.
func (b DropIndexBuilder) Name(name string) DropIndexBuilder {
	b.name = name
	return b
}

// IfExists adds IF EXISTS to the statement.
// Prevents an error if the index does not exist.
// Supported by PostgreSQL, SQLite, and MariaDB; MySQL does not support it.
//
//	sqblpg.DropIndex("idx_users_email").IfExists()
//	// DROP INDEX IF EXISTS "idx_users_email"
func (b DropIndexBuilder) IfExists() DropIndexBuilder {
	b.ifExists = true
	return b
}

// On adds ON table to the statement.
// Required by MySQL and MariaDB; PostgreSQL and SQLite reject it.
//
//	sqblmysql.DropIndex("idx_users_email").On("users")
//	// DROP INDEX `idx_users_email` ON `users`
func (b DropIndexBuilder) On(table string) DropIndexBuilder {
	b.on = table
	return b
}
