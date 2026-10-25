package builder

import (
	"bytes"

	"github.com/kotofurumiya/sqbl/dialect"
	"github.com/kotofurumiya/sqbl/internal/sqlbuf"
)

// SqlCreateDatabaseBuilder constructs a CREATE DATABASE statement using a fluent method chain.
// Use a dialect-specific constructor (e.g. sqblpg.CreateDatabase) to create an instance.
//
//	sqblpg.CreateDatabase("app_test")
//	sqblmysql.CreateDatabase("app_test").IfNotExists()
type SqlCreateDatabaseBuilder struct {
	dialect     dialect.SqlDialect
	name        string
	ifNotExists bool
}

var _ SqlBuilder = SqlCreateDatabaseBuilder{}

// ToSql renders the CREATE DATABASE statement with a trailing semicolon.
func (b SqlCreateDatabaseBuilder) ToSql() string {
	d := requireDialect(b.dialect)
	buf := sqlbuf.GetStringBuffer()
	b.renderSQL(buf, d)
	buf.WriteByte(';')
	s := buf.String()
	sqlbuf.PutStringBuffer(buf)
	return s
}

// ToSqlWithDialect renders the CREATE DATABASE statement using the given dialect, without a trailing semicolon.
func (b SqlCreateDatabaseBuilder) ToSqlWithDialect(d dialect.SqlDialect) string {
	requireDialect(d)
	buf := sqlbuf.GetStringBuffer()
	b.renderSQL(buf, d)
	s := buf.String()
	sqlbuf.PutStringBuffer(buf)
	return s
}

func (b SqlCreateDatabaseBuilder) renderSQL(buf *bytes.Buffer, d dialect.SqlDialect) {
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
func (b SqlCreateDatabaseBuilder) Dialect(d dialect.SqlDialect) SqlCreateDatabaseBuilder {
	b.dialect = d
	return b
}

// Name sets the database name. It is quoted as a single identifier.
func (b SqlCreateDatabaseBuilder) Name(name string) SqlCreateDatabaseBuilder {
	b.name = name
	return b
}

// IfNotExists adds IF NOT EXISTS to the statement. Supported by MySQL; PostgreSQL rejects it.
//
//	sqblmysql.CreateDatabase("app_test").IfNotExists()
//	// CREATE DATABASE IF NOT EXISTS `app_test`
func (b SqlCreateDatabaseBuilder) IfNotExists() SqlCreateDatabaseBuilder {
	b.ifNotExists = true
	return b
}
