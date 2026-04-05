package syntax

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/kotofurumiya/sqbl/dialect"
)

// isSimpleIdentifier reports whether s should be treated as an identifier
// (quoted via QuoteIdentifier) rather than a raw SQL expression (passed through as-is).
//
// Characters that indicate a SQL expression rather than an identifier:
//   - '(' — function calls such as COUNT(*), SUM(amount), NOW()
//   - '\” — string literals such as 'value', '2025-01-01'
//
// Special tokens that are always passed through as-is:
//   - '*' — the wildcard token used in COUNT(*) and SELECT *
//
// Spaces are intentionally NOT excluded: SQL identifiers may legally contain
// spaces when quoted (e.g. "my column"), so a space is not a reliable signal
// that a string is an expression rather than an identifier name.
func isSimpleIdentifier(s string) bool {
	if s == "*" {
		return false
	}
	return !strings.ContainsAny(s, "('")
}

// SqlFragment is a type that can render itself into a SQL buffer for a given dialect.
type SqlFragment interface {
	AppendSQL(buf *bytes.Buffer, d dialect.SqlDialect)
}

var _ SqlFragment = StringExpr("")

// StringExpr is a SqlFragment backed by a plain string.
// Simple identifiers (no parentheses or quotes) are quoted by the dialect;
// expressions such as "COUNT(*)" are passed through as-is.
type StringExpr string

// NewStringExpr creates a StringExpr from a plain SQL string or identifier.
func NewStringExpr(s string) StringExpr {
	return StringExpr(s)
}

// AppendSQL implements SqlFragment. Simple identifiers are quoted; expressions are written as-is.
func (s StringExpr) AppendSQL(buf *bytes.Buffer, d dialect.SqlDialect) {
	if isSimpleIdentifier(string(s)) {
		d.QuoteIdentifier(buf, string(s))
	} else {
		buf.WriteString(string(s))
	}
}

// ToFragment converts an arbitrary value to a SqlFragment.
func ToFragment(v any) SqlFragment {
	switch v := v.(type) {
	case SqlFragment:
		return v
	case string:
		return NewStringExpr(v)
	default:
		return NewStringExpr(fmt.Sprint(v))
	}
}

// Aliased represents an expression with an AS alias (e.g. "users" AS "u").
type Aliased struct {
	Source SqlFragment // StringSource | *SqlSelectBuilder
	Alias  string
}

var _ SqlFragment = Aliased{}

// AppendSQL implements SqlFragment, writing the source followed by AS alias into buf.
func (a Aliased) AppendSQL(buf *bytes.Buffer, d dialect.SqlDialect) {
	a.Source.AppendSQL(buf, d)
	buf.WriteString(" AS ")
	d.Quote(buf, a.Alias)
}

// Order represents an ORDER BY expression with a direction.
type Order struct {
	col string
	dir string // "ASC" or "DESC"
}

var _ SqlFragment = Order{}

// AppendSQL implements SqlFragment, writing the column name followed by ASC or DESC into buf.
func (o Order) AppendSQL(buf *bytes.Buffer, d dialect.SqlDialect) {
	d.QuoteIdentifier(buf, o.col)
	buf.WriteByte(' ')
	buf.WriteString(o.dir)
}

// Asc creates an ascending ORDER BY expression.
//
//	syntax.Asc("name")  // name ASC
func Asc(col string) Order {
	return Order{col: col, dir: "ASC"}
}

// Desc creates a descending ORDER BY expression.
//
//	syntax.Desc("created_at")  // created_at DESC
func Desc(col string) Order {
	return Order{col: col, dir: "DESC"}
}

// As creates an aliased expression for use in Select(), From(), or Join().
//
//	syntax.As("users", "u")                   // users AS u
//	syntax.As("SUM(p.amount)", "total_spent") // SUM(p.amount) AS total_spent
//	syntax.As(subquery, "sub")                // (SELECT ...) AS sub
func As(source any, alias string) Aliased {
	return Aliased{Source: ToFragment(source), Alias: alias}
}
