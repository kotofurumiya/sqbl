package builder

import (
	"bytes"

	"github.com/kotofurumiya/sqbl/dialect"
	"github.com/kotofurumiya/sqbl/internal/sqlbuf"
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

var _ SqlBuilder = SqlDeleteBuilder{}

// ToSql renders the DELETE statement with a trailing semicolon.
func (b SqlDeleteBuilder) ToSql() string {
	buf := sqlbuf.GetStringBuffer()
	b.renderSQL(buf, b.dialect)
	buf.WriteByte(';')
	s := buf.String()
	sqlbuf.PutStringBuffer(buf)
	return s
}

// ToSqlWithDialect renders the DELETE statement using the given dialect, without a trailing semicolon.
func (b SqlDeleteBuilder) ToSqlWithDialect(d dialect.SqlDialect) string {
	buf := sqlbuf.GetStringBuffer()
	b.renderSQL(buf, d)
	s := buf.String()
	sqlbuf.PutStringBuffer(buf)
	return s
}

func (b SqlDeleteBuilder) renderSQL(buf *bytes.Buffer, d dialect.SqlDialect) {
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
func (b SqlDeleteBuilder) Dialect(d dialect.SqlDialect) SqlDeleteBuilder {
	b.dialect = d
	return b
}

// From sets the target table for the DELETE statement.
//
//	b.From("users")
func (b SqlDeleteBuilder) From(table string) SqlDeleteBuilder {
	b.from = table
	return b
}

// Where sets the WHERE condition.
//
//	b.Where(sqblpg.Eq("id", sqbl.P(1)))
func (b SqlDeleteBuilder) Where(expr syntax.SqlFragment) SqlDeleteBuilder {
	b.where = expr
	return b
}

// Returning adds a RETURNING clause to the DELETE statement.
// The specified columns are returned for each deleted row.
//
//	sqblpg.DeleteFrom("sessions").Where(...).Returning("id", "user_id")
func (b SqlDeleteBuilder) Returning(cols ...string) SqlDeleteBuilder {
	b.returning = cols
	return b
}
