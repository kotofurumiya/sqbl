package sqblpg

import "github.com/kotofurumiya/sqbl/syntax"

// As creates an aliased expression for use in Select(), From(), or Join().
//
//	sqblpg.As("users", "u")                   // users AS u
//	sqblpg.As("SUM(p.amount)", "total_spent") // SUM(p.amount) AS total_spent
//	sqblpg.As(subquery, "sub")                // (SELECT ...) AS sub
func As(source any, alias string) syntax.Aliased {
	return syntax.As(source, alias)
}

// Eq returns a = comparison expression.
//
//	sqblpg.Eq("status", "active") // status = active
func Eq(left any, right any) syntax.ComparisonExpr {
	return syntax.Eq(left, right)
}

// Ne returns a <> comparison expression.
//
//	sqblpg.Ne("status", "deleted") // status <> deleted
func Ne(left any, right any) syntax.ComparisonExpr {
	return syntax.Ne(left, right)
}

// Lt returns a < comparison expression.
//
//	sqblpg.Lt("age", 18) // age < 18
func Lt(left any, right any) syntax.ComparisonExpr {
	return syntax.Lt(left, right)
}

// Lte returns a <= comparison expression.
//
//	sqblpg.Lte("age", 18) // age <= 18
func Lte(left any, right any) syntax.ComparisonExpr {
	return syntax.Lte(left, right)
}

// Gt returns a > comparison expression.
//
//	sqblpg.Gt("score", 100) // score > 100
func Gt(left any, right any) syntax.ComparisonExpr {
	return syntax.Gt(left, right)
}

// Gte returns a >= comparison expression.
//
//	sqblpg.Gte("score", 100) // score >= 100
func Gte(left any, right any) syntax.ComparisonExpr {
	return syntax.Gte(left, right)
}

// And combines multiple expressions with AND.
//
//	sqblpg.And(sqblpg.Eq("active", true), sqblpg.Gt("age", 18))
//	// active = true AND age > 18
func And(exprs ...syntax.SqlFragment) syntax.LogicalExpr {
	return syntax.And(exprs...)
}

// Or combines multiple expressions with OR.
//
//	sqblpg.Or(sqblpg.Eq("role", "admin"), sqblpg.Eq("role", "moderator"))
//	// role = admin OR role = moderator
func Or(exprs ...syntax.SqlFragment) syntax.LogicalExpr {
	return syntax.Or(exprs...)
}

// Asc creates an ascending ORDER BY expression.
//
//	sqblpg.Asc("name")  // name ASC
func Asc(col string) syntax.Order {
	return syntax.Asc(col)
}

// Desc creates a descending ORDER BY expression.
//
//	sqblpg.Desc("created_at")  // created_at DESC
func Desc(col string) syntax.Order {
	return syntax.Desc(col)
}

// Not wraps an expression with NOT.
//
//	sqblpg.Not(sqblpg.Eq("active", false))
//	// NOT ("active" = FALSE)
func Not(expr syntax.SqlFragment) syntax.NotExpr {
	return syntax.Not(expr)
}

// In returns an IN expression.
//
//	sqblpg.In("status", "'active'", "'pending'")
//	// status IN ('active', 'pending')
func In(left any, values ...any) syntax.InExpr {
	return syntax.In(left, values...)
}

// NotIn returns a NOT IN expression.
//
//	sqblpg.NotIn("status", "'deleted'", "'banned'")
//	// status NOT IN ('deleted', 'banned')
func NotIn(left any, values ...any) syntax.InExpr {
	return syntax.NotIn(left, values...)
}

// IsNull returns an IS NULL expression.
//
//	sqblpg.IsNull("deleted_at")
//	// deleted_at IS NULL
func IsNull(col any) syntax.NullExpr {
	return syntax.IsNull(col)
}

// IsNotNull returns an IS NOT NULL expression.
//
//	sqblpg.IsNotNull("email")
//	// email IS NOT NULL
func IsNotNull(col any) syntax.NullExpr {
	return syntax.IsNotNull(col)
}

// Between returns a BETWEEN ... AND ... expression.
//
//	sqblpg.Between("age", 18, 65)
//	// age BETWEEN 18 AND 65
func Between(col any, low, high any) syntax.BetweenExpr {
	return syntax.Between(col, low, high)
}

// Like returns a LIKE comparison expression.
// The pattern should be a SQL string literal including quotes (e.g. "'%foo%'").
//
//	sqblpg.Like("name", "'%foo%'")
//	// name LIKE '%foo%'
func Like(left any, pattern string) syntax.ComparisonExpr {
	return syntax.Like(left, pattern)
}

// ILike returns an ILIKE comparison expression (case-insensitive LIKE, PostgreSQL-specific).
// The pattern should be a SQL string literal including quotes (e.g. "'%foo%'").
//
//	sqblpg.ILike("name", "'%foo%'")
//	// name ILIKE '%foo%'
func ILike(left any, pattern string) syntax.ComparisonExpr {
	return syntax.ILike(left, pattern)
}

// P creates a bind parameter placeholder.
//
//	sqblpg.P()          → positional: ?
//	sqblpg.P(1)         → indexed:    $1
//	sqblpg.P(":status") → named:      :status
func P(args ...any) syntax.Parameter {
	return syntax.P(args...)
}

// Fn creates a SQL function call expression.
//
//	sqblpg.Fn("SUM", "amount")          // SUM(amount)
//	sqblpg.Fn("COUNT", "*")             // COUNT(*)
//	sqblpg.Fn("COALESCE", "x", sqbl.P(1)) // COALESCE(x, $1)
func Fn(name string, args ...any) syntax.SqlFn {
	return syntax.Fn(name, args...)
}

// Over wraps an expression with a window OVER clause.
//
//	sqblpg.Over(sqblpg.Fn("ROW_NUMBER")).PartitionBy("dept").OrderBy("salary")
func Over(expr any) syntax.WindowExpr {
	return syntax.Over(expr)
}
