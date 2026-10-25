package syntax

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/kotofurumiya/sqbl/dialect"
)

// SqlFragment is a type that can render itself into a SQL buffer for a given dialect.
type SqlFragment interface {
	AppendSQL(buf *bytes.Buffer, d dialect.SqlDialect)
}

var (
	_ SqlFragment = Ident("")
	_ SqlFragment = Raw("")
)

// Ident is an identifier such as a column or table name.
// It is quoted by the dialect, and each dot-separated part is quoted separately,
// so a name that itself contains a dot cannot be expressed with Ident.
// A plain string passed where an expression is expected is treated as an Ident.
//
// "*" is written as-is, and a trailing ".*" is kept unquoted.
//
//	syntax.Ident("users.id") // "users"."id"
//	syntax.Ident("u.*")      // "u".*
type Ident string

// AppendSQL implements SqlFragment, writing the quoted identifier into buf.
func (i Ident) AppendSQL(buf *bytes.Buffer, d dialect.SqlDialect) {
	s := string(i)
	if s == "*" {
		buf.WriteByte('*')
		return
	}
	if table, ok := strings.CutSuffix(s, ".*"); ok {
		d.QuoteIdentifier(buf, table)
		buf.WriteString(".*")
		return
	}
	d.QuoteIdentifier(buf, s)
}

// Raw is a SQL fragment written as-is, without quoting or escaping.
// Never build a Raw from untrusted input; pass values with P instead.
//
//	syntax.Raw("n + 1")
//	syntax.Raw("'pending'")
type Raw string

// AppendSQL implements SqlFragment, writing the string into buf unchanged.
func (r Raw) AppendSQL(buf *bytes.Buffer, _ dialect.SqlDialect) {
	buf.WriteString(string(r))
}

// boolValue is a boolean literal rendered with the dialect's representation.
type boolValue bool

// AppendSQL implements SqlFragment, writing the dialect's boolean literal into buf.
func (b boolValue) AppendSQL(buf *bytes.Buffer, d dialect.SqlDialect) {
	d.Bool(buf, bool(b))
}

// ToFragment converts a value accepted by expression arguments to a SqlFragment.
// See AppendValue for how each type is interpreted.
func ToFragment(v any) SqlFragment {
	switch v := v.(type) {
	case SqlFragment:
		return v
	case string:
		return Ident(v)
	case bool:
		return boolValue(v)
	case nil:
		return Raw("NULL")
	default:
		return Raw(fmt.Sprint(v))
	}
}

// AppendValue writes a value accepted by expression arguments into buf.
//   - SqlFragment: rendered by its own AppendSQL.
//   - string: an identifier (same as Ident). Use Raw for SQL text and P for values.
//   - bool: the dialect's boolean literal.
//   - nil: NULL.
//   - anything else: formatted with fmt, e.g. numbers.
//
// It is equivalent to ToFragment(v).AppendSQL(buf, d) without the allocation.
func AppendValue(buf *bytes.Buffer, d dialect.SqlDialect, v any) {
	switch v := v.(type) {
	case SqlFragment:
		v.AppendSQL(buf, d)
	case string:
		Ident(v).AppendSQL(buf, d)
	case bool:
		d.Bool(buf, v)
	case nil:
		buf.WriteString("NULL")
	default:
		fmt.Fprint(buf, v)
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
// The source accepts the same values as AppendValue; the alias is quoted as a single identifier.
//
//	syntax.As("users", "u")                                 // "users" AS "u"
//	syntax.As(syntax.Fn("SUM", "p.amount"), "total_spent") // SUM("p"."amount") AS "total_spent"
//	syntax.As(subquery, "sub")                              // (SELECT ...) AS "sub"
func As(source any, alias string) Aliased {
	return Aliased{Source: ToFragment(source), Alias: alias}
}
