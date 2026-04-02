package syntax

import (
	"fmt"
	"strings"

	"github.com/kotofurumiya/sqbl/dialect"
)

// isSimpleIdentifier reports whether s should be treated as an identifier
// (quoted via QuoteIdentifier) rather than a raw SQL expression (passed through as-is).
//
// Characters that indicate a SQL expression rather than an identifier:
//   - '(' — function calls such as COUNT(*), SUM(amount), NOW()
//   - '\'' — string literals such as 'value', '2025-01-01'
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

// SqlFragment is a type that can render itself as a SQL string for a given dialect.
type SqlFragment interface {
	ToSqlWithDialect(d dialect.SqlDialect) string
}

var _ SqlFragment = &StringSource{}

// StringSource is a SqlFragment backed by a raw SQL string.
// It renders the string as-is regardless of dialect.
type StringSource struct {
	str string
}

// NewStringSource creates a StringSource from a raw SQL string.
func NewStringSource(s string) *StringSource {
	return &StringSource{str: s}
}

// ToSqlWithDialect implements SqlFragment. Simple identifiers are quoted; expressions are returned as-is.
func (s *StringSource) ToSqlWithDialect(d dialect.SqlDialect) string {
	if isSimpleIdentifier(s.str) {
		return d.QuoteIdentifier(s.str)
	}
	return s.str
}

// ToFragment converts an arbitrary value to a SqlFragment.
func ToFragment(v any) SqlFragment {
	switch v := v.(type) {
	case SqlFragment:
		return v
	case string:
		return NewStringSource(v)
	default:
		return NewStringSource(fmt.Sprint(v))
	}
}

// Aliased represents an expression with an AS alias (e.g. "users" AS "u").
type Aliased struct {
	Source SqlFragment // StringSource | *SqlSelectBuilder
	Alias  string
}

var _ SqlFragment = (*Aliased)(nil)

// ToSqlWithDialect implements SqlFragment, rendering the source followed by AS alias.
func (a *Aliased) ToSqlWithDialect(d dialect.SqlDialect) string {
	srcSQL := a.Source.ToSqlWithDialect(d)
	return srcSQL + " AS " + d.Quote(a.Alias)
}

// Order represents an ORDER BY expression with a direction.
type Order struct {
	col string
	dir string // "ASC" or "DESC"
}

var _ SqlFragment = Order{}

// ToSqlWithDialect implements SqlFragment, rendering the column name followed by ASC or DESC.
func (o Order) ToSqlWithDialect(d dialect.SqlDialect) string {
	return d.QuoteIdentifier(o.col) + " " + o.dir
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
func As(source any, alias string) *Aliased {
	return &Aliased{Source: ToFragment(source), Alias: alias}
}
