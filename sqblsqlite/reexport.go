package sqblsqlite

import "github.com/kotofurumiya/sqbl/syntax"

// As creates an aliased expression for use in Select(), From(), or Join().
//
//	sqblsqlite.As("users", "u")                   // users AS u
//	sqblsqlite.As("SUM(p.amount)", "total_spent") // SUM(p.amount) AS total_spent
//	sqblsqlite.As(subquery, "sub")                // (SELECT ...) AS sub
func As(source any, alias string) *syntax.Aliased {
	return syntax.As(source, alias)
}

// Eq returns a = comparison expression.
//
//	sqblsqlite.Eq("status", "active") // status = active
func Eq(left any, right any) syntax.ComparisonExpr {
	return syntax.Eq(left, right)
}

// Ne returns a <> comparison expression.
//
//	sqblsqlite.Ne("status", "deleted") // status <> deleted
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
//	sqblsqlite.Or(sqblsqlite.Eq("role", "admin"), sqblsqlite.Eq("role", "moderator"))
//	// role = admin OR role = moderator
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
//	sqblsqlite.In("status", "'active'", "'pending'")
//	// status IN ('active', 'pending')
func In(left any, values ...any) syntax.InExpr {
	return syntax.In(left, values...)
}

// NotIn returns a NOT IN expression.
//
//	sqblsqlite.NotIn("status", "'deleted'", "'banned'")
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
// The pattern should be a SQL string literal including quotes (e.g. "'%foo%'").
//
//	sqblsqlite.Like("name", "'%foo%'")
//	// name LIKE '%foo%'
func Like(left any, pattern string) syntax.ComparisonExpr {
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
func Fn(name string, args ...any) *syntax.SqlFn {
	return syntax.Fn(name, args...)
}

// Over wraps an expression with a window OVER clause.
//
//	sqblsqlite.Over(sqblsqlite.Fn("ROW_NUMBER")).PartitionBy("dept").OrderBy("salary")
func Over(expr any) *syntax.WindowExpr {
	return syntax.Over(expr)
}
