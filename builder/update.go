package builder

import (
	"bytes"
	"slices"

	"github.com/kotofurumiya/sqbl/dialect"
	"github.com/kotofurumiya/sqbl/syntax"
)

// setClause represents a single SET col = val entry.
type setClause struct {
	Col string
	Val any
}

var _ syntax.SQLFragment = setClause{}

// AppendSQL implements syntax.SQLFragment, writing "col" = val into buf.
func (s setClause) AppendSQL(buf *bytes.Buffer, d dialect.Dialect) {
	d.QuoteIdentifier(buf, s.Col)
	buf.WriteString(" = ")
	syntax.AppendValue(buf, d, s.Val)
}

// UpdateBuilder constructs a SQL UPDATE statement using a fluent method chain.
// Use a dialect-specific constructor (e.g. sqblpg.Update) to create an instance.
//
//	q := sqblpg.Update("users").
//	    Set("name", sqblpg.P(1)).
//	    Set("email", sqblpg.P(2)).
//	    Where(sqblpg.Eq("id", sqblpg.P(3)))
//	sql, err := q.ToSQL()
type UpdateBuilder struct {
	dialect   dialect.Dialect
	table     string
	sets      []setClause
	where     []syntax.SQLFragment // ANDed
	returning []any
}

var _ Statement = UpdateBuilder{}

// ToSQL renders the statement with the builder's dialect and a trailing semicolon.
func (b UpdateBuilder) ToSQL() (string, error) { return render(b, b.dialect, true) }

// ToSQLWithDialect renders the statement with d and no trailing semicolon.
func (b UpdateBuilder) ToSQLWithDialect(d dialect.Dialect) (string, error) {
	return render(b, d, false)
}

// MustSQL is like ToSQL but panics on error.
func (b UpdateBuilder) MustSQL() string { return must(b.ToSQL()) }

// MustSQLWithDialect is like ToSQLWithDialect but panics on error.
func (b UpdateBuilder) MustSQLWithDialect(d dialect.Dialect) string {
	return must(b.ToSQLWithDialect(d))
}

func (b UpdateBuilder) renderSQL(buf *bytes.Buffer, d dialect.Dialect) {
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
			s.AppendSQL(buf, d)
		}
	}

	writeConditions(buf, d, " WHERE ", b.where)

	if len(b.returning) > 0 {
		buf.WriteString(" RETURNING ")
		writeValues(buf, d, b.returning)
	}
}

// Dialect sets the SQL dialect used when rendering the query.
func (b UpdateBuilder) Dialect(d dialect.Dialect) UpdateBuilder {
	b.dialect = d
	return b
}

// Table sets the target table for the UPDATE statement.
//
//	b.Table("users")
func (b UpdateBuilder) Table(table string) UpdateBuilder {
	b.table = table
	return b
}

// Set adds a SET assignment to the UPDATE statement.
// Call Set multiple times to assign multiple columns.
// The value accepts the same values as syntax.AppendValue, so a plain string
// is an identifier; use syntax.Raw for an expression.
//
//	b.Set("name", sqblpg.P(1)).Set("email", sqblpg.P(2))
//	b.Set("count", sqblpg.Raw(`"count" + 1`))
//	b.Set("deleted_at", nil) // "deleted_at" = NULL
func (b UpdateBuilder) Set(col string, val any) UpdateBuilder {
	b.sets = append(append([]setClause(nil), b.sets...), setClause{Col: col, Val: val})
	return b
}

// Where adds a WHERE condition.
// Calling Where again ANDs the new condition with the earlier ones; nil is ignored.
// Without any condition, every row is updated.
//
//	b.Where(sqblpg.Eq("id", sqblpg.P(1)))
func (b UpdateBuilder) Where(expr syntax.SQLFragment) UpdateBuilder {
	b.where = addCondition(b.where, expr)
	return b
}

// Returning sets the RETURNING list, replacing any earlier list.
// Each item follows the rules of syntax.AppendValue, so "*" returns every column.
// Supported by PostgreSQL, SQLite 3.35+, and MariaDB 10.5+.
//
//	sqblpg.Update("users").Set("name", sqblpg.P(1)).Where(sqblpg.Eq("id", sqblpg.P(2))).Returning("id", "name")
func (b UpdateBuilder) Returning(cols ...any) UpdateBuilder {
	b.returning = slices.Clone(cols)
	return b
}
