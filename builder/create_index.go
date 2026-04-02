package builder

import (
	"strings"

	"github.com/kotofurumiya/sqbl/dialect"
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

var _ SqlBuilder = (*SqlCreateIndexBuilder)(nil)

// renderSQL assembles the CREATE INDEX statement for the given dialect.
//
// The output follows this structure:
//
//	CREATE [UNIQUE] INDEX [IF NOT EXISTS] "name" ON "table" [USING method] ("col1", "col2") [WHERE expr]
//
// All column names are quoted via the dialect. The USING method and WHERE
// predicate are passed through verbatim, so callers are responsible for
// their correctness.
func (b *SqlCreateIndexBuilder) renderSQL(d dialect.SqlDialect) string {
	// Header: "CREATE INDEX" or "CREATE UNIQUE INDEX", with optional IF NOT EXISTS.
	var sb strings.Builder
	sb.WriteString("CREATE ")
	if b.unique {
		sb.WriteString("UNIQUE ")
	}
	sb.WriteString("INDEX ")
	if b.ifNotExists {
		sb.WriteString("IF NOT EXISTS ")
	}
	sb.WriteString(d.QuoteIdentifier(b.name))

	// ON "table"
	sb.WriteString(" ON ")
	sb.WriteString(d.QuoteIdentifier(b.table))

	// USING method — omitted when not set.
	if b.using != "" {
		sb.WriteString(" USING ")
		sb.WriteString(b.using)
	}

	// Column list: ("col1", "col2").
	// Each column name is quoted by the dialect.
	sb.WriteString(" (")
	sb.WriteString(quoteIdentifiers(d, b.columns))
	sb.WriteString(")")

	// WHERE predicate for partial indexes — omitted when not set.
	if b.where != nil {
		sb.WriteString(" WHERE ")
		sb.WriteString(b.where.ToSqlWithDialect(d))
	}

	return sb.String()
}

// ToSql renders the CREATE INDEX statement with a trailing semicolon.
func (b *SqlCreateIndexBuilder) ToSql() string {
	return b.renderSQL(b.dialect) + ";"
}

// ToSqlWithDialect renders the CREATE INDEX statement without a trailing semicolon.
func (b *SqlCreateIndexBuilder) ToSqlWithDialect(d dialect.SqlDialect) string {
	return b.renderSQL(d)
}

// Dialect sets the SQL dialect.
func (b *SqlCreateIndexBuilder) Dialect(d dialect.SqlDialect) *SqlCreateIndexBuilder {
	b2 := *b
	b2.dialect = d
	return &b2
}

// Name sets the index name.
func (b *SqlCreateIndexBuilder) Name(name string) *SqlCreateIndexBuilder {
	b2 := *b
	b2.name = name
	return &b2
}

// Unique marks the index as UNIQUE.
func (b *SqlCreateIndexBuilder) Unique() *SqlCreateIndexBuilder {
	b2 := *b
	b2.unique = true
	return &b2
}

// IfNotExists adds IF NOT EXISTS to the CREATE INDEX statement.
func (b *SqlCreateIndexBuilder) IfNotExists() *SqlCreateIndexBuilder {
	b2 := *b
	b2.ifNotExists = true
	return &b2
}

// On sets the target table for the index.
func (b *SqlCreateIndexBuilder) On(table string) *SqlCreateIndexBuilder {
	b2 := *b
	b2.table = table
	return &b2
}

// Columns appends one or more column names to the index column list.
// Each name is quoted as an identifier by the dialect at render time.
//
//	.Columns("email")          →  ("email")
//	.Columns("last", "first")  →  ("last", "first")
func (b *SqlCreateIndexBuilder) Columns(cols ...string) *SqlCreateIndexBuilder {
	b2 := *b
	b2.columns = append(append([]string(nil), b.columns...), cols...)
	return &b2
}

// Using sets the index access method (e.g. "btree", "hash", "gist", "gin").
// The value is written verbatim after USING without quoting.
//
//	.Using("hash")  →  USING hash
func (b *SqlCreateIndexBuilder) Using(method string) *SqlCreateIndexBuilder {
	b2 := *b
	b2.using = method
	return &b2
}

// Where sets a partial-index predicate.
// expr is a dialect-aware Expression, rendered the same way as a WHERE condition.
//
//	.Where(syntax.Eq("active", true))  →  WHERE "active" = TRUE
//	.Where(syntax.IsNull("deleted_at"))  →  WHERE "deleted_at" IS NULL
func (b *SqlCreateIndexBuilder) Where(expr syntax.SqlFragment) *SqlCreateIndexBuilder {
	b2 := *b
	b2.where = expr
	return &b2
}
