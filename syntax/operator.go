package syntax

import (
	"bytes"

	"github.com/kotofurumiya/sqbl/dialect"
)

var _ SqlFragment = ComparisonExpr{}
var _ SqlFragment = LogicalExpr{}

// ComparisonExpr represents a binary comparison such as col = value.
// Left and Right accept the same values as AppendValue; a plain string is an identifier.
type ComparisonExpr struct {
	Left     any
	Operator string
	Right    any
}

// AppendSQL implements SqlFragment, writing the comparison into buf.
func (e ComparisonExpr) AppendSQL(buf *bytes.Buffer, d dialect.SqlDialect) {
	AppendValue(buf, d, e.Left)
	buf.WriteByte(' ')
	buf.WriteString(e.Operator)
	buf.WriteByte(' ')
	AppendValue(buf, d, e.Right)
}

// Eq returns an equality (=) comparison expression.
// A plain string on either side is an identifier, so compare against a value
// with P, or with Raw for a literal written into the SQL.
//
//	syntax.Eq("status", syntax.P(1))        // "status" = $1
//	syntax.Eq("u.id", "o.user_id")          // "u"."id" = "o"."user_id"
//	syntax.Eq("status", syntax.Raw("'ok'")) // "status" = 'ok'
//	syntax.Eq(syntax.Fn("SUM", "x"), 100)   // SUM("x") = 100
func Eq(left any, right any) ComparisonExpr {
	return ComparisonExpr{Left: left, Operator: "=", Right: right}
}

// Ne returns an inequality (<>) comparison expression.
//
//	syntax.Ne("status", syntax.P(1))  // "status" <> $1
func Ne(left any, right any) ComparisonExpr {
	return ComparisonExpr{Left: left, Operator: "<>", Right: right}
}

// Lt returns a less-than (<) comparison expression.
//
//	syntax.Lt("age", 18)  // "age" < 18
func Lt(left any, right any) ComparisonExpr {
	return ComparisonExpr{Left: left, Operator: "<", Right: right}
}

// Lte returns a less-than-or-equal (<=) comparison expression.
//
//	syntax.Lte("age", 18)  // "age" <= 18
func Lte(left any, right any) ComparisonExpr {
	return ComparisonExpr{Left: left, Operator: "<=", Right: right}
}

// Gt returns a greater-than (>) comparison expression.
//
//	syntax.Gt("score", 100)              // "score" > 100
//	syntax.Gt(syntax.Fn("SUM", "x"), 0) // SUM("x") > 0
func Gt(left any, right any) ComparisonExpr {
	return ComparisonExpr{Left: left, Operator: ">", Right: right}
}

// Gte returns a greater-than-or-equal (>=) comparison expression.
//
//	syntax.Gte("score", 100)  // "score" >= 100
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
//	// "active" = TRUE AND "age" > 18
func And(exprs ...SqlFragment) LogicalExpr {
	return LogicalExpr{Operator: "AND", Expressions: exprs}
}

// Or combines multiple expressions with OR.
//
//	syntax.Or(syntax.Eq("role", syntax.P(1)), syntax.Eq("role", syntax.P(2)))
//	// "role" = $1 OR "role" = $2
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
// Left and each value accept the same values as AppendValue.
type InExpr struct {
	Left   any
	Negate bool
	Values []any
}

var _ SqlFragment = InExpr{}

// AppendSQL implements SqlFragment, writing the IN or NOT IN expression into buf.
func (e InExpr) AppendSQL(buf *bytes.Buffer, d dialect.SqlDialect) {
	AppendValue(buf, d, e.Left)
	if e.Negate {
		buf.WriteString(" NOT IN (")
	} else {
		buf.WriteString(" IN (")
	}
	for i, v := range e.Values {
		if i > 0 {
			buf.WriteString(", ")
		}
		AppendValue(buf, d, v)
	}
	buf.WriteByte(')')
}

// In returns an IN expression.
//
//	syntax.In("status", syntax.P(1), syntax.P(2))
//	// "status" IN ($1, $2)
//
//	syntax.In("id", subquery)
//	// "id" IN (SELECT ...)
func In(left any, values ...any) InExpr {
	return InExpr{Left: left, Negate: false, Values: values}
}

// NotIn returns a NOT IN expression.
//
//	syntax.NotIn("status", syntax.Raw("'deleted'"), syntax.Raw("'banned'"))
//	// "status" NOT IN ('deleted', 'banned')
func NotIn(left any, values ...any) InExpr {
	return InExpr{Left: left, Negate: true, Values: values}
}

// NullExpr represents an IS NULL or IS NOT NULL expression.
type NullExpr struct {
	Col    any // accepts the same values as AppendValue
	Negate bool
}

var _ SqlFragment = NullExpr{}

// AppendSQL implements SqlFragment, writing IS NULL or IS NOT NULL into buf.
func (e NullExpr) AppendSQL(buf *bytes.Buffer, d dialect.SqlDialect) {
	AppendValue(buf, d, e.Col)
	if e.Negate {
		buf.WriteString(" IS NOT NULL")
	} else {
		buf.WriteString(" IS NULL")
	}
}

// IsNull returns an IS NULL expression.
//
//	syntax.IsNull("deleted_at")
//	// "deleted_at" IS NULL
func IsNull(col any) NullExpr {
	return NullExpr{Col: col, Negate: false}
}

// IsNotNull returns an IS NOT NULL expression.
//
//	syntax.IsNotNull("email")
//	// "email" IS NOT NULL
func IsNotNull(col any) NullExpr {
	return NullExpr{Col: col, Negate: true}
}

// BetweenExpr represents a BETWEEN ... AND ... expression.
// Col and both bounds accept the same values as AppendValue.
type BetweenExpr struct {
	Col  any
	Low  any
	High any
}

var _ SqlFragment = BetweenExpr{}

// AppendSQL implements SqlFragment, writing the BETWEEN expression into buf.
func (e BetweenExpr) AppendSQL(buf *bytes.Buffer, d dialect.SqlDialect) {
	AppendValue(buf, d, e.Col)
	buf.WriteString(" BETWEEN ")
	AppendValue(buf, d, e.Low)
	buf.WriteString(" AND ")
	AppendValue(buf, d, e.High)
}

// Between returns a BETWEEN ... AND ... expression.
//
//	syntax.Between("age", 18, 65)
//	// "age" BETWEEN 18 AND 65
func Between(col any, low, high any) BetweenExpr {
	return BetweenExpr{Col: col, Low: low, High: high}
}

// Like returns a LIKE comparison expression.
// The pattern accepts the same values as AppendValue; pass it with P, or with Raw
// for a literal written into the SQL.
//
//	syntax.Like("name", syntax.P(1))
//	// "name" LIKE $1
//
//	syntax.Like(syntax.Fn("LOWER", "name"), syntax.Raw("'%foo%'"))
//	// LOWER("name") LIKE '%foo%'
func Like(left any, pattern any) ComparisonExpr {
	return ComparisonExpr{Left: left, Operator: "LIKE", Right: pattern}
}

// ILike returns an ILIKE comparison expression (case-insensitive LIKE, PostgreSQL-specific).
// The pattern accepts the same values as Like.
//
//	syntax.ILike("name", syntax.P(1))
//	// "name" ILIKE $1
func ILike(left any, pattern any) ComparisonExpr {
	return ComparisonExpr{Left: left, Operator: "ILIKE", Right: pattern}
}
