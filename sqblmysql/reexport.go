package sqblmysql

import "github.com/kotofurumiya/sqbl/syntax"

// Ident is an identifier such as a column or table name. It is quoted, and
// each dot-separated part is quoted separately. A plain string passed where an
// expression is expected is treated the same way.
//
//	sqblmysql.Ident("users.id") // `users`.`id`
type Ident = syntax.Ident

// Raw is a SQL fragment written as-is, without quoting or escaping.
// Never build it from untrusted input; pass values with P instead.
//
//	sqblmysql.Raw("CURRENT_TIMESTAMP")
//	sqblmysql.Raw("'pending'")
type Raw = syntax.Raw

// As creates an aliased expression for use in Select(), From(), or Join().
// The source accepts the same values as Eq; the alias is quoted as a single identifier.
//
//	sqblmysql.As("users", "u")                           // `users` AS `u`
//	sqblmysql.As(sqblmysql.Fn("SUM", "amount"), "total") // SUM(`amount`) AS `total`
//	sqblmysql.As(subquery, "sub")                        // (SELECT ...) AS `sub`
func As(source any, alias string) syntax.Aliased {
	return syntax.As(source, alias)
}

// AsColumns creates a table alias with column aliases, for naming the columns
// of a function call, VALUES list, or subquery used in From() or Join().
//
//	sqblmysql.AsColumns(subquery, "t", "a", "b") // (SELECT ...) AS `t` (`a`, `b`)
func AsColumns(source any, alias string, columns ...string) syntax.Aliased {
	return syntax.AsColumns(source, alias, columns...)
}

// Eq returns an equality (=) comparison expression.
// A plain string on either side is an identifier, so compare against a value
// with P, or with Raw for a literal written into the SQL.
// A nil operand fails to render; use IsNull instead.
//
//	sqblmysql.Eq("status", sqblmysql.P())         // `status` = ?
//	sqblmysql.Eq("u.id", "o.user_id")             // `u`.`id` = `o`.`user_id`
//	sqblmysql.Eq("status", sqblmysql.Raw("'ok'")) // `status` = 'ok'
func Eq(left any, right any) syntax.ComparisonExpr {
	return syntax.Eq(left, right)
}

// Ne returns an inequality (<>) comparison expression.
//
//	sqblmysql.Ne("status", sqblmysql.P()) // `status` <> ?
func Ne(left any, right any) syntax.ComparisonExpr {
	return syntax.Ne(left, right)
}

// Lt returns a less-than (<) comparison expression.
//
//	sqblmysql.Lt("age", 18) // `age` < 18
func Lt(left any, right any) syntax.ComparisonExpr {
	return syntax.Lt(left, right)
}

// Lte returns a less-than-or-equal (<=) comparison expression.
//
//	sqblmysql.Lte("age", 18) // `age` <= 18
func Lte(left any, right any) syntax.ComparisonExpr {
	return syntax.Lte(left, right)
}

// Gt returns a greater-than (>) comparison expression.
//
//	sqblmysql.Gt("score", 100) // `score` > 100
func Gt(left any, right any) syntax.ComparisonExpr {
	return syntax.Gt(left, right)
}

// Gte returns a greater-than-or-equal (>=) comparison expression.
//
//	sqblmysql.Gte("score", 100) // `score` >= 100
func Gte(left any, right any) syntax.ComparisonExpr {
	return syntax.Gte(left, right)
}

// And combines conditions with AND. With no conditions it renders TRUE,
// so a filter list built at run time can be empty.
//
//	sqblmysql.And(sqblmysql.Eq("active", true), sqblmysql.Gt("age", 18))
//	// `active` = TRUE AND `age` > 18
func And(exprs ...syntax.SQLFragment) syntax.LogicalExpr {
	return syntax.And(exprs...)
}

// Or combines conditions with OR. With no conditions it renders FALSE,
// so a filter list built at run time can be empty.
//
//	sqblmysql.Or(sqblmysql.Eq("role", sqblmysql.P()), sqblmysql.Eq("role", sqblmysql.P()))
//	// `role` = ? OR `role` = ?
func Or(exprs ...syntax.SQLFragment) syntax.LogicalExpr {
	return syntax.Or(exprs...)
}

// Not negates a condition.
//
//	sqblmysql.Not(sqblmysql.Eq("active", false)) // NOT (`active` = FALSE)
func Not(expr syntax.SQLFragment) syntax.NotExpr {
	return syntax.Not(expr)
}

// Asc creates an ascending ORDER BY expression.
// The expression accepts the same values as Eq, so a plain string is an identifier.
//
//	sqblmysql.Asc("name")                        // `name` ASC
//	sqblmysql.Asc(sqblmysql.Fn("LOWER", "name")) // LOWER(`name`) ASC
func Asc(expr any) syntax.Order {
	return syntax.Asc(expr)
}

// Desc creates a descending ORDER BY expression.
// The expression accepts the same values as Eq, so a plain string is an identifier.
//
//	sqblmysql.Desc("created_at") // `created_at` DESC
func Desc(expr any) syntax.Order {
	return syntax.Desc(expr)
}

// In returns an IN expression. A lone subquery is used as the list itself.
// With no values it renders FALSE, so a list built at run time can be empty.
//
//	sqblmysql.In("status", sqblmysql.P(), sqblmysql.P()) // `status` IN (?, ?)
//	sqblmysql.In("id", subquery)                         // `id` IN (SELECT ...)
func In(left any, values ...any) syntax.InExpr {
	return syntax.In(left, values...)
}

// NotIn returns a NOT IN expression. With no values it renders TRUE.
//
//	sqblmysql.NotIn("status", sqblmysql.Raw("'deleted'"), sqblmysql.Raw("'banned'"))
//	// `status` NOT IN ('deleted', 'banned')
func NotIn(left any, values ...any) syntax.InExpr {
	return syntax.NotIn(left, values...)
}

// IsNull returns an IS NULL expression.
//
//	sqblmysql.IsNull("deleted_at") // `deleted_at` IS NULL
func IsNull(col any) syntax.NullExpr {
	return syntax.IsNull(col)
}

// IsNotNull returns an IS NOT NULL expression.
//
//	sqblmysql.IsNotNull("email") // `email` IS NOT NULL
func IsNotNull(col any) syntax.NullExpr {
	return syntax.IsNotNull(col)
}

// Between returns a BETWEEN ... AND ... expression.
//
//	sqblmysql.Between("age", 18, 65) // `age` BETWEEN 18 AND 65
func Between(col any, low, high any) syntax.BetweenExpr {
	return syntax.Between(col, low, high)
}

// Like returns a LIKE comparison expression.
// The pattern accepts the same values as Eq; pass it with P, or with Raw for a literal.
//
//	sqblmysql.Like("name", sqblmysql.P())            // `name` LIKE ?
//	sqblmysql.Like("name", sqblmysql.Raw("'%foo%'")) // `name` LIKE '%foo%'
func Like(left any, pattern any) syntax.ComparisonExpr {
	return syntax.Like(left, pattern)
}

// P creates a bind parameter placeholder. The value itself is passed to the
// database driver as a query argument, never written into the SQL.
//
//	sqblmysql.P()  // ?
//	sqblmysql.P(1) // ? as well: MySQL has no numbered placeholders
//
// Prefer P(). The number in P(n) is ignored: arguments bind in the order the
// placeholders appear in the SQL, and an index used twice needs its argument
// passed twice. Named placeholders such as P(":status") are written as-is.
// P takes a placeholder, not a value: never pass user input to P.
func P(args ...any) syntax.Parameter {
	return syntax.P(args...)
}

// Fn creates a SQL function call expression. The name is written as-is, so
// never build it from untrusted input. Each argument accepts the same values as Eq.
//
//	sqblmysql.Fn("SUM", "amount")                       // SUM(`amount`)
//	sqblmysql.Fn("COUNT", "*")                          // COUNT(*)
//	sqblmysql.Fn("COALESCE", "nickname", sqblmysql.P()) // COALESCE(`nickname`, ?)
func Fn(name string, args ...any) syntax.FnExpr {
	return syntax.Fn(name, args...)
}

// Cast creates a CAST(expr AS type) expression.
// The type is written as-is, so never build it from untrusted input.
//
//	sqblmysql.Cast(sqblmysql.P(), "SIGNED") // CAST(? AS SIGNED)
//	sqblmysql.Cast("price", "CHAR")         // CAST(`price` AS CHAR)
func Cast(expr any, typ string) syntax.CastExpr {
	return syntax.Cast(expr, typ)
}

// Over wraps an expression with a window OVER clause.
//
//	sqblmysql.Over(sqblmysql.Fn("ROW_NUMBER")).PartitionBy("dept").OrderBy(sqblmysql.Desc("salary"))
//	// ROW_NUMBER() OVER (PARTITION BY `dept` ORDER BY `salary` DESC)
func Over(expr any) syntax.WindowExpr {
	return syntax.Over(expr)
}
