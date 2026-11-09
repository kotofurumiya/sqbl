package builder

import (
	"bytes"
	"slices"

	"github.com/kotofurumiya/sqbl/dialect"
	"github.com/kotofurumiya/sqbl/internal/sqlbuf"
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

// ToSQL renders the DELETE statement with a trailing semicolon.
func (b DeleteBuilder) ToSQL() string {
	d := requireDialect(b.dialect)
	buf := sqlbuf.GetStringBuffer()
	b.renderSQL(buf, d)
	buf.WriteByte(';')
	s := buf.String()
	sqlbuf.PutStringBuffer(buf)
	return s
}

// ToSQLWithDialect renders the DELETE statement using the given dialect, without a trailing semicolon.
func (b DeleteBuilder) ToSQLWithDialect(d dialect.Dialect) string {
	requireDialect(d)
	buf := sqlbuf.GetStringBuffer()
	b.renderSQL(buf, d)
	s := buf.String()
	sqlbuf.PutStringBuffer(buf)
	return s
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
