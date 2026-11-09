package builder

import (
	"bytes"
	"slices"

	"github.com/kotofurumiya/sqbl/dialect"
	"github.com/kotofurumiya/sqbl/syntax"
)

// DeleteBuilder constructs a SQL DELETE statement using a fluent method chain.
// Use a dialect-specific constructor (e.g. sqblpg.DeleteFrom) to create an instance.
//
//	q := sqblpg.DeleteFrom("users").
//	    Where(sqblpg.Eq("id", sqbl.P(1)))
//	sql := q.ToSQL()
type DeleteBuilder struct {
	dialect   dialect.Dialect
	from      string
	where     syntax.SQLFragment
	returning []string
}

var _ Statement = DeleteBuilder{}

// ToSQL renders the statement with the builder's dialect and a trailing semicolon.
func (b DeleteBuilder) ToSQL() (string, error) { return render(b, b.dialect, true) }

// ToSQLWithDialect renders the statement with d and no trailing semicolon.
func (b DeleteBuilder) ToSQLWithDialect(d dialect.Dialect) (string, error) {
	return render(b, d, false)
}

// MustSQL is like ToSQL but panics on error.
func (b DeleteBuilder) MustSQL() string { return must(b.ToSQL()) }

// MustSQLWithDialect is like ToSQLWithDialect but panics on error.
func (b DeleteBuilder) MustSQLWithDialect(d dialect.Dialect) string {
	return must(b.ToSQLWithDialect(d))
}

func (b DeleteBuilder) renderSQL(buf *bytes.Buffer, d dialect.Dialect) {
	// DELETE FROM table
	buf.WriteString("DELETE FROM")
	if b.from != "" {
		buf.WriteByte(' ')
		d.QuoteIdentifier(buf, b.from)
	}

	// WHERE condition
	if b.where != nil {
		buf.WriteString(" WHERE ")
		b.where.AppendSQL(buf, d)
	}

	// RETURNING col1, col2, ...
	if len(b.returning) > 0 {
		buf.WriteString(" RETURNING ")
		for i, col := range b.returning {
			if i > 0 {
				buf.WriteString(", ")
			}
			d.QuoteIdentifier(buf, col)
		}
	}
}

// Dialect sets the SQL dialect used when rendering the query.
func (b DeleteBuilder) Dialect(d dialect.Dialect) DeleteBuilder {
	b.dialect = d
	return b
}

// From sets the target table for the DELETE statement.
//
//	b.From("users")
func (b DeleteBuilder) From(table string) DeleteBuilder {
	b.from = table
	return b
}

// Where sets the WHERE condition.
//
//	b.Where(sqblpg.Eq("id", sqbl.P(1)))
func (b DeleteBuilder) Where(expr syntax.SQLFragment) DeleteBuilder {
	b.where = expr
	return b
}

// Returning adds a RETURNING clause to the DELETE statement.
// The specified columns are returned for each deleted row.
//
//	sqblpg.DeleteFrom("sessions").Where(...).Returning("id", "user_id")
func (b DeleteBuilder) Returning(cols ...string) DeleteBuilder {
	b.returning = slices.Clone(cols)
	return b
}
