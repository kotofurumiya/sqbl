package builder

import (
	"bytes"

	"github.com/kotofurumiya/sqbl/dialect"
)

// DropDatabaseBuilder constructs a DROP DATABASE statement using a fluent method chain.
// Use a dialect-specific constructor (e.g. sqblpg.DropDatabase) to create an instance.
//
//	sqblpg.DropDatabase("app_test").IfExists()
//	sqblpg.DropDatabase("app_test").IfExists().Force()
type DropDatabaseBuilder struct {
	dialect  dialect.Dialect
	name     string
	ifExists bool
	force    bool
}

var _ Statement = DropDatabaseBuilder{}

// ToSQL renders the statement with the builder's dialect and a trailing semicolon.
func (b DropDatabaseBuilder) ToSQL() (string, error) { return render(b, b.dialect, true) }

// ToSQLWithDialect renders the statement with d and no trailing semicolon.
func (b DropDatabaseBuilder) ToSQLWithDialect(d dialect.Dialect) (string, error) {
	return render(b, d, false)
}

// MustSQL is like ToSQL but panics on error.
func (b DropDatabaseBuilder) MustSQL() string { return must(b.ToSQL()) }

// MustSQLWithDialect is like ToSQLWithDialect but panics on error.
func (b DropDatabaseBuilder) MustSQLWithDialect(d dialect.Dialect) string {
	return must(b.ToSQLWithDialect(d))
}

func (b DropDatabaseBuilder) renderSQL(buf *bytes.Buffer, d dialect.Dialect) {
	buf.WriteString("DROP DATABASE")
	if b.ifExists {
		buf.WriteString(" IF EXISTS")
	}
	if b.name != "" {
		buf.WriteByte(' ')
		d.Quote(buf, b.name)
	}
	if b.force {
		buf.WriteString(" WITH (FORCE)")
	}
}

// Dialect sets the SQL dialect used when rendering the query.
func (b DropDatabaseBuilder) Dialect(d dialect.Dialect) DropDatabaseBuilder {
	b.dialect = d
	return b
}

// Name sets the database name. It is quoted as a single identifier.
func (b DropDatabaseBuilder) Name(name string) DropDatabaseBuilder {
	b.name = name
	return b
}

// IfExists adds IF EXISTS to the statement, preventing an error if the database does not exist.
//
//	sqblpg.DropDatabase("app_test").IfExists()
//	// DROP DATABASE IF EXISTS "app_test"
func (b DropDatabaseBuilder) IfExists() DropDatabaseBuilder {
	b.ifExists = true
	return b
}

// Force adds WITH (FORCE), which terminates other connections to the database before dropping it.
// Supported by PostgreSQL 13 and later.
//
//	sqblpg.DropDatabase("app_test").Force()
//	// DROP DATABASE "app_test" WITH (FORCE)
func (b DropDatabaseBuilder) Force() DropDatabaseBuilder {
	b.force = true
	return b
}
