package sqbl

import "github.com/kotofurumiya/sqbl/syntax"

// Ident is an identifier such as a column or table name. It is quoted, and
// each dot-separated part is quoted separately. A plain string passed where an
// expression is expected is treated the same way.
//
//	sqbl.Ident("users.id")
type Ident = syntax.Ident

// Raw is a SQL fragment written as-is. Never build it from untrusted input.
//
//	sqbl.Raw("n + 1")
//	sqbl.Raw("'pending'")
type Raw = syntax.Raw

// As creates an aliased expression for use in Select(), From(), or Join().
//
//	sqbl.As("users", "u")                     // users AS u
//	sqbl.As(sqbl.Fn("SUM", "amount"), "total") // SUM(amount) AS total
//	sqbl.As(subquery, "sub")                  // (SELECT ...) AS sub
func As(source any, alias string) syntax.Aliased {
	return syntax.As(source, alias)
}

// Eq returns a = comparison expression.
//
//	sqbl.Eq("status", sqbl.P(1)) // status = <param>
//	sqbl.Eq("u.id", "o.user_id") // u.id = o.user_id (a plain string is an identifier)
func Eq(left any, right any) syntax.ComparisonExpr {
	return syntax.Eq(left, right)
}

// Ne returns a <> comparison expression.
//
//	sqbl.Ne("status", sqbl.P(1)) // status <> <param>
func Ne(left any, right any) syntax.ComparisonExpr {
	return syntax.Ne(left, right)
}

// Lt returns a < comparison expression.
//
//	sqbl.Lt("age", 18) // age < 18
func Lt(left any, right any) syntax.ComparisonExpr {
	return syntax.Lt(left, right)
}

// Lte returns a <= comparison expression.
//
//	sqbl.Lte("age", 18) // age <= 18
func Lte(left any, right any) syntax.ComparisonExpr {
	return syntax.Lte(left, right)
}

// Gt returns a > comparison expression.
//
//	sqbl.Gt("score", 100) // score > 100
func Gt(left any, right any) syntax.ComparisonExpr {
	return syntax.Gt(left, right)
}

// Gte returns a >= comparison expression.
//
//	sqbl.Gte("score", 100) // score >= 100
func Gte(left any, right any) syntax.ComparisonExpr {
	return syntax.Gte(left, right)
}

// And combines multiple expressions with AND.
//
//	sqbl.And(sqbl.Eq("active", true), sqbl.Gt("age", 18))
//	// active = true AND age > 18
func And(exprs ...syntax.SqlFragment) syntax.LogicalExpr {
	return syntax.And(exprs...)
}

// Or combines multiple expressions with OR.
//
//	sqbl.Or(sqbl.Eq("role", sqbl.P(1)), sqbl.Eq("role", sqbl.P(2)))
//	// role = <param> OR role = <param>
func Or(exprs ...syntax.SqlFragment) syntax.LogicalExpr {
	return syntax.Or(exprs...)
}

// Asc creates an ascending ORDER BY expression.
//
//	sqbl.Asc("name")  // name ASC
func Asc(col string) syntax.Order {
	return syntax.Asc(col)
}

// Desc creates a descending ORDER BY expression.
//
//	sqbl.Desc("created_at")  // created_at DESC
func Desc(col string) syntax.Order {
	return syntax.Desc(col)
}

// Not wraps an expression with NOT.
//
//	sqbl.Not(sqbl.Eq("active", false))
//	// NOT ("active" = FALSE)
func Not(expr syntax.SqlFragment) syntax.NotExpr {
	return syntax.Not(expr)
}

// In returns an IN expression.
//
//	sqbl.In("status", sqbl.P(1), sqbl.P(2))
//	// status IN (<param>, <param>)
func In(left any, values ...any) syntax.InExpr {
	return syntax.In(left, values...)
}

// NotIn returns a NOT IN expression.
//
//	sqbl.NotIn("status", sqbl.Raw("'deleted'"), sqbl.Raw("'banned'"))
//	// status NOT IN ('deleted', 'banned')
func NotIn(left any, values ...any) syntax.InExpr {
	return syntax.NotIn(left, values...)
}

// IsNull returns an IS NULL expression.
//
//	sqbl.IsNull("deleted_at")
//	// deleted_at IS NULL
func IsNull(col any) syntax.NullExpr {
	return syntax.IsNull(col)
}

// IsNotNull returns an IS NOT NULL expression.
//
//	sqbl.IsNotNull("email")
//	// email IS NOT NULL
func IsNotNull(col any) syntax.NullExpr {
	return syntax.IsNotNull(col)
}

// Between returns a BETWEEN ... AND ... expression.
//
//	sqbl.Between("age", 18, 65)
//	// age BETWEEN 18 AND 65
func Between(col any, low, high any) syntax.BetweenExpr {
	return syntax.Between(col, low, high)
}

// Like returns a LIKE comparison expression.
// The pattern accepts the same values as Eq; pass it with P, or with Raw for a literal.
//
//	sqbl.Like("name", sqbl.Raw("'%foo%'"))
//	// name LIKE '%foo%'
func Like(left any, pattern any) syntax.ComparisonExpr {
	return syntax.Like(left, pattern)
}

// ILike returns an ILIKE comparison expression (case-insensitive LIKE, PostgreSQL-specific).
// The pattern accepts the same values as Eq; pass it with P, or with Raw for a literal.
//
//	sqbl.ILike("name", sqbl.Raw("'%foo%'"))
//	// name ILIKE '%foo%'
func ILike(left any, pattern any) syntax.ComparisonExpr {
	return syntax.ILike(left, pattern)
}

// P creates a bind parameter placeholder.
//
//	sqbl.P()          → positional: ?
//	sqbl.P(1)         → indexed:    $1 (PostgreSQL), ? (MySQL, SQLite)
//	sqbl.P(":status") → named:      :status
func P(args ...any) syntax.Parameter {
	return syntax.P(args...)
}

// Fn creates a SQL function call expression.
//
//	sqbl.Fn("SUM", "amount")            // SUM(amount)
//	sqbl.Fn("COUNT", "*")               // COUNT(*)
//	sqbl.Fn("COALESCE", "x", sqbl.P(1)) // COALESCE(x, <param>)
func Fn(name string, args ...any) syntax.SqlFn {
	return syntax.Fn(name, args...)
}

// Over wraps an expression with a window OVER clause.
//
//	sqbl.Over(sqbl.Fn("ROW_NUMBER")).PartitionBy("dept").OrderBy("salary")
func Over(expr any) syntax.WindowExpr {
	return syntax.Over(expr)
}
