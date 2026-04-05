package sqblmysql

import "github.com/kotofurumiya/sqbl/syntax"

// As creates an aliased expression for use in Select(), From(), or Join().
//
//	sqblmysql.As("users", "u")                   // users AS u
//	sqblmysql.As("SUM(p.amount)", "total_spent") // SUM(p.amount) AS total_spent
//	sqblmysql.As(subquery, "sub")                // (SELECT ...) AS sub
func As(source any, alias string) syntax.Aliased {
	return syntax.As(source, alias)
}

// Eq returns a = comparison expression.
//
//	sqblmysql.Eq("status", "active") // status = active
func Eq(left any, right any) syntax.ComparisonExpr {
	return syntax.Eq(left, right)
}

// Ne returns a <> comparison expression.
//
//	sqblmysql.Ne("status", "deleted") // status <> deleted
func Ne(left any, right any) syntax.ComparisonExpr {
	return syntax.Ne(left, right)
}

// Lt returns a < comparison expression.
//
//	sqblmysql.Lt("age", 18) // age < 18
func Lt(left any, right any) syntax.ComparisonExpr {
	return syntax.Lt(left, right)
}

// Lte returns a <= comparison expression.
//
//	sqblmysql.Lte("age", 18) // age <= 18
func Lte(left any, right any) syntax.ComparisonExpr {
	return syntax.Lte(left, right)
}

// Gt returns a > comparison expression.
//
//	sqblmysql.Gt("score", 100) // score > 100
func Gt(left any, right any) syntax.ComparisonExpr {
	return syntax.Gt(left, right)
}

// Gte returns a >= comparison expression.
//
//	sqblmysql.Gte("score", 100) // score >= 100
func Gte(left any, right any) syntax.ComparisonExpr {
	return syntax.Gte(left, right)
}

// And combines multiple expressions with AND.
//
//	sqblmysql.And(sqblmysql.Eq("active", true), sqblmysql.Gt("age", 18))
//	// active = TRUE AND age > 18
func And(exprs ...syntax.SqlFragment) syntax.LogicalExpr {
	return syntax.And(exprs...)
}

// Or combines multiple expressions with OR.
//
//	sqblmysql.Or(sqblmysql.Eq("role", "admin"), sqblmysql.Eq("role", "moderator"))
//	// role = admin OR role = moderator
func Or(exprs ...syntax.SqlFragment) syntax.LogicalExpr {
	return syntax.Or(exprs...)
}

// Asc creates an ascending ORDER BY expression.
//
//	sqblmysql.Asc("name")  // name ASC
func Asc(col string) syntax.Order {
	return syntax.Asc(col)
}

// Desc creates a descending ORDER BY expression.
//
//	sqblmysql.Desc("created_at")  // created_at DESC
func Desc(col string) syntax.Order {
	return syntax.Desc(col)
}

// Not wraps an expression with NOT.
//
//	sqblmysql.Not(sqblmysql.Eq("active", false))
//	// NOT (`active` = FALSE)
func Not(expr syntax.SqlFragment) syntax.NotExpr {
	return syntax.Not(expr)
}

// In returns an IN expression.
//
//	sqblmysql.In("status", "'active'", "'pending'")
//	// status IN ('active', 'pending')
func In(left any, values ...any) syntax.InExpr {
	return syntax.In(left, values...)
}

// NotIn returns a NOT IN expression.
//
//	sqblmysql.NotIn("status", "'deleted'", "'banned'")
//	// status NOT IN ('deleted', 'banned')
func NotIn(left any, values ...any) syntax.InExpr {
	return syntax.NotIn(left, values...)
}

// IsNull returns an IS NULL expression.
//
//	sqblmysql.IsNull("deleted_at")
//	// deleted_at IS NULL
func IsNull(col any) syntax.NullExpr {
	return syntax.IsNull(col)
}

// IsNotNull returns an IS NOT NULL expression.
//
//	sqblmysql.IsNotNull("email")
//	// email IS NOT NULL
func IsNotNull(col any) syntax.NullExpr {
	return syntax.IsNotNull(col)
}

// Between returns a BETWEEN ... AND ... expression.
//
//	sqblmysql.Between("age", 18, 65)
//	// age BETWEEN 18 AND 65
func Between(col any, low, high any) syntax.BetweenExpr {
	return syntax.Between(col, low, high)
}

// Like returns a LIKE comparison expression.
// The pattern should be a SQL string literal including quotes (e.g. "'%foo%'").
//
//	sqblmysql.Like("name", "'%foo%'")
//	// name LIKE '%foo%'
func Like(left any, pattern string) syntax.ComparisonExpr {
	return syntax.Like(left, pattern)
}

// P creates a bind parameter placeholder.
//
//	sqblmysql.P()          → positional: ?
//	sqblmysql.P(1)         → indexed:    ?
//	sqblmysql.P(":status") → named:      :status
func P(args ...any) syntax.Parameter {
	return syntax.P(args...)
}

// Fn creates a SQL function call expression.
//
//	sqblmysql.Fn("SUM", "amount")   // SUM(amount)
//	sqblmysql.Fn("COUNT", "*")      // COUNT(*)
func Fn(name string, args ...any) syntax.SqlFn {
	return syntax.Fn(name, args...)
}

// Over wraps an expression with a window OVER clause.
//
//	sqblmysql.Over(sqblmysql.Fn("ROW_NUMBER")).PartitionBy("dept").OrderBy("salary")
func Over(expr any) syntax.WindowExpr {
	return syntax.Over(expr)
}
