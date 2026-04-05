package syntax

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/kotofurumiya/sqbl/dialect"
)

var _ SqlFragment = ComparisonExpr{}
var _ SqlFragment = LogicalExpr{}

// ComparisonExpr represents a binary comparison such as col = value.
type ComparisonExpr struct {
	Left     any // string identifier, SqlFragment (e.g. Fn), or raw expression
	Operator string
	Right    any
}

// AppendSQL implements SqlFragment, writing the comparison into buf.
func (e ComparisonExpr) AppendSQL(buf *bytes.Buffer, d dialect.SqlDialect) {
	switch v := e.Left.(type) {
	case string:
		if isSimpleIdentifier(v) {
			d.QuoteIdentifier(buf, v)
		} else {
			buf.WriteString(v)
		}
	case SqlFragment:
		v.AppendSQL(buf, d)
	default:
		buf.WriteString(fmt.Sprint(v))
	}

	buf.WriteByte(' ')
	buf.WriteString(e.Operator)
	buf.WriteByte(' ')

	switch v := e.Right.(type) {
	case SqlFragment:
		v.AppendSQL(buf, d)
	case bool:
		d.Bool(buf, v)
	case string:
		// isSimpleIdentifier already excludes strings containing '(' or '\''.
		// The extra '"' check lets callers pass pre-quoted identifiers (e.g. "table"."col")
		// through as-is rather than double-quoting them.
		if isSimpleIdentifier(v) && !strings.ContainsAny(v, "\"") {
			d.QuoteIdentifier(buf, v)
		} else {
			buf.WriteString(v)
		}
	default:
		buf.WriteString(fmt.Sprintf("%v", v))
	}
}

// Eq returns an equality (=) comparison expression.
//
//	syntax.Eq("status", "active")         // status = active
//	syntax.Eq(syntax.Fn("SUM", "x"), 100) // SUM(x) = 100
func Eq(left any, right any) ComparisonExpr {
	return ComparisonExpr{Left: left, Operator: "=", Right: right}
}

// Ne returns an inequality (<>) comparison expression.
//
//	syntax.Ne("status", "deleted")  // status <> deleted
func Ne(left any, right any) ComparisonExpr {
	return ComparisonExpr{Left: left, Operator: "<>", Right: right}
}

// Lt returns a less-than (<) comparison expression.
//
//	syntax.Lt("age", 18)  // age < 18
func Lt(left any, right any) ComparisonExpr {
	return ComparisonExpr{Left: left, Operator: "<", Right: right}
}

// Lte returns a less-than-or-equal (<=) comparison expression.
//
//	syntax.Lte("age", 18)  // age <= 18
func Lte(left any, right any) ComparisonExpr {
	return ComparisonExpr{Left: left, Operator: "<=", Right: right}
}

// Gt returns a greater-than (>) comparison expression.
//
//	syntax.Gt("score", 100)              // score > 100
//	syntax.Gt(syntax.Fn("SUM", "x"), 0) // SUM(x) > 0
func Gt(left any, right any) ComparisonExpr {
	return ComparisonExpr{Left: left, Operator: ">", Right: right}
}

// Gte returns a greater-than-or-equal (>=) comparison expression.
//
//	syntax.Gte("score", 100)  // score >= 100
func Gte(left any, right any) ComparisonExpr {
	return ComparisonExpr{Left: left, Operator: ">=", Right: right}
}

// LogicalExpr combines multiple SQL fragments with a logical operator (AND/OR).
type LogicalExpr struct {
	Operator    string
	Expressions []SqlFragment
}

// AppendSQL implements SqlFragment, writing all sub-expressions joined by the logical operator into buf.
// Nested LogicalExpr are wrapped in parentheses to preserve evaluation order.
func (e LogicalExpr) AppendSQL(buf *bytes.Buffer, d dialect.SqlDialect) {
	for i, exp := range e.Expressions {
		if i > 0 {
			buf.WriteByte(' ')
			buf.WriteString(e.Operator)
			buf.WriteByte(' ')
		}
		if _, ok := exp.(LogicalExpr); ok {
			buf.WriteByte('(')
			exp.AppendSQL(buf, d)
			buf.WriteByte(')')
		} else {
			exp.AppendSQL(buf, d)
		}
	}
}

// And combines multiple expressions with AND.
//
//	syntax.And(syntax.Eq("active", true), syntax.Gt("age", 18))
//	// active = TRUE AND age > 18
func And(exprs ...SqlFragment) LogicalExpr {
	return LogicalExpr{Operator: "AND", Expressions: exprs}
}

// Or combines multiple expressions with OR.
//
//	syntax.Or(syntax.Eq("role", "admin"), syntax.Eq("role", "moderator"))
//	// role = admin OR role = moderator
func Or(exprs ...SqlFragment) LogicalExpr {
	return LogicalExpr{Operator: "OR", Expressions: exprs}
}

// NotExpr represents a NOT (...) expression.
type NotExpr struct {
	Expr SqlFragment
}

var _ SqlFragment = NotExpr{}

// AppendSQL implements SqlFragment, writing NOT (...) into buf.
func (e NotExpr) AppendSQL(buf *bytes.Buffer, d dialect.SqlDialect) {
	buf.WriteString("NOT (")
	e.Expr.AppendSQL(buf, d)
	buf.WriteByte(')')
}

// Not wraps an expression with NOT.
//
//	syntax.Not(syntax.Eq("active", false))
//	// NOT ("active" = FALSE)
func Not(expr SqlFragment) NotExpr {
	return NotExpr{Expr: expr}
}

// InExpr represents an IN or NOT IN expression.
type InExpr struct {
	Left   any // string identifier, SqlFragment (e.g. Fn), or raw expression
	Negate bool
	Values []any
}

var _ SqlFragment = InExpr{}

// AppendSQL implements SqlFragment, writing the IN or NOT IN expression into buf.
func (e InExpr) AppendSQL(buf *bytes.Buffer, d dialect.SqlDialect) {
	ToFragment(e.Left).AppendSQL(buf, d)
	if e.Negate {
		buf.WriteString(" NOT IN (")
	} else {
		buf.WriteString(" IN (")
	}
	for i, v := range e.Values {
		if i > 0 {
			buf.WriteString(", ")
		}
		switch val := v.(type) {
		case SqlFragment:
			val.AppendSQL(buf, d)
		case bool:
			d.Bool(buf, val)
		case string:
			// String values are SQL literals supplied by the caller (e.g. "'active'").
			buf.WriteString(val)
		default:
			buf.WriteString(fmt.Sprintf("%v", val))
		}
	}
	buf.WriteByte(')')
}

// In returns an IN expression.
//
//	syntax.In("status", "'active'", "'pending'")
//	// status IN ('active', 'pending')
//
//	syntax.In(syntax.Fn("LOWER", "email"), P(1))
//	// LOWER(email) IN ($1)
func In(left any, values ...any) InExpr {
	return InExpr{Left: left, Negate: false, Values: values}
}

// NotIn returns a NOT IN expression.
//
//	syntax.NotIn("status", "'deleted'", "'banned'")
//	// status NOT IN ('deleted', 'banned')
func NotIn(left any, values ...any) InExpr {
	return InExpr{Left: left, Negate: true, Values: values}
}

// NullExpr represents an IS NULL or IS NOT NULL expression.
type NullExpr struct {
	Col    any // string identifier, SqlFragment (e.g. Fn), or raw expression
	Negate bool
}

var _ SqlFragment = NullExpr{}

// AppendSQL implements SqlFragment, writing IS NULL or IS NOT NULL into buf.
func (e NullExpr) AppendSQL(buf *bytes.Buffer, d dialect.SqlDialect) {
	ToFragment(e.Col).AppendSQL(buf, d)
	if e.Negate {
		buf.WriteString(" IS NOT NULL")
	} else {
		buf.WriteString(" IS NULL")
	}
}

// IsNull returns an IS NULL expression.
//
//	syntax.IsNull("deleted_at")
//	// deleted_at IS NULL
func IsNull(col any) NullExpr {
	return NullExpr{Col: col, Negate: false}
}

// IsNotNull returns an IS NOT NULL expression.
//
//	syntax.IsNotNull("email")
//	// email IS NOT NULL
func IsNotNull(col any) NullExpr {
	return NullExpr{Col: col, Negate: true}
}

// BetweenExpr represents a BETWEEN ... AND ... expression.
type BetweenExpr struct {
	Col  any // string identifier, SqlFragment (e.g. Fn), or raw expression
	Low  any
	High any
}

var _ SqlFragment = BetweenExpr{}

// AppendSQL implements SqlFragment, writing the BETWEEN expression into buf.
func (e BetweenExpr) AppendSQL(buf *bytes.Buffer, d dialect.SqlDialect) {
	ToFragment(e.Col).AppendSQL(buf, d)
	buf.WriteString(" BETWEEN ")
	appendBound(buf, d, e.Low)
	buf.WriteString(" AND ")
	appendBound(buf, d, e.High)
}

// appendBound writes a single BETWEEN bound value into buf.
// Bound values are SQL literals supplied by the caller (e.g. 18, "'2025-01-01'").
func appendBound(buf *bytes.Buffer, d dialect.SqlDialect, v any) {
	switch val := v.(type) {
	case SqlFragment:
		val.AppendSQL(buf, d)
	case bool:
		d.Bool(buf, val)
	case string:
		buf.WriteString(val)
	default:
		buf.WriteString(fmt.Sprintf("%v", val))
	}
}

// Between returns a BETWEEN ... AND ... expression.
//
//	syntax.Between("age", 18, 65)
//	// age BETWEEN 18 AND 65
func Between(col any, low, high any) BetweenExpr {
	return BetweenExpr{Col: col, Low: low, High: high}
}

// Like returns a LIKE comparison expression.
// The pattern should be a SQL string literal including quotes (e.g. "'%foo%'").
//
//	syntax.Like("name", "'%foo%'")
//	// name LIKE '%foo%'
//
//	syntax.Like(syntax.Fn("LOWER", "name"), "'%foo%'")
//	// LOWER(name) LIKE '%foo%'
func Like(left any, pattern string) ComparisonExpr {
	return ComparisonExpr{Left: left, Operator: "LIKE", Right: pattern}
}

// ILike returns an ILIKE comparison expression (case-insensitive LIKE, PostgreSQL-specific).
// The pattern should be a SQL string literal including quotes (e.g. "'%foo%'").
//
//	syntax.ILike("name", "'%foo%'")
//	// name ILIKE '%foo%'
//
//	syntax.ILike(syntax.Fn("LOWER", "name"), "'%foo%'")
//	// LOWER(name) ILIKE '%foo%'
func ILike(left any, pattern string) ComparisonExpr {
	return ComparisonExpr{Left: left, Operator: "ILIKE", Right: pattern}
}
