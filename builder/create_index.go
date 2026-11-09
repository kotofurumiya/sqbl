package builder

import (
	"bytes"
	"slices"

	"github.com/kotofurumiya/sqbl/dialect"
	"github.com/kotofurumiya/sqbl/syntax"
)

// CreateIndexBuilder builds a CREATE INDEX statement.
// Use a dialect-specific constructor (e.g. sqblpg.CreateIndex) to create an instance.
//
//	sqblpg.CreateIndex("idx_users_email").On("users").Columns("email").Unique()
//	// CREATE UNIQUE INDEX "idx_users_email" ON "users" ("email")
type CreateIndexBuilder struct {
	dialect     dialect.Dialect
	unique      bool
	ifNotExists bool
	name        string
	table       string
	columns     []any                // key columns or expressions
	using       string               // index method, e.g. "btree", "hash", "gist" — verbatim, unquoted
	where       []syntax.SQLFragment // partial-index predicate, ANDed
}

var _ Statement = CreateIndexBuilder{}

// ToSQL renders the statement with the builder's dialect and a trailing semicolon.
func (b CreateIndexBuilder) ToSQL() (string, error) { return render(b, b.dialect, true) }

// ToSQLWithDialect renders the statement with d and no trailing semicolon.
func (b CreateIndexBuilder) ToSQLWithDialect(d dialect.Dialect) (string, error) {
	return render(b, d, false)
}

// MustSQL is like ToSQL but panics on error.
func (b CreateIndexBuilder) MustSQL() string { return must(b.ToSQL()) }

// MustSQLWithDialect is like ToSQLWithDialect but panics on error.
func (b CreateIndexBuilder) MustSQLWithDialect(d dialect.Dialect) string {
	return must(b.ToSQLWithDialect(d))
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
	writeValues(buf, d, b.columns)
	buf.WriteByte(')')

	// WHERE predicate for partial indexes — omitted when not set.
	writeConditions(buf, d, " WHERE ", b.where)
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

// Columns sets the index key columns, replacing any earlier list.
// Each item follows the rules of syntax.AppendValue: a plain string is a quoted
// column, and an expression or a direction can be given with Raw or Desc.
//
//	.Columns("last", "first")                   →  ("last", "first")
//	.Columns(sqblpg.Desc("created_at"))         →  ("created_at" DESC)
//	.Columns(sqblpg.Raw(`(lower("email"))`))   →  ((lower("email")))
func (b CreateIndexBuilder) Columns(cols ...any) CreateIndexBuilder {
	b.columns = slices.Clone(cols)
	return b
}

// Using sets the index access method (e.g. "btree", "hash", "gist", "gin"), replacing any earlier one.
// The value is written verbatim after the table name, which is PostgreSQL's placement;
// MySQL expects USING elsewhere. Never build it from untrusted input.
//
//	.Using("hash")  →  ON "users" USING hash ("email")
func (b CreateIndexBuilder) Using(method string) CreateIndexBuilder {
	b.using = method
	return b
}

// Where adds a predicate that makes the index partial. Supported by PostgreSQL and SQLite.
// Calling it again ANDs the new predicate with the earlier ones; nil is ignored.
//
//	.Where(sqblpg.IsNull("deleted_at"))  →  WHERE "deleted_at" IS NULL
func (b CreateIndexBuilder) Where(expr syntax.SQLFragment) CreateIndexBuilder {
	b.where = addCondition(b.where, expr)
	return b
}
