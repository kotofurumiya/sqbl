package sqblsqlite

import "github.com/kotofurumiya/sqbl/syntax"

// Ident is an identifier such as a column or table name. It is quoted, and
// each dot-separated part is quoted separately. A plain string passed where an
// expression is expected is treated the same way.
//
//	sqblsqlite.Ident("users.id") // "users"."id"
type Ident = syntax.Ident

// Raw is a SQL fragment written as-is, without quoting or escaping.
// Never build it from untrusted input; pass values with P instead.
//
//	sqblsqlite.Raw("CURRENT_TIMESTAMP")
//	sqblsqlite.Raw("'pending'")
type Raw = syntax.Raw

// As creates an aliased expression for use in Select(), From(), or Join().
// The source accepts the same values as Eq; the alias is quoted as a single identifier.
//
//	sqblsqlite.As("users", "u")                            // "users" AS "u"
//	sqblsqlite.As(sqblsqlite.Fn("SUM", "amount"), "total") // SUM("amount") AS "total"
//	sqblsqlite.As(subquery, "sub")                         // (SELECT ...) AS "sub"
func As(source any, alias string) syntax.Aliased {
	return syntax.As(source, alias)
}

// Eq returns an equality (=) comparison expression.
// A plain string on either side is an identifier, so compare against a value
// with P, or with Raw for a literal written into the SQL.
// A nil operand fails to render; use IsNull instead.
//
//	sqblsqlite.Eq("status", sqblsqlite.P(1))        // "status" = ?1
//	sqblsqlite.Eq("u.id", "o.user_id")              // "u"."id" = "o"."user_id"
//	sqblsqlite.Eq("status", sqblsqlite.Raw("'ok'")) // "status" = 'ok'
func Eq(left any, right any) syntax.ComparisonExpr {
	return syntax.Eq(left, right)
}

// Ne returns an inequality (<>) comparison expression.
//
//	sqblsqlite.Ne("status", sqblsqlite.P(1)) // "status" <> ?1
func Ne(left any, right any) syntax.ComparisonExpr {
	return syntax.Ne(left, right)
}

// Lt returns a less-than (<) comparison expression.
//
//	sqblsqlite.Lt("age", 18) // "age" < 18
func Lt(left any, right any) syntax.ComparisonExpr {
	return syntax.Lt(left, right)
}

// Lte returns a less-than-or-equal (<=) comparison expression.
//
//	sqblsqlite.Lte("age", 18) // "age" <= 18
func Lte(left any, right any) syntax.ComparisonExpr {
	return syntax.Lte(left, right)
}

// Gt returns a greater-than (>) comparison expression.
//
//	sqblsqlite.Gt("score", 100) // "score" > 100
func Gt(left any, right any) syntax.ComparisonExpr {
	return syntax.Gt(left, right)
}

// Gte returns a greater-than-or-equal (>=) comparison expression.
//
//	sqblsqlite.Gte("score", 100) // "score" >= 100
func Gte(left any, right any) syntax.ComparisonExpr {
	return syntax.Gte(left, right)
}

// And combines conditions with AND. With no conditions it renders TRUE,
// so a filter list built at run time can be empty.
//
//	sqblsqlite.And(sqblsqlite.Eq("active", true), sqblsqlite.Gt("age", 18))
//	// "active" = 1 AND "age" > 18
func And(exprs ...syntax.SQLFragment) syntax.LogicalExpr {
	return syntax.And(exprs...)
}

// Or combines conditions with OR. With no conditions it renders FALSE,
// so a filter list built at run time can be empty.
//
//	sqblsqlite.Or(sqblsqlite.Eq("role", sqblsqlite.P(1)), sqblsqlite.Eq("role", sqblsqlite.P(2)))
//	// "role" = ?1 OR "role" = ?2
func Or(exprs ...syntax.SQLFragment) syntax.LogicalExpr {
	return syntax.Or(exprs...)
}

// Not negates a condition.
//
//	sqblsqlite.Not(sqblsqlite.Eq("active", false)) // NOT ("active" = 0)
func Not(expr syntax.SQLFragment) syntax.NotExpr {
	return syntax.Not(expr)
}

// Asc creates an ascending ORDER BY expression.
// The expression accepts the same values as Eq, so a plain string is an identifier.
//
//	sqblsqlite.Asc("name")                         // "name" ASC
//	sqblsqlite.Asc(sqblsqlite.Fn("LOWER", "name")) // LOWER("name") ASC
func Asc(expr any) syntax.Order {
	return syntax.Asc(expr)
}

// Desc creates a descending ORDER BY expression.
// The expression accepts the same values as Eq, so a plain string is an identifier.
//
//	sqblsqlite.Desc("created_at") // "created_at" DESC
func Desc(expr any) syntax.Order {
	return syntax.Desc(expr)
}

// In returns an IN expression. A lone subquery is used as the list itself.
// With no values it renders FALSE, so a list built at run time can be empty.
//
//	sqblsqlite.In("status", sqblsqlite.P(1), sqblsqlite.P(2)) // "status" IN (?1, ?2)
//	sqblsqlite.In("id", subquery)                             // "id" IN (SELECT ...)
func In(left any, values ...any) syntax.InExpr {
	return syntax.In(left, values...)
}

// NotIn returns a NOT IN expression. With no values it renders TRUE.
//
//	sqblsqlite.NotIn("status", sqblsqlite.Raw("'deleted'"), sqblsqlite.Raw("'banned'"))
//	// "status" NOT IN ('deleted', 'banned')
func NotIn(left any, values ...any) syntax.InExpr {
	return syntax.NotIn(left, values...)
}

// IsNull returns an IS NULL expression.
//
//	sqblsqlite.IsNull("deleted_at") // "deleted_at" IS NULL
func IsNull(col any) syntax.NullExpr {
	return syntax.IsNull(col)
}

// IsNotNull returns an IS NOT NULL expression.
//
//	sqblsqlite.IsNotNull("email") // "email" IS NOT NULL
func IsNotNull(col any) syntax.NullExpr {
	return syntax.IsNotNull(col)
}

// Between returns a BETWEEN ... AND ... expression.
//
//	sqblsqlite.Between("age", 18, 65) // "age" BETWEEN 18 AND 65
func Between(col any, low, high any) syntax.BetweenExpr {
	return syntax.Between(col, low, high)
}

// Like returns a LIKE comparison expression.
// The pattern accepts the same values as Eq; pass it with P, or with Raw for a literal.
//
//	sqblsqlite.Like("name", sqblsqlite.P(1))           // "name" LIKE ?1
//	sqblsqlite.Like("name", sqblsqlite.Raw("'%foo%'")) // "name" LIKE '%foo%'
func Like(left any, pattern any) syntax.ComparisonExpr {
	return syntax.Like(left, pattern)
}

// P creates a bind parameter placeholder. The value itself is passed to the
// database driver as a query argument, never written into the SQL.
//
//	sqblsqlite.P()          // ?
//	sqblsqlite.P(1)         // ?1
//	sqblsqlite.P(":status") // :status (also @name and $name)
//
// An index must be an integer of 1 or more, and a name must start with :, @ or $.
// P takes a placeholder, not a value: never pass user input to P.
func P(args ...any) syntax.Parameter {
	return syntax.P(args...)
}

// Fn creates a SQL function call expression. The name is written as-is, so
// never build it from untrusted input. Each argument accepts the same values as Eq.
//
//	sqblsqlite.Fn("SUM", "amount")                         // SUM("amount")
//	sqblsqlite.Fn("COUNT", "*")                            // COUNT(*)
//	sqblsqlite.Fn("COALESCE", "nickname", sqblsqlite.P(1)) // COALESCE("nickname", ?1)
func Fn(name string, args ...any) syntax.FnExpr {
	return syntax.Fn(name, args...)
}

// Cast creates a CAST(expr AS type) expression.
// The type is written as-is, so never build it from untrusted input.
//
//	sqblsqlite.Cast(sqblsqlite.P(1), "BIGINT") // CAST(?1 AS BIGINT)
//	sqblsqlite.Cast("price", "TEXT")           // CAST("price" AS TEXT)
func Cast(expr any, typ string) syntax.CastExpr {
	return syntax.Cast(expr, typ)
}

// Over wraps an expression with a window OVER clause.
//
//	sqblsqlite.Over(sqblsqlite.Fn("ROW_NUMBER")).PartitionBy("dept").OrderBy(sqblsqlite.Desc("salary"))
//	// ROW_NUMBER() OVER (PARTITION BY "dept" ORDER BY "salary" DESC)
func Over(expr any) syntax.WindowExpr {
	return syntax.Over(expr)
}
