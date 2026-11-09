package sqblpg

import "github.com/kotofurumiya/sqbl/syntax"

// Ident is an identifier such as a column or table name. It is quoted, and
// each dot-separated part is quoted separately. A plain string passed where an
// expression is expected is treated the same way.
//
//	sqblpg.Ident("users.id") // "users"."id"
type Ident = syntax.Ident

// Raw is a SQL fragment written as-is, without quoting or escaping.
// Never build it from untrusted input; pass values with P instead.
//
//	sqblpg.Raw("CURRENT_TIMESTAMP")
//	sqblpg.Raw("'pending'")
type Raw = syntax.Raw

// As creates an aliased expression for use in Select(), From(), or Join().
// The source accepts the same values as Eq; the alias is quoted as a single identifier.
//
//	sqblpg.As("users", "u")                        // "users" AS "u"
//	sqblpg.As(sqblpg.Fn("SUM", "amount"), "total") // SUM("amount") AS "total"
//	sqblpg.As(subquery, "sub")                     // (SELECT ...) AS "sub"
func As(source any, alias string) syntax.Aliased {
	return syntax.As(source, alias)
}

// AsColumns creates a table alias with column aliases, for naming the columns
// of a function call, VALUES list, or subquery used in From() or Join().
//
//	sqblpg.AsColumns(sqblpg.Fn("unnest", sqblpg.P(1)), "ids", "id") // unnest($1) AS "ids" ("id")
func AsColumns(source any, alias string, columns ...string) syntax.Aliased {
	return syntax.AsColumns(source, alias, columns...)
}

// Eq returns an equality (=) comparison expression.
// A plain string on either side is an identifier, so compare against a value
// with P, or with Raw for a literal written into the SQL.
// A nil operand fails to render; use IsNull instead.
//
//	sqblpg.Eq("status", sqblpg.P(1))        // "status" = $1
//	sqblpg.Eq("u.id", "o.user_id")          // "u"."id" = "o"."user_id"
//	sqblpg.Eq("status", sqblpg.Raw("'ok'")) // "status" = 'ok'
func Eq(left any, right any) syntax.ComparisonExpr {
	return syntax.Eq(left, right)
}

// Ne returns an inequality (<>) comparison expression.
//
//	sqblpg.Ne("status", sqblpg.P(1)) // "status" <> $1
func Ne(left any, right any) syntax.ComparisonExpr {
	return syntax.Ne(left, right)
}

// Lt returns a less-than (<) comparison expression.
//
//	sqblpg.Lt("age", 18) // "age" < 18
func Lt(left any, right any) syntax.ComparisonExpr {
	return syntax.Lt(left, right)
}

// Lte returns a less-than-or-equal (<=) comparison expression.
//
//	sqblpg.Lte("age", 18) // "age" <= 18
func Lte(left any, right any) syntax.ComparisonExpr {
	return syntax.Lte(left, right)
}

// Gt returns a greater-than (>) comparison expression.
//
//	sqblpg.Gt("score", 100) // "score" > 100
func Gt(left any, right any) syntax.ComparisonExpr {
	return syntax.Gt(left, right)
}

// Gte returns a greater-than-or-equal (>=) comparison expression.
//
//	sqblpg.Gte("score", 100) // "score" >= 100
func Gte(left any, right any) syntax.ComparisonExpr {
	return syntax.Gte(left, right)
}

// And combines conditions with AND. With no conditions it renders TRUE,
// so a filter list built at run time can be empty.
//
//	sqblpg.And(sqblpg.Eq("active", true), sqblpg.Gt("age", 18)) // "active" = TRUE AND "age" > 18
func And(exprs ...syntax.SQLFragment) syntax.LogicalExpr {
	return syntax.And(exprs...)
}

// Or combines conditions with OR. With no conditions it renders FALSE,
// so a filter list built at run time can be empty.
//
//	sqblpg.Or(sqblpg.Eq("role", sqblpg.P(1)), sqblpg.Eq("role", sqblpg.P(2)))
//	// "role" = $1 OR "role" = $2
func Or(exprs ...syntax.SQLFragment) syntax.LogicalExpr {
	return syntax.Or(exprs...)
}

// Not negates a condition.
//
//	sqblpg.Not(sqblpg.Eq("active", false)) // NOT ("active" = FALSE)
func Not(expr syntax.SQLFragment) syntax.NotExpr {
	return syntax.Not(expr)
}

// Asc creates an ascending ORDER BY expression.
// The expression accepts the same values as Eq, so a plain string is an identifier.
//
//	sqblpg.Asc("name")                     // "name" ASC
//	sqblpg.Asc(sqblpg.Fn("LOWER", "name")) // LOWER("name") ASC
func Asc(expr any) syntax.Order {
	return syntax.Asc(expr)
}

// Desc creates a descending ORDER BY expression.
// The expression accepts the same values as Eq, so a plain string is an identifier.
//
//	sqblpg.Desc("created_at") // "created_at" DESC
func Desc(expr any) syntax.Order {
	return syntax.Desc(expr)
}

// In returns an IN expression. A lone subquery is used as the list itself.
// With no values it renders FALSE, so a list built at run time can be empty.
//
//	sqblpg.In("status", sqblpg.P(1), sqblpg.P(2)) // "status" IN ($1, $2)
//	sqblpg.In("id", subquery)                     // "id" IN (SELECT ...)
func In(left any, values ...any) syntax.InExpr {
	return syntax.In(left, values...)
}

// NotIn returns a NOT IN expression. With no values it renders TRUE.
//
//	sqblpg.NotIn("status", sqblpg.Raw("'deleted'"), sqblpg.Raw("'banned'"))
//	// "status" NOT IN ('deleted', 'banned')
func NotIn(left any, values ...any) syntax.InExpr {
	return syntax.NotIn(left, values...)
}

// IsNull returns an IS NULL expression.
//
//	sqblpg.IsNull("deleted_at") // "deleted_at" IS NULL
func IsNull(col any) syntax.NullExpr {
	return syntax.IsNull(col)
}

// IsNotNull returns an IS NOT NULL expression.
//
//	sqblpg.IsNotNull("email") // "email" IS NOT NULL
func IsNotNull(col any) syntax.NullExpr {
	return syntax.IsNotNull(col)
}

// Between returns a BETWEEN ... AND ... expression.
//
//	sqblpg.Between("age", 18, 65) // "age" BETWEEN 18 AND 65
func Between(col any, low, high any) syntax.BetweenExpr {
	return syntax.Between(col, low, high)
}

// Like returns a LIKE comparison expression.
// The pattern accepts the same values as Eq; pass it with P, or with Raw for a literal.
//
//	sqblpg.Like("name", sqblpg.P(1))           // "name" LIKE $1
//	sqblpg.Like("name", sqblpg.Raw("'%foo%'")) // "name" LIKE '%foo%'
func Like(left any, pattern any) syntax.ComparisonExpr {
	return syntax.Like(left, pattern)
}

// ILike returns an ILIKE comparison expression (case-insensitive LIKE, PostgreSQL-specific).
// The pattern accepts the same values as Eq; pass it with P, or with Raw for a literal.
//
//	sqblpg.ILike("name", sqblpg.P(1)) // "name" ILIKE $1
func ILike(left any, pattern any) syntax.ComparisonExpr {
	return syntax.ILike(left, pattern)
}

// P creates a bind parameter placeholder. The value itself is passed to the
// database driver as a query argument, never written into the SQL.
//
//	sqblpg.P(1)         // $1, the form pgx and lib/pq expect
//	sqblpg.P()          // ?, only for drivers that rewrite ? themselves
//	sqblpg.P(":status") // :status, only for drivers or tools that bind names
//
// An index must be an integer of 1 or more, and a name must start with :, @ or $.
// P takes a placeholder, not a value: never pass user input to P.
func P(args ...any) syntax.Parameter {
	return syntax.P(args...)
}

// Fn creates a SQL function call expression. The name is written as-is, so
// never build it from untrusted input. Each argument accepts the same values as Eq.
//
//	sqblpg.Fn("SUM", "amount")                     // SUM("amount")
//	sqblpg.Fn("COUNT", "*")                        // COUNT(*)
//	sqblpg.Fn("COALESCE", "nickname", sqblpg.P(1)) // COALESCE("nickname", $1)
func Fn(name string, args ...any) syntax.FnExpr {
	return syntax.Fn(name, args...)
}

// Cast creates a CAST(expr AS type) expression.
// The type is written as-is, so never build it from untrusted input.
//
//	sqblpg.Cast(sqblpg.P(1), "BIGINT") // CAST($1 AS BIGINT)
//	sqblpg.Cast("price", "TEXT")       // CAST("price" AS TEXT)
func Cast(expr any, typ string) syntax.CastExpr {
	return syntax.Cast(expr, typ)
}

// Over wraps an expression with a window OVER clause.
//
//	sqblpg.Over(sqblpg.Fn("ROW_NUMBER")).PartitionBy("dept").OrderBy(sqblpg.Desc("salary"))
//	// ROW_NUMBER() OVER (PARTITION BY "dept" ORDER BY "salary" DESC)
func Over(expr any) syntax.WindowExpr {
	return syntax.Over(expr)
}
