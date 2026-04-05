package builder

import (
	"bytes"

	"github.com/kotofurumiya/sqbl/dialect"
	"github.com/kotofurumiya/sqbl/internal/sqlbuf"
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

var _ SqlBuilder = SqlUpdateBuilder{}

// ToSql renders the UPDATE statement with a trailing semicolon.
func (b SqlUpdateBuilder) ToSql() string {
	buf := sqlbuf.GetStringBuffer()
	b.renderSQL(buf, b.dialect)
	buf.WriteByte(';')
	s := buf.String()
	sqlbuf.PutStringBuffer(buf)
	return s
}

// ToSqlWithDialect renders the UPDATE statement using the given dialect, without a trailing semicolon.
func (b SqlUpdateBuilder) ToSqlWithDialect(d dialect.SqlDialect) string {
	buf := sqlbuf.GetStringBuffer()
	b.renderSQL(buf, d)
	s := buf.String()
	sqlbuf.PutStringBuffer(buf)
	return s
}

func (b SqlUpdateBuilder) renderSQL(buf *bytes.Buffer, d dialect.SqlDialect) {
	// UPDATE table
	buf.WriteString("UPDATE")
	if b.table != "" {
		buf.WriteByte(' ')
		d.QuoteIdentifier(buf, b.table)
	}

	// SET col = val, col = val, ...
	if len(b.sets) > 0 {
		buf.WriteString(" SET ")
		for i, s := range b.sets {
			if i > 0 {
				buf.WriteString(", ")
			}
			d.QuoteIdentifier(buf, s.Col)
			buf.WriteString(" = ")
			appendValue(buf, d, s.Val)
		}
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
func (b SqlUpdateBuilder) Dialect(d dialect.SqlDialect) SqlUpdateBuilder {
	b.dialect = d
	return b
}

// Table sets the target table for the UPDATE statement.
//
//	b.Table("users")
func (b SqlUpdateBuilder) Table(table string) SqlUpdateBuilder {
	b.table = table
	return b
}

// Set adds a SET assignment to the UPDATE statement.
// Call Set multiple times to assign multiple columns.
//
//	b.Set("name", sqbl.P(1)).Set("email", sqbl.P(2))
func (b SqlUpdateBuilder) Set(col string, val any) SqlUpdateBuilder {
	b.sets = append(append([]setClause(nil), b.sets...), setClause{Col: col, Val: val})
	return b
}

// Where sets the WHERE condition.
//
//	b.Where(sqblpg.Eq("id", sqbl.P(1)))
func (b SqlUpdateBuilder) Where(expr syntax.SqlFragment) SqlUpdateBuilder {
	b.where = expr
	return b
}

// Returning adds a RETURNING clause to the UPDATE statement.
// The specified columns are returned for each updated row.
//
//	sqblpg.Update("users").Set("name", sqbl.P(1)).Where(...).Returning("id", "name")
func (b SqlUpdateBuilder) Returning(cols ...string) SqlUpdateBuilder {
	b.returning = cols
	return b
}
