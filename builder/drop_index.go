package builder

import (
	"github.com/kotofurumiya/sqbl/dialect"
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

var _ SqlBuilder = (*SqlDropIndexBuilder)(nil)

// ToSql renders the DROP INDEX statement with a trailing semicolon.
func (b *SqlDropIndexBuilder) ToSql() string {
	return b.renderSQL(b.dialect) + ";"
}

// ToSqlWithDialect renders the DROP INDEX statement using the given dialect, without a trailing semicolon.
func (b *SqlDropIndexBuilder) ToSqlWithDialect(d dialect.SqlDialect) string {
	return b.renderSQL(d)
}

func (b *SqlDropIndexBuilder) renderSQL(d dialect.SqlDialect) string {
	_, isMysql := d.(*dialect.MysqlDialect)

	drop := "DROP INDEX"
	if b.ifExists && !isMysql {
		drop += " IF EXISTS"
	}
	if b.name != "" {
		drop += " " + d.QuoteIdentifier(b.name)
	}

	// MySQL requires ON table clause
	if isMysql && b.on != "" {
		drop += " ON " + d.QuoteIdentifier(b.on)
	}

	return drop
}

// Dialect sets the SQL dialect used when rendering the query.
func (b *SqlDropIndexBuilder) Dialect(d dialect.SqlDialect) *SqlDropIndexBuilder {
	b2 := *b
	b2.dialect = d
	return &b2
}

// Name sets the index name.
func (b *SqlDropIndexBuilder) Name(name string) *SqlDropIndexBuilder {
	b2 := *b
	b2.name = name
	return &b2
}

// IfExists adds IF EXISTS to the statement (PostgreSQL and SQLite only).
// Prevents an error if the index does not exist.
//
//	sqblpg.DropIndex("idx_users_email").IfExists()
//	// DROP INDEX IF EXISTS "idx_users_email"
func (b *SqlDropIndexBuilder) IfExists() *SqlDropIndexBuilder {
	b2 := *b
	b2.ifExists = true
	return &b2
}

// On sets the table name for the DROP INDEX statement (MySQL only).
// MySQL requires specifying the table when dropping an index.
//
//	sqblmysql.DropIndex("idx_users_email").On("users")
//	// DROP INDEX `idx_users_email` ON `users`
func (b *SqlDropIndexBuilder) On(table string) *SqlDropIndexBuilder {
	b2 := *b
	b2.on = table
	return &b2
}
