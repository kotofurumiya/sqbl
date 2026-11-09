package syntax

import (
	"bytes"
	"math"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/kotofurumiya/sqbl/dialect"
	"github.com/kotofurumiya/sqbl/internal/sqlerr"
)

// SQLFragment is a piece of SQL that renders itself for a given dialect.
//
// AppendSQL writes into buf and must not keep buf after returning: builders
// render into pooled buffers that are reused by later renders.
type SQLFragment interface {
	AppendSQL(buf *bytes.Buffer, d dialect.Dialect)
}

var (
	_ SQLFragment = Ident("")
	_ SQLFragment = Raw("")
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

// AppendSQL implements SQLFragment, writing the quoted identifier into buf.
func (i Ident) AppendSQL(buf *bytes.Buffer, d dialect.Dialect) {
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

// Subquery is a SQLFragment that renders itself enclosed in parentheses, such as a
// SELECT builder. Expressions that need parentheses around their operand, like IN,
// write it as-is instead of adding another pair.
type Subquery interface {
	SQLFragment
	IsSubquery()
}

// Raw is a SQL fragment written as-is, without quoting or escaping.
// Never build a Raw from untrusted input; pass values with P instead.
//
// Raw is never inspected, so it must carry its own parentheses where precedence
// matters: And(Raw("a OR b"), x) renders a OR b AND x.
//
//	syntax.Raw("n + 1")
//	syntax.Raw("'pending'")
type Raw string

// AppendSQL implements SQLFragment, writing the string into buf unchanged.
func (r Raw) AppendSQL(buf *bytes.Buffer, _ dialect.Dialect) {
	buf.WriteString(string(r))
}

// boolValue is a boolean literal rendered with the dialect's representation.
type boolValue bool

// AppendSQL implements SQLFragment, writing the dialect's boolean literal into buf.
func (b boolValue) AppendSQL(buf *bytes.Buffer, d dialect.Dialect) {
	d.Bool(buf, bool(b))
}

// ToFragment converts a value accepted by expression arguments to a SQLFragment.
// See AppendValue for how each type is interpreted. A value of an unsupported
// type is not rejected here; the returned fragment fails when it is rendered.
func ToFragment(v any) SQLFragment {
	switch v := v.(type) {
	case SQLFragment:
		return v
	case string:
		return Ident(v)
	case bool:
		return boolValue(v)
	case nil:
		return Raw("NULL")
	default:
		return literal{v}
	}
}

// AppendValue writes a value accepted by expression arguments into buf.
//   - SQLFragment: rendered by its own AppendSQL.
//   - string: an identifier (same as Ident). Use Raw for SQL text and P for values.
//   - bool, including named bool types: the dialect's boolean literal.
//   - nil: NULL.
//   - integer and floating-point kinds, including named types: a numeric literal.
//     NaN and infinities have no SQL literal and fail.
//
// Any other value fails the render, including named string types, time.Time,
// []byte, pointers and slices: pass such values with P and bind them as arguments.
//
// It is equivalent to ToFragment(v).AppendSQL(buf, d) without the allocation.
func AppendValue(buf *bytes.Buffer, d dialect.Dialect, v any) {
	switch v := v.(type) {
	case SQLFragment:
		v.AppendSQL(buf, d)
	case string:
		Ident(v).AppendSQL(buf, d)
	case bool:
		d.Bool(buf, v)
	case nil:
		buf.WriteString("NULL")
	case int:
		buf.Write(strconv.AppendInt(buf.AvailableBuffer(), int64(v), 10))
	default:
		appendLiteral(buf, d, v)
	}
}

// literal is a Go value rendered as a SQL literal by appendLiteral.
type literal struct{ v any }

// AppendSQL implements SQLFragment.
func (l literal) AppendSQL(buf *bytes.Buffer, d dialect.Dialect) {
	appendLiteral(buf, d, l.v)
}

// appendLiteral writes v as a SQL literal, judged by its kind so that named
// numeric types work, and fails for kinds that have no safe literal form.
func appendLiteral(buf *bytes.Buffer, d dialect.Dialect, v any) {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		buf.Write(strconv.AppendInt(buf.AvailableBuffer(), rv.Int(), 10))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		buf.Write(strconv.AppendUint(buf.AvailableBuffer(), rv.Uint(), 10))
	case reflect.Float32, reflect.Float64:
		f := rv.Float()
		if math.IsNaN(f) || math.IsInf(f, 0) {
			sqlerr.Fail("cannot render %v (%T) as SQL: NaN and infinity have no SQL literal", v, v)
		}
		buf.Write(strconv.AppendFloat(buf.AvailableBuffer(), f, 'g', -1, rv.Type().Bits()))
	case reflect.Bool:
		d.Bool(buf, rv.Bool())
	case reflect.String:
		sqlerr.Fail("cannot render %T as SQL: convert it to string for an identifier, "+
			"or pass it with P and bind it as an argument", v)
	default:
		sqlerr.Fail("cannot render %T as SQL: pass it with P and bind it as an argument", v)
	}
}

// Aliased represents an expression with an AS alias (e.g. "users" AS "u"),
// optionally followed by column aliases (e.g. "ids" ("id")).
// Create one with As or AsColumns.
type Aliased struct {
	source  SQLFragment
	alias   string
	columns []string
}

var _ SQLFragment = Aliased{}

// AppendSQL implements SQLFragment, writing the source followed by AS alias
// and any column aliases into buf.
func (a Aliased) AppendSQL(buf *bytes.Buffer, d dialect.Dialect) {
	if a.source == nil {
		sqlerr.Fail("alias %q has no source; pass a table name, an expression, or a subquery to As", a.alias)
	}
	a.source.AppendSQL(buf, d)
	buf.WriteString(" AS ")
	d.Quote(buf, a.alias)
	if len(a.columns) == 0 {
		return
	}
	buf.WriteString(" (")
	for i, col := range a.columns {
		if i > 0 {
			buf.WriteString(", ")
		}
		d.Quote(buf, col)
	}
	buf.WriteByte(')')
}

// Order represents an ORDER BY expression with a direction.
type Order struct {
	expr SQLFragment
	dir  string // "ASC" or "DESC"
}

var _ SQLFragment = Order{}

// AppendSQL implements SQLFragment, writing the expression followed by ASC or DESC into buf.
func (o Order) AppendSQL(buf *bytes.Buffer, d dialect.Dialect) {
	o.expr.AppendSQL(buf, d)
	buf.WriteByte(' ')
	buf.WriteString(o.dir)
}

// Asc creates an ascending ORDER BY expression.
// The expression accepts the same values as AppendValue, so a plain string is an identifier.
//
//	syntax.Asc("name")                     // "name" ASC
//	syntax.Asc(syntax.Fn("LOWER", "name")) // LOWER("name") ASC
func Asc(expr any) Order {
	return Order{expr: ToFragment(expr), dir: "ASC"}
}

// Desc creates a descending ORDER BY expression.
// The expression accepts the same values as AppendValue, so a plain string is an identifier.
//
//	syntax.Desc("created_at") // "created_at" DESC
func Desc(expr any) Order {
	return Order{expr: ToFragment(expr), dir: "DESC"}
}

// As creates an aliased expression for use in Select(), From(), or Join().
//
// The source accepts the same values as AppendValue; the alias is quoted as a single identifier.
// A nil source fails when rendered.
//
//	syntax.As("users", "u")                                 // "users" AS "u"
//	syntax.As(syntax.Fn("SUM", "p.amount"), "total_spent") // SUM("p"."amount") AS "total_spent"
//	syntax.As(subquery, "sub")                              // (SELECT ...) AS "sub"
func As(source any, alias string) Aliased {
	return Aliased{source: aliasSource(source), alias: alias}
}

// AsColumns creates a table alias with column aliases, for naming the columns
// of a function call, VALUES list, or subquery used in From() or Join().
// The alias and each column are quoted as single identifiers.
// SQLite does not support column aliases.
//
//	syntax.AsColumns(syntax.Fn("unnest", syntax.P(1)), "ids", "id") // unnest($1) AS "ids" ("id")
func AsColumns(source any, alias string, columns ...string) Aliased {
	return Aliased{source: aliasSource(source), alias: alias, columns: slices.Clone(columns)}
}

// aliasSource converts source like ToFragment, except that nil stays nil:
// "NULL AS x" is never what an alias was meant for.
func aliasSource(source any) SQLFragment {
	if source == nil {
		return nil
	}
	return ToFragment(source)
}
