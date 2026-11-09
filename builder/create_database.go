package builder

import (
	"bytes"

	"github.com/kotofurumiya/sqbl/dialect"
)

// CreateDatabaseBuilder constructs a CREATE DATABASE statement using a fluent method chain.
// Use a dialect-specific constructor (e.g. sqblpg.CreateDatabase) to create an instance.
//
//	sqblpg.CreateDatabase("app_test")
//	sqblmysql.CreateDatabase("app_test").IfNotExists()
type CreateDatabaseBuilder struct {
	dialect     dialect.Dialect
	name        string
	ifNotExists bool
}

var _ Statement = CreateDatabaseBuilder{}

// ToSQL renders the statement with the builder's dialect and a trailing semicolon.
func (b CreateDatabaseBuilder) ToSQL() (string, error) { return render(b, b.dialect, true) }

// ToSQLWithDialect renders the statement with d and no trailing semicolon.
func (b CreateDatabaseBuilder) ToSQLWithDialect(d dialect.Dialect) (string, error) {
	return render(b, d, false)
}

// MustSQL is like ToSQL but panics on error.
func (b CreateDatabaseBuilder) MustSQL() string { return must(b.ToSQL()) }

// MustSQLWithDialect is like ToSQLWithDialect but panics on error.
func (b CreateDatabaseBuilder) MustSQLWithDialect(d dialect.Dialect) string {
	return must(b.ToSQLWithDialect(d))
}

func (b CreateDatabaseBuilder) renderSQL(buf *bytes.Buffer, d dialect.Dialect) {
	buf.WriteString("CREATE DATABASE")
	if b.ifNotExists {
		buf.WriteString(" IF NOT EXISTS")
	}
	if b.name != "" {
		buf.WriteByte(' ')
		d.Quote(buf, b.name)
	}
}

// Dialect sets the SQL dialect used when rendering the query.
func (b CreateDatabaseBuilder) Dialect(d dialect.Dialect) CreateDatabaseBuilder {
	b.dialect = d
	return b
}

// Name sets the database name. It is quoted as a single identifier.
func (b CreateDatabaseBuilder) Name(name string) CreateDatabaseBuilder {
	b.name = name
	return b
}

// IfNotExists adds IF NOT EXISTS to the statement. Supported by MySQL; PostgreSQL rejects it.
//
//	sqblmysql.CreateDatabase("app_test").IfNotExists()
//	// CREATE DATABASE IF NOT EXISTS `app_test`
func (b CreateDatabaseBuilder) IfNotExists() CreateDatabaseBuilder {
	b.ifNotExists = true
	return b
}
