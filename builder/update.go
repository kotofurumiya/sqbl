package builder

import (
	"strings"

	"github.com/kotofurumiya/sqbl/dialect"
	"github.com/kotofurumiya/sqbl/syntax"
)

// setClause represents a single SET col = val entry.
type setClause struct {
	Col string
	Val any
}

// SqlUpdateBuilder constructs a SQL UPDATE statement using a fluent method chain.
// Use a dialect-specific constructor (e.g. sqblpg.Update) to create an instance.
//
//	q := sqblpg.Update("users").
//	    Set("name", sqbl.P(1)).
//	    Set("email", sqbl.P(2)).
//	    Where(sqblpg.Eq("id", sqbl.P(3)))
//	sql := q.ToSql()
type SqlUpdateBuilder struct {
	dialect   dialect.SqlDialect
	table     string
	sets      []setClause
	where     syntax.SqlFragment
	returning []string
}

var _ SqlBuilder = (*SqlUpdateBuilder)(nil)

// ToSql renders the UPDATE statement with a trailing semicolon.
func (b *SqlUpdateBuilder) ToSql() string {
	return b.renderSQL(b.dialect) + ";"
}

// ToSqlWithDialect renders the UPDATE statement using the given dialect, without a trailing semicolon.
func (b *SqlUpdateBuilder) ToSqlWithDialect(d dialect.SqlDialect) string {
	return b.renderSQL(d)
}

func (b *SqlUpdateBuilder) renderSQL(d dialect.SqlDialect) string {
	var parts []string

	// UPDATE table
	update := "UPDATE"
	if b.table != "" {
		update += " " + d.QuoteIdentifier(b.table)
	}
	parts = append(parts, update)

	// SET col = val, ...
	if sets := mapJoin(b.sets, ", ", func(s setClause) string {
		return d.QuoteIdentifier(s.Col) + " = " + renderValue(d, s.Val)
	}); sets != "" {
		parts = append(parts, "SET "+sets)
	}

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
func (b *SqlUpdateBuilder) Dialect(d dialect.SqlDialect) *SqlUpdateBuilder {
	b2 := *b
	b2.dialect = d
	return &b2
}

// Table sets the target table for the UPDATE statement.
//
//	b.Table("users")
func (b *SqlUpdateBuilder) Table(table string) *SqlUpdateBuilder {
	b2 := *b
	b2.table = table
	return &b2
}

// Set adds a SET assignment to the UPDATE statement.
// Call Set multiple times to assign multiple columns.
//
//	b.Set("name", sqbl.P(1)).Set("email", sqbl.P(2))
func (b *SqlUpdateBuilder) Set(col string, val any) *SqlUpdateBuilder {
	b2 := *b
	b2.sets = append(append([]setClause(nil), b.sets...), setClause{Col: col, Val: val})
	return &b2
}

// Where sets the WHERE condition.
//
//	b.Where(sqblpg.Eq("id", sqbl.P(1)))
func (b *SqlUpdateBuilder) Where(expr syntax.SqlFragment) *SqlUpdateBuilder {
	b2 := *b
	b2.where = expr
	return &b2
}

// Returning adds a RETURNING clause to the UPDATE statement.
// The specified columns are returned for each updated row.
//
//	sqblpg.Update("users").Set("name", sqbl.P(1)).Where(...).Returning("id", "name")
func (b *SqlUpdateBuilder) Returning(cols ...string) *SqlUpdateBuilder {
	b2 := *b
	b2.returning = cols
	return &b2
}
