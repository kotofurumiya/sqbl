package builder

import (
	"bytes"

	"github.com/kotofurumiya/sqbl/dialect"
	"github.com/kotofurumiya/sqbl/internal/sqlbuf"
	"github.com/kotofurumiya/sqbl/syntax"
)

// SqlCreateIndexBuilder builds a CREATE INDEX statement.
type SqlCreateIndexBuilder struct {
	dialect     dialect.SqlDialect
	unique      bool
	ifNotExists bool
	name        string
	table       string
	columns     []string           // column names, quoted as identifiers
	using       string             // index method, e.g. "btree", "hash", "gist" — verbatim, unquoted
	where       syntax.SqlFragment // partial-index predicate
}

var _ SqlBuilder = SqlCreateIndexBuilder{}

// renderSQL assembles the CREATE INDEX statement for the given dialect.
//
// The output follows this structure:
//
//	CREATE [UNIQUE] INDEX [IF NOT EXISTS] "name" ON "table" [USING method] ("col1", "col2") [WHERE expr]
//
// All column names are quoted via the dialect. The USING method and WHERE
// predicate are passed through verbatim, so callers are responsible for
// their correctness.
// ToSql renders the CREATE INDEX statement with a trailing semicolon.
func (b SqlCreateIndexBuilder) ToSql() string {
	buf := sqlbuf.GetStringBuffer()
	b.renderSQL(buf, b.dialect)
	buf.WriteByte(';')
	s := buf.String()
	sqlbuf.PutStringBuffer(buf)
	return s
}

// ToSqlWithDialect renders the CREATE INDEX statement without a trailing semicolon.
func (b SqlCreateIndexBuilder) ToSqlWithDialect(d dialect.SqlDialect) string {
	buf := sqlbuf.GetStringBuffer()
	b.renderSQL(buf, d)
	s := buf.String()
	sqlbuf.PutStringBuffer(buf)
	return s
}

func (b SqlCreateIndexBuilder) renderSQL(buf *bytes.Buffer, d dialect.SqlDialect) {
	// CREATE [UNIQUE] INDEX [IF NOT EXISTS] "name"
	buf.WriteString("CREATE ")
	if b.unique {
		buf.WriteString("UNIQUE ")
	}
	buf.WriteString("INDEX ")
	if b.ifNotExists {
		buf.WriteString("IF NOT EXISTS ")
	}
	d.QuoteIdentifier(buf, b.name)

	// ON "table"
	buf.WriteString(" ON ")
	d.QuoteIdentifier(buf, b.table)

	// USING method — omitted when not set.
	if b.using != "" {
		buf.WriteString(" USING ")
		buf.WriteString(b.using)
	}

	// Column list: ("col1", "col2") — each name quoted by the dialect.
	buf.WriteString(" (")
	for i, col := range b.columns {
		if i > 0 {
			buf.WriteString(", ")
		}
		d.QuoteIdentifier(buf, col)
	}
	buf.WriteByte(')')

	// WHERE predicate for partial indexes — omitted when not set.
	if b.where != nil {
		buf.WriteString(" WHERE ")
		b.where.AppendSQL(buf, d)
	}
}

// Dialect sets the SQL dialect.
func (b SqlCreateIndexBuilder) Dialect(d dialect.SqlDialect) SqlCreateIndexBuilder {
	b.dialect = d
	return b
}

// Name sets the index name.
func (b SqlCreateIndexBuilder) Name(name string) SqlCreateIndexBuilder {
	b.name = name
	return b
}

// Unique marks the index as UNIQUE.
func (b SqlCreateIndexBuilder) Unique() SqlCreateIndexBuilder {
	b.unique = true
	return b
}

// IfNotExists adds IF NOT EXISTS to the CREATE INDEX statement.
func (b SqlCreateIndexBuilder) IfNotExists() SqlCreateIndexBuilder {
	b.ifNotExists = true
	return b
}

// On sets the target table for the index.
func (b SqlCreateIndexBuilder) On(table string) SqlCreateIndexBuilder {
	b.table = table
	return b
}

// Columns appends one or more column names to the index column list.
// Each name is quoted as an identifier by the dialect at render time.
//
//	.Columns("email")          →  ("email")
//	.Columns("last", "first")  →  ("last", "first")
func (b SqlCreateIndexBuilder) Columns(cols ...string) SqlCreateIndexBuilder {
	b.columns = append(append([]string(nil), b.columns...), cols...)
	return b
}

// Using sets the index access method (e.g. "btree", "hash", "gist", "gin").
// The value is written verbatim after USING without quoting.
//
//	.Using("hash")  →  USING hash
func (b SqlCreateIndexBuilder) Using(method string) SqlCreateIndexBuilder {
	b.using = method
	return b
}

// Where sets a partial-index predicate.
// expr is a dialect-aware Expression, rendered the same way as a WHERE condition.
//
//	.Where(syntax.Eq("active", true))  →  WHERE "active" = TRUE
//	.Where(syntax.IsNull("deleted_at"))  →  WHERE "deleted_at" IS NULL
func (b SqlCreateIndexBuilder) Where(expr syntax.SqlFragment) SqlCreateIndexBuilder {
	b.where = expr
	return b
}
