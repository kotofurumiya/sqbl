package builder

import (
	"strings"

	"github.com/kotofurumiya/sqbl/dialect"
	"github.com/kotofurumiya/sqbl/syntax"
)

// SqlDeleteBuilder constructs a SQL DELETE statement using a fluent method chain.
// Use a dialect-specific constructor (e.g. sqblpg.DeleteFrom) to create an instance.
//
//	q := sqblpg.DeleteFrom("users").
//	    Where(sqblpg.Eq("id", sqbl.P(1)))
//	sql := q.ToSql()
type SqlDeleteBuilder struct {
	dialect   dialect.SqlDialect
	from      string
	where     syntax.SqlFragment
	returning []string
}

var _ SqlBuilder = (*SqlDeleteBuilder)(nil)

// ToSql renders the DELETE statement with a trailing semicolon.
func (b *SqlDeleteBuilder) ToSql() string {
	return b.renderSQL(b.dialect) + ";"
}

// ToSqlWithDialect renders the DELETE statement using the given dialect, without a trailing semicolon.
func (b *SqlDeleteBuilder) ToSqlWithDialect(d dialect.SqlDialect) string {
	return b.renderSQL(d)
}

func (b *SqlDeleteBuilder) renderSQL(d dialect.SqlDialect) string {
	var parts []string

	// DELETE FROM table
	from := "DELETE FROM"
	if b.from != "" {
		from += " " + d.QuoteIdentifier(b.from)
	}
	parts = append(parts, from)

	// WHERE
	if b.where != nil {
		parts = append(parts, "WHERE "+b.where.ToSqlWithDialect(d))
	}

	// RETURNING
	if ret := quoteIdentifiers(d, b.returning); ret != "" {
		parts = append(parts, "RETURNING "+ret)
	}

	return strings.Join(parts, " ")
}

// Dialect sets the SQL dialect used when rendering the query.
func (b *SqlDeleteBuilder) Dialect(d dialect.SqlDialect) *SqlDeleteBuilder {
	b2 := *b
	b2.dialect = d
	return &b2
}

// From sets the target table for the DELETE statement.
//
//	b.From("users")
func (b *SqlDeleteBuilder) From(table string) *SqlDeleteBuilder {
	b2 := *b
	b2.from = table
	return &b2
}

// Where sets the WHERE condition.
//
//	b.Where(sqblpg.Eq("id", sqbl.P(1)))
func (b *SqlDeleteBuilder) Where(expr syntax.SqlFragment) *SqlDeleteBuilder {
	b2 := *b
	b2.where = expr
	return &b2
}

// Returning adds a RETURNING clause to the DELETE statement.
// The specified columns are returned for each deleted row.
//
//	sqblpg.DeleteFrom("sessions").Where(...).Returning("id", "user_id")
func (b *SqlDeleteBuilder) Returning(cols ...string) *SqlDeleteBuilder {
	b2 := *b
	b2.returning = cols
	return &b2
}
