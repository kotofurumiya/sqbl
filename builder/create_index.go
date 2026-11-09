package builder

import (
	"bytes"

	"github.com/kotofurumiya/sqbl/dialect"
	"github.com/kotofurumiya/sqbl/internal/sqlbuf"
	"github.com/kotofurumiya/sqbl/syntax"
)

// CreateIndexBuilder builds a CREATE INDEX statement.
type CreateIndexBuilder struct {
	dialect     dialect.Dialect
	unique      bool
	ifNotExists bool
	name        string
	table       string
	columns     []string           // column names, quoted as identifiers
	using       string             // index method, e.g. "btree", "hash", "gist" — verbatim, unquoted
	where       syntax.SQLFragment // partial-index predicate
}

var _ Statement = CreateIndexBuilder{}

// renderSQL assembles the CREATE INDEX statement for the given dialect.
//
// The output follows this structure:
//
//	CREATE [UNIQUE] INDEX [IF NOT EXISTS] "name" ON "table" [USING method] ("col1", "col2") [WHERE expr]
//
// All column names are quoted via the dialect. The USING method and WHERE
// predicate are passed through verbatim, so callers are responsible for
// their correctness.

// ToSQL renders the CREATE INDEX statement with a trailing semicolon.
func (b CreateIndexBuilder) ToSQL() string {
	d := requireDialect(b.dialect)
	buf := sqlbuf.GetStringBuffer()
	b.renderSQL(buf, d)
	buf.WriteByte(';')
	s := buf.String()
	sqlbuf.PutStringBuffer(buf)
	return s
}

// ToSQLWithDialect renders the CREATE INDEX statement without a trailing semicolon.
func (b CreateIndexBuilder) ToSQLWithDialect(d dialect.Dialect) string {
	requireDialect(d)
	buf := sqlbuf.GetStringBuffer()
	b.renderSQL(buf, d)
	s := buf.String()
	sqlbuf.PutStringBuffer(buf)
	return s
}

func (b CreateIndexBuilder) renderSQL(buf *bytes.Buffer, d dialect.Dialect) {
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
func (b CreateIndexBuilder) Dialect(d dialect.Dialect) CreateIndexBuilder {
	b.dialect = d
	return b
}

// Name sets the index name.
func (b CreateIndexBuilder) Name(name string) CreateIndexBuilder {
	b.name = name
	return b
}

// Unique marks the index as UNIQUE.
func (b CreateIndexBuilder) Unique() CreateIndexBuilder {
	b.unique = true
	return b
}

// IfNotExists adds IF NOT EXISTS to the CREATE INDEX statement.
func (b CreateIndexBuilder) IfNotExists() CreateIndexBuilder {
	b.ifNotExists = true
	return b
}

// On sets the target table for the index.
func (b CreateIndexBuilder) On(table string) CreateIndexBuilder {
	b.table = table
	return b
}

// Columns appends one or more column names to the index column list.
// Each name is quoted as an identifier by the dialect at render time.
//
//	.Columns("email")          →  ("email")
//	.Columns("last", "first")  →  ("last", "first")
func (b CreateIndexBuilder) Columns(cols ...string) CreateIndexBuilder {
	b.columns = append(append([]string(nil), b.columns...), cols...)
	return b
}

// Using sets the index access method (e.g. "btree", "hash", "gist", "gin").
// The value is written verbatim after USING without quoting.
//
//	.Using("hash")  →  USING hash
func (b CreateIndexBuilder) Using(method string) CreateIndexBuilder {
	b.using = method
	return b
}

// Where sets a partial-index predicate.
// expr is a dialect-aware Expression, rendered the same way as a WHERE condition.
//
//	.Where(syntax.Eq("active", true))  →  WHERE "active" = TRUE
//	.Where(syntax.IsNull("deleted_at"))  →  WHERE "deleted_at" IS NULL
func (b CreateIndexBuilder) Where(expr syntax.SQLFragment) CreateIndexBuilder {
	b.where = expr
	return b
}
