package sqblsqlite

import "github.com/kotofurumiya/sqbl/syntax"

// Ident is an identifier such as a column or table name. It is quoted, and
// each dot-separated part is quoted separately. A plain string passed where an
// expression is expected is treated the same way.
//
//	sqblsqlite.Ident("users.id")
type Ident = syntax.Ident

// Raw is a SQL fragment written as-is. Never build it from untrusted input.
//
//	sqblsqlite.Raw("n + 1")
//	sqblsqlite.Raw("'pending'")
type Raw = syntax.Raw

// As creates an aliased expression for use in Select(), From(), or Join().
//
//	sqblsqlite.As("users", "u")                   // users AS u
//	sqblsqlite.As(sqblsqlite.Fn("SUM", "amount"), "total") // SUM(amount) AS total
//	sqblsqlite.As(subquery, "sub")                // (SELECT ...) AS sub
func As(source any, alias string) syntax.Aliased {
	return syntax.As(source, alias)
}

// Eq returns a = comparison expression.
//
//	sqblsqlite.Eq("status", sqblsqlite.P(1)) // status = <param>
//	sqblsqlite.Eq("u.id", "o.user_id")   // u.id = o.user_id (a plain string is an identifier)
func Eq(left any, right any) syntax.ComparisonExpr {
	return syntax.Eq(left, right)
}

// Ne returns a <> comparison expression.
//
//	sqblsqlite.Ne("status", sqblsqlite.P(1)) // status <> <param>
func Ne(left any, right any) syntax.ComparisonExpr {
	return syntax.Ne(left, right)
}

// Lt returns a < comparison expression.
//
//	sqblsqlite.Lt("age", 18) // age < 18
func Lt(left any, right any) syntax.ComparisonExpr {
	return syntax.Lt(left, right)
}

// Lte returns a <= comparison expression.
//
//	sqblsqlite.Lte("age", 18) // age <= 18
func Lte(left any, right any) syntax.ComparisonExpr {
	return syntax.Lte(left, right)
}

// Gt returns a > comparison expression.
//
//	sqblsqlite.Gt("score", 100) // score > 100
func Gt(left any, right any) syntax.ComparisonExpr {
	return syntax.Gt(left, right)
}

// Gte returns a >= comparison expression.
//
//	sqblsqlite.Gte("score", 100) // score >= 100
func Gte(left any, right any) syntax.ComparisonExpr {
	return syntax.Gte(left, right)
}

// And combines multiple expressions with AND.
//
//	sqblsqlite.And(sqblsqlite.Eq("active", true), sqblsqlite.Gt("age", 18))
//	// active = true AND age > 18
func And(exprs ...syntax.SqlFragment) syntax.LogicalExpr {
	return syntax.And(exprs...)
}

// Or combines multiple expressions with OR.
//
//	sqblsqlite.Or(sqblsqlite.Eq("role", sqblsqlite.P(1)), sqblsqlite.Eq("role", sqblsqlite.P(2)))
//	// role = <param> OR role = <param>
func Or(exprs ...syntax.SqlFragment) syntax.LogicalExpr {
	return syntax.Or(exprs...)
}

// Asc creates an ascending ORDER BY expression.
//
//	sqblsqlite.Asc("name")  // name ASC
func Asc(col string) syntax.Order {
	return syntax.Asc(col)
}

// Desc creates a descending ORDER BY expression.
//
//	sqblsqlite.Desc("created_at")  // created_at DESC
func Desc(col string) syntax.Order {
	return syntax.Desc(col)
}

// Not wraps an expression with NOT.
//
//	sqblsqlite.Not(sqblsqlite.Eq("active", false))
//	// NOT ("active" = 0)
func Not(expr syntax.SqlFragment) syntax.NotExpr {
	return syntax.Not(expr)
}

// In returns an IN expression.
//
//	sqblsqlite.In("status", sqblsqlite.P(1), sqblsqlite.P(2))
//	// status IN (<param>, <param>)
func In(left any, values ...any) syntax.InExpr {
	return syntax.In(left, values...)
}

// NotIn returns a NOT IN expression.
//
//	sqblsqlite.NotIn("status", sqblsqlite.Raw("'deleted'"), sqblsqlite.Raw("'banned'"))
//	// status NOT IN ('deleted', 'banned')
func NotIn(left any, values ...any) syntax.InExpr {
	return syntax.NotIn(left, values...)
}

// IsNull returns an IS NULL expression.
//
//	sqblsqlite.IsNull("deleted_at")
//	// deleted_at IS NULL
func IsNull(col any) syntax.NullExpr {
	return syntax.IsNull(col)
}

// IsNotNull returns an IS NOT NULL expression.
//
//	sqblsqlite.IsNotNull("email")
//	// email IS NOT NULL
func IsNotNull(col any) syntax.NullExpr {
	return syntax.IsNotNull(col)
}

// Between returns a BETWEEN ... AND ... expression.
//
//	sqblsqlite.Between("age", 18, 65)
//	// age BETWEEN 18 AND 65
func Between(col any, low, high any) syntax.BetweenExpr {
	return syntax.Between(col, low, high)
}

// Like returns a LIKE comparison expression.
// The pattern accepts the same values as Eq; pass it with P, or with Raw for a literal.
//
//	sqblsqlite.Like("name", sqblsqlite.Raw("'%foo%'"))
//	// name LIKE '%foo%'
func Like(left any, pattern any) syntax.ComparisonExpr {
	return syntax.Like(left, pattern)
}

// P creates a bind parameter placeholder.
//
//	sqblsqlite.P()          → positional: ?
//	sqblsqlite.P(1)         → indexed:    ?
//	sqblsqlite.P(":status") → named:      :status
func P(args ...any) syntax.Parameter {
	return syntax.P(args...)
}

// Fn creates a SQL function call expression.
//
//	sqblsqlite.Fn("SUM", "amount")   // SUM(amount)
//	sqblsqlite.Fn("COUNT", "*")      // COUNT(*)
func Fn(name string, args ...any) syntax.SqlFn {
	return syntax.Fn(name, args...)
}

// Over wraps an expression with a window OVER clause.
//
//	sqblsqlite.Over(sqblsqlite.Fn("ROW_NUMBER")).PartitionBy("dept").OrderBy("salary")
func Over(expr any) syntax.WindowExpr {
	return syntax.Over(expr)
}
