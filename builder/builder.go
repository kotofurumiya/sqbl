// Package builder holds the statement builders behind sqbl's entry points:
// SELECT and its set operations, INSERT, UPDATE, DELETE, and the DDL statements.
//
// Builders are immutable values. Every method returns a modified copy, so a
// partly built statement can be shared and extended in several directions
// without the branches affecting each other.
//
// Start from a dialect package such as sqblpg rather than from a zero value:
// the dialect decides quoting and placeholders when the statement is rendered.
//
//	q := sqblpg.From("users").Select("id").Where(sqblpg.Eq("id", sqblpg.P(1)))
//	sql, err := q.ToSQL() // SELECT "id" FROM "users" WHERE "id" = $1;
package builder

import (
	"bytes"
	"errors"
	"slices"

	"github.com/kotofurumiya/sqbl/dialect"
	"github.com/kotofurumiya/sqbl/internal/sqlbuf"
	"github.com/kotofurumiya/sqbl/internal/sqlerr"
	"github.com/kotofurumiya/sqbl/syntax"
)

// Statement is implemented by all SQL statement builders.
//
// Rendering fails with an error, instead of producing broken SQL, when the
// statement holds input that has no valid SQL form, such as a value of an
// unsupported Go type, a JOIN without an ON condition, or a missing dialect.
// The Must variants panic instead, for statements built from constants where
// a failure is a programming error.
type Statement interface {
	// ToSQL renders the statement with the builder's own dialect and a trailing semicolon.
	ToSQL() (string, error)
	// ToSQLWithDialect renders the statement with d and no trailing semicolon.
	ToSQLWithDialect(d dialect.Dialect) (string, error)
	// MustSQL is like ToSQL but panics on error.
	MustSQL() string
	// MustSQLWithDialect is like ToSQLWithDialect but panics on error.
	MustSQLWithDialect(d dialect.Dialect) string
}

// renderer is the rendering core every builder implements.
type renderer interface {
	renderSQL(buf *bytes.Buffer, d dialect.Dialect)
}

// render runs r.renderSQL with d and turns a failure raised inside any fragment into an error.
func render(r renderer, d dialect.Dialect, semicolon bool) (sql string, err error) {
	if d == nil {
		return "", errors.New("sqbl: no dialect to render with; use ToSQLWithDialect, " +
			"or start from a dialect package such as sqblpg")
	}
	buf := sqlbuf.GetStringBuffer()
	defer sqlbuf.PutStringBuffer(buf)
	defer sqlerr.Catch(&err)
	r.renderSQL(buf, d)
	if semicolon {
		buf.WriteByte(';')
	}
	return buf.String(), nil
}

// must returns sql, or panics with err.
func must(sql string, err error) string {
	if err != nil {
		panic(err)
	}
	return sql
}

// failFragment stands in for an argument that cannot be rendered.
// Rendering it fails the whole statement with its message, so a mistake made
// while building surfaces as an error from ToSQL instead of as broken SQL.
type failFragment string

// AppendSQL implements syntax.SQLFragment by failing the render.
func (f failFragment) AppendSQL(*bytes.Buffer, dialect.Dialect) {
	sqlerr.Fail("%s", string(f))
}

// source converts a table-like argument of method to a fragment.
// nil would otherwise render as the NULL literal, which is never a valid table.
func source(method string, v any) syntax.SQLFragment {
	if v == nil {
		return failFragment(method + " got nil; pass a table name, an aliased source or a subquery")
	}
	return syntax.ToFragment(v)
}

// addCondition returns conds with expr appended, ignoring a nil expr.
// Conditions are ANDed when rendered, so each call narrows the statement further.
func addCondition(conds []syntax.SQLFragment, expr syntax.SQLFragment) []syntax.SQLFragment {
	if expr == nil {
		return conds
	}
	return append(slices.Clone(conds), expr)
}

// writeConditions writes keyword followed by conds joined with AND.
// Nothing is written when conds is empty.
func writeConditions(buf *bytes.Buffer, d dialect.Dialect, keyword string, conds []syntax.SQLFragment) {
	switch len(conds) {
	case 0:
		return
	case 1:
		buf.WriteString(keyword)
		conds[0].AppendSQL(buf, d)
	default:
		buf.WriteString(keyword)
		syntax.And(conds...).AppendSQL(buf, d)
	}
}

// writeValues writes vals separated by ", ", each following the rules of syntax.AppendValue.
func writeValues(buf *bytes.Buffer, d dialect.Dialect, vals []any) {
	for i, v := range vals {
		if i > 0 {
			buf.WriteString(", ")
		}
		syntax.AppendValue(buf, d, v)
	}
}

// quoteIdentifiers writes each name quoted by the dialect into buf, separated by ", ".
//
//	quoteIdentifiers(buf, d, []string{"id", "name"})  →  "id", "name"
//	quoteIdentifiers(buf, d, nil)                     →  (nothing written)
func quoteIdentifiers(buf *bytes.Buffer, d dialect.Dialect, names []string) {
	for i, name := range names {
		if i > 0 {
			buf.WriteString(", ")
		}
		d.QuoteIdentifier(buf, name)
	}
}
