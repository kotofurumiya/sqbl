package syntax

import (
	"bytes"
	"slices"

	"github.com/kotofurumiya/sqbl/dialect"
	"github.com/kotofurumiya/sqbl/internal/sqlerr"
)

var _ SQLFragment = ComparisonExpr{}
var _ SQLFragment = LogicalExpr{}

// appendOperand writes an operand of a comparison, IN, BETWEEN or IS NULL.
// And, Or and Not bind looser than those operators, so they are parenthesized
// to keep their meaning.
func appendOperand(buf *bytes.Buffer, d dialect.Dialect, v any) {
	switch v.(type) {
	case LogicalExpr, NotExpr:
		buf.WriteByte('(')
		AppendValue(buf, d, v)
		buf.WriteByte(')')
	default:
		AppendValue(buf, d, v)
	}
}

// ComparisonExpr represents a binary comparison such as col = value.
// Create one with Eq, Ne, Lt, Lte, Gt, Gte, Like or ILike.
type ComparisonExpr struct {
	left  any
	op    string
	right any
}

// AppendSQL implements SQLFragment, writing the comparison into buf.
// A nil operand fails: comparing with NULL never matches.
func (e ComparisonExpr) AppendSQL(buf *bytes.Buffer, d dialect.Dialect) {
	if e.left == nil || e.right == nil {
		sqlerr.Fail("%s with nil never matches, because NULL is not equal to anything; "+
			"use IsNull or IsNotNull", e.op)
	}
	appendOperand(buf, d, e.left)
	buf.WriteByte(' ')
	buf.WriteString(e.op)
	buf.WriteByte(' ')
	appendOperand(buf, d, e.right)
}

// Eq returns an equality (=) comparison expression.
// Both operands accept the same values as AppendValue. A plain string on either
// side is an identifier, so compare against a value with P, or with Raw for a
// literal written into the SQL. A nil operand fails; use IsNull instead.
//
//	syntax.Eq("status", syntax.P(1))        // "status" = $1
//	syntax.Eq("u.id", "o.user_id")          // "u"."id" = "o"."user_id"
//	syntax.Eq("status", syntax.Raw("'ok'")) // "status" = 'ok'
//	syntax.Eq(syntax.Fn("SUM", "x"), 100)   // SUM("x") = 100
func Eq(left any, right any) ComparisonExpr {
	return ComparisonExpr{left: left, op: "=", right: right}
}

// Ne returns an inequality (<>) comparison expression.
//
//	syntax.Ne("status", syntax.P(1)) // "status" <> $1
func Ne(left any, right any) ComparisonExpr {
	return ComparisonExpr{left: left, op: "<>", right: right}
}

// Lt returns a less-than (<) comparison expression.
//
//	syntax.Lt("age", 18) // "age" < 18
func Lt(left any, right any) ComparisonExpr {
	return ComparisonExpr{left: left, op: "<", right: right}
}

// Lte returns a less-than-or-equal (<=) comparison expression.
//
//	syntax.Lte("age", 18) // "age" <= 18
func Lte(left any, right any) ComparisonExpr {
	return ComparisonExpr{left: left, op: "<=", right: right}
}

// Gt returns a greater-than (>) comparison expression.
//
//	syntax.Gt("score", 100)             // "score" > 100
//	syntax.Gt(syntax.Fn("SUM", "x"), 0) // SUM("x") > 0
func Gt(left any, right any) ComparisonExpr {
	return ComparisonExpr{left: left, op: ">", right: right}
}

// Gte returns a greater-than-or-equal (>=) comparison expression.
//
//	syntax.Gte("score", 100) // "score" >= 100
func Gte(left any, right any) ComparisonExpr {
	return ComparisonExpr{left: left, op: ">=", right: right}
}

// Like returns a LIKE comparison expression.
// The pattern accepts the same values as Eq; pass it with P, or with Raw for a
// literal written into the SQL.
//
//	syntax.Like("name", syntax.P(1))
//	// "name" LIKE $1
//
//	syntax.Like(syntax.Fn("LOWER", "name"), syntax.Raw("'%foo%'"))
//	// LOWER("name") LIKE '%foo%'
func Like(left any, pattern any) ComparisonExpr {
	return ComparisonExpr{left: left, op: "LIKE", right: pattern}
}

// ILike returns an ILIKE comparison expression (case-insensitive LIKE, PostgreSQL-specific).
// The pattern accepts the same values as Eq.
//
//	syntax.ILike("name", syntax.P(1))
//	// "name" ILIKE $1
func ILike(left any, pattern any) ComparisonExpr {
	return ComparisonExpr{left: left, op: "ILIKE", right: pattern}
}

// LogicalExpr combines conditions with AND or OR.
// Create one with And or Or.
type LogicalExpr struct {
	op    string
	exprs []SQLFragment
	empty bool // the identity value rendered when exprs is empty
}

// AppendSQL implements SQLFragment, writing the conditions joined by the operator into buf.
// Nested And and Or are parenthesized to keep their meaning.
func (e LogicalExpr) AppendSQL(buf *bytes.Buffer, d dialect.Dialect) {
	if len(e.exprs) == 0 {
		d.Bool(buf, e.empty)
		return
	}
	for i, exp := range e.exprs {
		if exp == nil {
			sqlerr.Fail("%s operand %d is nil", e.op, i+1)
		}
		if i > 0 {
			buf.WriteByte(' ')
			buf.WriteString(e.op)
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

// And combines conditions with AND.
// With no conditions it renders TRUE, so a filter list built at run time can be empty.
//
//	syntax.And(syntax.Eq("active", true), syntax.Gt("age", 18))
//	// "active" = TRUE AND "age" > 18
func And(exprs ...SQLFragment) LogicalExpr {
	return LogicalExpr{op: "AND", exprs: slices.Clone(exprs), empty: true}
}

// Or combines conditions with OR.
// With no conditions it renders FALSE, so a filter list built at run time can be empty.
//
//	syntax.Or(syntax.Eq("role", syntax.P(1)), syntax.Eq("role", syntax.P(2)))
//	// "role" = $1 OR "role" = $2
func Or(exprs ...SQLFragment) LogicalExpr {
	return LogicalExpr{op: "OR", exprs: slices.Clone(exprs), empty: false}
}

// NotExpr represents a NOT (...) expression.
// Create one with Not.
type NotExpr struct {
	expr SQLFragment
}

var _ SQLFragment = NotExpr{}

// AppendSQL implements SQLFragment, writing NOT (...) into buf.
func (e NotExpr) AppendSQL(buf *bytes.Buffer, d dialect.Dialect) {
	if e.expr == nil {
		sqlerr.Fail("Not has a nil condition")
	}
	buf.WriteString("NOT (")
	e.expr.AppendSQL(buf, d)
	buf.WriteByte(')')
}

// Not negates a condition.
//
//	syntax.Not(syntax.Eq("active", false))
//	// NOT ("active" = FALSE)
func Not(expr SQLFragment) NotExpr {
	return NotExpr{expr: expr}
}

// InExpr represents an IN or NOT IN expression.
// Create one with In or NotIn.
type InExpr struct {
	left   any
	negate bool
	values []any
}

var _ SQLFragment = InExpr{}

// AppendSQL implements SQLFragment, writing the IN or NOT IN expression into buf.
func (e InExpr) AppendSQL(buf *bytes.Buffer, d dialect.Dialect) {
	if len(e.values) == 0 {
		// x IN () is a syntax error; an empty list matches nothing, so NOT IN matches everything.
		d.Bool(buf, e.negate)
		return
	}
	appendOperand(buf, d, e.left)
	if e.negate {
		buf.WriteString(" NOT IN ")
	} else {
		buf.WriteString(" IN ")
	}
	if len(e.values) == 1 {
		if sub, ok := e.values[0].(Subquery); ok {
			sub.AppendSQL(buf, d)
			return
		}
	}
	buf.WriteByte('(')
	for i, v := range e.values {
		if i > 0 {
			buf.WriteString(", ")
		}
		AppendValue(buf, d, v)
	}
	buf.WriteByte(')')
}

// In returns an IN expression. Each value accepts the same values as AppendValue.
// A lone subquery is used as the list itself. With no values it renders FALSE,
// so a list built at run time can be empty.
//
//	syntax.In("status", syntax.P(1), syntax.P(2))
//	// "status" IN ($1, $2)
//
//	syntax.In("id", subquery)
//	// "id" IN (SELECT ...)
func In(left any, values ...any) InExpr {
	return InExpr{left: left, negate: false, values: slices.Clone(values)}
}

// NotIn returns a NOT IN expression. With no values it renders TRUE.
//
//	syntax.NotIn("status", syntax.Raw("'deleted'"), syntax.Raw("'banned'"))
//	// "status" NOT IN ('deleted', 'banned')
func NotIn(left any, values ...any) InExpr {
	return InExpr{left: left, negate: true, values: slices.Clone(values)}
}

// NullExpr represents an IS NULL or IS NOT NULL expression.
// Create one with IsNull or IsNotNull.
type NullExpr struct {
	col    any
	negate bool
}

var _ SQLFragment = NullExpr{}

// AppendSQL implements SQLFragment, writing IS NULL or IS NOT NULL into buf.
func (e NullExpr) AppendSQL(buf *bytes.Buffer, d dialect.Dialect) {
	appendOperand(buf, d, e.col)
	if e.negate {
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
	return NullExpr{col: col, negate: false}
}

// IsNotNull returns an IS NOT NULL expression.
//
//	syntax.IsNotNull("email")
//	// "email" IS NOT NULL
func IsNotNull(col any) NullExpr {
	return NullExpr{col: col, negate: true}
}

// BetweenExpr represents a BETWEEN ... AND ... expression.
// Create one with Between.
type BetweenExpr struct {
	col  any
	low  any
	high any
}

var _ SQLFragment = BetweenExpr{}

// AppendSQL implements SQLFragment, writing the BETWEEN expression into buf.
func (e BetweenExpr) AppendSQL(buf *bytes.Buffer, d dialect.Dialect) {
	appendOperand(buf, d, e.col)
	buf.WriteString(" BETWEEN ")
	appendOperand(buf, d, e.low)
	buf.WriteString(" AND ")
	appendOperand(buf, d, e.high)
}

// Between returns a BETWEEN ... AND ... expression.
// The column and both bounds accept the same values as AppendValue.
//
//	syntax.Between("age", 18, 65)
//	// "age" BETWEEN 18 AND 65
func Between(col any, low, high any) BetweenExpr {
	return BetweenExpr{col: col, low: low, high: high}
}
