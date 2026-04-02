package syntax

import (
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

// ToSqlWithDialect renders the comparison as a SQL string using the given dialect.
func (e ComparisonExpr) ToSqlWithDialect(d dialect.SqlDialect) string {
	var left string
	switch v := e.Left.(type) {
	case string:
		if isSimpleIdentifier(v) {
			left = d.QuoteIdentifier(v)
		} else {
			left = v
		}
	case SqlFragment:
		left = v.ToSqlWithDialect(d)
	default:
		left = fmt.Sprint(v)
	}

	var r string
	switch v := e.Right.(type) {
	case SqlFragment:
		r = v.ToSqlWithDialect(d)
	case bool:
		r = d.Bool(v)
	case string:
		// isSimpleIdentifier already excludes strings containing '(' or '\''.
		// The extra '"' check lets callers pass pre-quoted identifiers (e.g. "table"."col")
		// through as-is rather than double-quoting them.
		if isSimpleIdentifier(v) && !strings.ContainsAny(v, "\"") {
			r = d.QuoteIdentifier(v)
		} else {
			r = v
		}
	default:
		r = fmt.Sprintf("%v", v)
	}

	return fmt.Sprintf("%s %s %v", left, e.Operator, r)
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

// ToSqlWithDialect renders all sub-expressions joined by the logical operator.
// Nested LogicalExpr are wrapped in parentheses to preserve evaluation order.
func (e LogicalExpr) ToSqlWithDialect(d dialect.SqlDialect) string {
	strs := make([]string, 0, len(e.Expressions))
	for _, exp := range e.Expressions {
		s := exp.ToSqlWithDialect(d)
		if _, ok := exp.(LogicalExpr); ok {
			s = "(" + s + ")"
		}
		strs = append(strs, s)
	}

	sep := fmt.Sprintf(" %s ", e.Operator)

	return strings.Join(strs, sep)
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

// ToSqlWithDialect renders the expression wrapped in NOT (...).
func (e NotExpr) ToSqlWithDialect(d dialect.SqlDialect) string {
	return "NOT (" + e.Expr.ToSqlWithDialect(d) + ")"
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

// ToSqlWithDialect renders the IN or NOT IN expression.
func (e InExpr) ToSqlWithDialect(d dialect.SqlDialect) string {
	left := ToFragment(e.Left).ToSqlWithDialect(d)

	vals := make([]string, 0, len(e.Values))
	for _, v := range e.Values {
		switch val := v.(type) {
		case SqlFragment:
			vals = append(vals, val.ToSqlWithDialect(d))
		case bool:
			vals = append(vals, d.Bool(val))
		case string:
			// String values are SQL literals supplied by the caller (e.g. "'active'").
			vals = append(vals, val)
		default:
			vals = append(vals, fmt.Sprintf("%v", val))
		}
	}

	op := "IN"
	if e.Negate {
		op = "NOT IN"
	}
	return fmt.Sprintf("%s %s (%s)", left, op, strings.Join(vals, ", "))
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

// ToSqlWithDialect renders the IS NULL or IS NOT NULL expression.
func (e NullExpr) ToSqlWithDialect(d dialect.SqlDialect) string {
	col := ToFragment(e.Col).ToSqlWithDialect(d)
	if e.Negate {
		return col + " IS NOT NULL"
	}
	return col + " IS NULL"
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

// ToSqlWithDialect renders the BETWEEN expression.
func (e BetweenExpr) ToSqlWithDialect(d dialect.SqlDialect) string {
	col := ToFragment(e.Col).ToSqlWithDialect(d)

	// Bound values are SQL literals supplied by the caller (e.g. 18, "'2025-01-01'").
	renderBound := func(v any) string {
		switch val := v.(type) {
		case SqlFragment:
			return val.ToSqlWithDialect(d)
		case bool:
			return d.Bool(val)
		case string:
			return val
		default:
			return fmt.Sprintf("%v", val)
		}
	}

	return fmt.Sprintf("%s BETWEEN %s AND %s", col, renderBound(e.Low), renderBound(e.High))
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
