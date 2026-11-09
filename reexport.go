package sqbl

import "github.com/kotofurumiya/sqbl/syntax"

// Ident is an identifier such as a column or table name. It is quoted, and
// each dot-separated part is quoted separately. A plain string passed where an
// expression is expected is treated the same way.
//
//	sqbl.Ident("users.id") // "users"."id"
type Ident = syntax.Ident

// Raw is a SQL fragment written as-is, without quoting or escaping.
// Never build it from untrusted input; pass values with P instead.
//
//	sqbl.Raw("CURRENT_TIMESTAMP")
//	sqbl.Raw("'pending'")
type Raw = syntax.Raw

// As creates an aliased expression for use in Select(), From(), or Join().
// The source accepts the same values as Eq; the alias is quoted as a single identifier.
//
//	sqbl.As("users", "u")                      // "users" AS "u"
//	sqbl.As(sqbl.Fn("SUM", "amount"), "total") // SUM("amount") AS "total"
//	sqbl.As(subquery, "sub")                   // (SELECT ...) AS "sub"
func As(source any, alias string) syntax.Aliased {
	return syntax.As(source, alias)
}

// AsColumns creates a table alias with column aliases, for naming the columns
// of a function call, VALUES list, or subquery used in From() or Join().
//
//	sqbl.AsColumns(sqbl.Fn("unnest", sqbl.P(1)), "ids", "id") // unnest($1) AS "ids" ("id")
func AsColumns(source any, alias string, columns ...string) syntax.Aliased {
	return syntax.AsColumns(source, alias, columns...)
}

// Eq returns an equality (=) comparison expression.
// A plain string on either side is an identifier, so compare against a value
// with P, or with Raw for a literal written into the SQL.
// A nil operand fails to render; use IsNull instead.
//
//	sqbl.Eq("status", sqbl.P(1))        // "status" = $1
//	sqbl.Eq("u.id", "o.user_id")        // "u"."id" = "o"."user_id"
//	sqbl.Eq("status", sqbl.Raw("'ok'")) // "status" = 'ok'
func Eq(left any, right any) syntax.ComparisonExpr {
	return syntax.Eq(left, right)
}

// Ne returns an inequality (<>) comparison expression.
//
//	sqbl.Ne("status", sqbl.P(1)) // "status" <> $1
func Ne(left any, right any) syntax.ComparisonExpr {
	return syntax.Ne(left, right)
}

// Lt returns a less-than (<) comparison expression.
//
//	sqbl.Lt("age", 18) // "age" < 18
func Lt(left any, right any) syntax.ComparisonExpr {
	return syntax.Lt(left, right)
}

// Lte returns a less-than-or-equal (<=) comparison expression.
//
//	sqbl.Lte("age", 18) // "age" <= 18
func Lte(left any, right any) syntax.ComparisonExpr {
	return syntax.Lte(left, right)
}

// Gt returns a greater-than (>) comparison expression.
//
//	sqbl.Gt("score", 100) // "score" > 100
func Gt(left any, right any) syntax.ComparisonExpr {
	return syntax.Gt(left, right)
}

// Gte returns a greater-than-or-equal (>=) comparison expression.
//
//	sqbl.Gte("score", 100) // "score" >= 100
func Gte(left any, right any) syntax.ComparisonExpr {
	return syntax.Gte(left, right)
}

// And combines conditions with AND. With no conditions it renders TRUE,
// so a filter list built at run time can be empty.
//
//	sqbl.And(sqbl.Eq("active", true), sqbl.Gt("age", 18)) // "active" = TRUE AND "age" > 18
func And(exprs ...syntax.SQLFragment) syntax.LogicalExpr {
	return syntax.And(exprs...)
}

// Or combines conditions with OR. With no conditions it renders FALSE,
// so a filter list built at run time can be empty.
//
//	sqbl.Or(sqbl.Eq("role", sqbl.P(1)), sqbl.Eq("role", sqbl.P(2))) // "role" = $1 OR "role" = $2
func Or(exprs ...syntax.SQLFragment) syntax.LogicalExpr {
	return syntax.Or(exprs...)
}

// Not negates a condition.
//
//	sqbl.Not(sqbl.Eq("active", false)) // NOT ("active" = FALSE)
func Not(expr syntax.SQLFragment) syntax.NotExpr {
	return syntax.Not(expr)
}

// Asc creates an ascending ORDER BY expression.
// The expression accepts the same values as Eq, so a plain string is an identifier.
//
//	sqbl.Asc("name")                   // "name" ASC
//	sqbl.Asc(sqbl.Fn("LOWER", "name")) // LOWER("name") ASC
func Asc(expr any) syntax.Order {
	return syntax.Asc(expr)
}

// Desc creates a descending ORDER BY expression.
// The expression accepts the same values as Eq, so a plain string is an identifier.
//
//	sqbl.Desc("created_at") // "created_at" DESC
func Desc(expr any) syntax.Order {
	return syntax.Desc(expr)
}

// In returns an IN expression. A lone subquery is used as the list itself.
// With no values it renders FALSE, so a list built at run time can be empty.
//
//	sqbl.In("status", sqbl.P(1), sqbl.P(2)) // "status" IN ($1, $2)
//	sqbl.In("id", subquery)                 // "id" IN (SELECT ...)
func In(left any, values ...any) syntax.InExpr {
	return syntax.In(left, values...)
}

// NotIn returns a NOT IN expression. With no values it renders TRUE.
//
//	sqbl.NotIn("status", sqbl.Raw("'deleted'"), sqbl.Raw("'banned'"))
//	// "status" NOT IN ('deleted', 'banned')
func NotIn(left any, values ...any) syntax.InExpr {
	return syntax.NotIn(left, values...)
}

// IsNull returns an IS NULL expression.
//
//	sqbl.IsNull("deleted_at") // "deleted_at" IS NULL
func IsNull(col any) syntax.NullExpr {
	return syntax.IsNull(col)
}

// IsNotNull returns an IS NOT NULL expression.
//
//	sqbl.IsNotNull("email") // "email" IS NOT NULL
func IsNotNull(col any) syntax.NullExpr {
	return syntax.IsNotNull(col)
}

// Between returns a BETWEEN ... AND ... expression.
//
//	sqbl.Between("age", 18, 65) // "age" BETWEEN 18 AND 65
func Between(col any, low, high any) syntax.BetweenExpr {
	return syntax.Between(col, low, high)
}

// Like returns a LIKE comparison expression.
// The pattern accepts the same values as Eq; pass it with P, or with Raw for a literal.
//
//	sqbl.Like("name", sqbl.P(1))           // "name" LIKE $1
//	sqbl.Like("name", sqbl.Raw("'%foo%'")) // "name" LIKE '%foo%'
func Like(left any, pattern any) syntax.ComparisonExpr {
	return syntax.Like(left, pattern)
}

// ILike returns an ILIKE comparison expression (case-insensitive LIKE, PostgreSQL-specific).
// The pattern accepts the same values as Eq; pass it with P, or with Raw for a literal.
//
//	sqbl.ILike("name", sqbl.P(1)) // "name" ILIKE $1
func ILike(left any, pattern any) syntax.ComparisonExpr {
	return syntax.ILike(left, pattern)
}

// P creates a bind parameter placeholder. The value itself is passed to the
// database driver as a query argument, never written into the SQL.
//
//	sqbl.P()          // ? in every dialect
//	sqbl.P(1)         // $1 (PostgreSQL), ?1 (SQLite), ? (MySQL)
//	sqbl.P(":status") // :status in every dialect
//
// An index must be an integer of 1 or more, and a name must start with :, @ or $.
// MySQL has no numbered placeholders, so P(n) renders a plain ? there: arguments
// bind in the order the placeholders appear, and an index used twice needs its
// argument passed twice. P takes a placeholder, not a value: never pass user input to P.
func P(args ...any) syntax.Parameter {
	return syntax.P(args...)
}

// Fn creates a SQL function call expression. The name is written as-is, so
// never build it from untrusted input. Each argument accepts the same values as Eq.
//
//	sqbl.Fn("SUM", "amount")                   // SUM("amount")
//	sqbl.Fn("COUNT", "*")                      // COUNT(*)
//	sqbl.Fn("COALESCE", "nickname", sqbl.P(1)) // COALESCE("nickname", $1)
func Fn(name string, args ...any) syntax.FnExpr {
	return syntax.Fn(name, args...)
}

// Cast creates a CAST(expr AS type) expression.
// The type is written as-is, so never build it from untrusted input.
//
//	sqbl.Cast(sqbl.P(1), "BIGINT") // CAST($1 AS BIGINT)
//	sqbl.Cast("price", "TEXT")     // CAST("price" AS TEXT)
func Cast(expr any, typ string) syntax.CastExpr {
	return syntax.Cast(expr, typ)
}

// Over wraps an expression with a window OVER clause.
//
//	sqbl.Over(sqbl.Fn("ROW_NUMBER")).PartitionBy("dept").OrderBy(sqbl.Desc("salary"))
//	// ROW_NUMBER() OVER (PARTITION BY "dept" ORDER BY "salary" DESC)
func Over(expr any) syntax.WindowExpr {
	return syntax.Over(expr)
}
