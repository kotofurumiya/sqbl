package builder

import (
	"strings"

	"github.com/kotofurumiya/sqbl/dialect"
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

var _ SqlBuilder = (*SqlDropTableBuilder)(nil)

// ToSql renders the DROP TABLE statement with a trailing semicolon.
func (b *SqlDropTableBuilder) ToSql() string {
	return b.renderSQL(b.dialect) + ";"
}

// ToSqlWithDialect renders the DROP TABLE statement using the given dialect, without a trailing semicolon.
func (b *SqlDropTableBuilder) ToSqlWithDialect(d dialect.SqlDialect) string {
	return b.renderSQL(d)
}

func (b *SqlDropTableBuilder) renderSQL(d dialect.SqlDialect) string {
	var parts []string

	drop := "DROP TABLE"
	if b.ifExists {
		drop += " IF EXISTS"
	}
	if b.table != "" {
		drop += " " + d.QuoteIdentifier(b.table)
	}
	parts = append(parts, drop)

	if b.cascade {
		if _, isPg := d.(*dialect.PostgresDialect); isPg {
			parts = append(parts, "CASCADE")
		}
	}

	return strings.Join(parts, " ")
}

// Dialect sets the SQL dialect used when rendering the query.
func (b *SqlDropTableBuilder) Dialect(d dialect.SqlDialect) *SqlDropTableBuilder {
	b2 := *b
	b2.dialect = d
	return &b2
}

// Table sets the target table name.
func (b *SqlDropTableBuilder) Table(table string) *SqlDropTableBuilder {
	b2 := *b
	b2.table = table
	return &b2
}

// IfExists adds IF EXISTS to the statement, preventing an error if the table does not exist.
//
//	sqblpg.DropTable("users").IfExists()
//	// DROP TABLE IF EXISTS "users"
func (b *SqlDropTableBuilder) IfExists() *SqlDropTableBuilder {
	b2 := *b
	b2.ifExists = true
	return &b2
}

// Cascade adds CASCADE to the statement (PostgreSQL only).
// Automatically drops objects that depend on the table.
//
//	sqblpg.DropTable("users").Cascade()
//	// DROP TABLE "users" CASCADE
func (b *SqlDropTableBuilder) Cascade() *SqlDropTableBuilder {
	b2 := *b
	b2.cascade = true
	return &b2
}
