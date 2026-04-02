package main

import (
	"fmt"

	sqbl "github.com/kotofurumiya/sqbl/sqblpg"
)

func main() {
	// Simple CTE: name a subquery and reference it in the main query
	// WITH "active_users" AS (
	//   SELECT "id", "name" FROM "users" WHERE "active" = TRUE
	// )
	// SELECT "au"."id", "au"."name", COUNT("o"."id") AS "order_count"
	// FROM "active_users" AS "au"
	// LEFT JOIN "orders" AS "o" ON "au"."id" = "o"."user_id"
	// GROUP BY "au"."id", "au"."name";
	activeUsers := sqbl.From("users").
		Select("id", "name").
		Where(sqbl.Eq("active", true))
	sql1 := sqbl.From(sqbl.As("active_users", "au")).
		With("active_users", activeUsers).
		Select("au.id", "au.name", sqbl.As(sqbl.Fn("COUNT", "o.id"), "order_count")).
		LeftJoin(sqbl.As("orders", "o"), sqbl.Eq("au.id", "o.user_id")).
		GroupBy("au.id", "au.name").
		ToSql()
	fmt.Println(sql1)

	// Multiple CTEs chained together
	// WITH
	//   "monthly_sales" AS (
	//     SELECT "user_id", SUM("amount") AS "total" FROM "orders"
	//     WHERE "created_at" >= $1 GROUP BY "user_id"
	//   ),
	//   "top_sellers" AS (
	//     SELECT "user_id" FROM "monthly_sales" WHERE "total" > $2
	//   )
	// SELECT "u"."id", "u"."name"
	// FROM "users" AS "u"
	// INNER JOIN "top_sellers" AS "ts" ON "u"."id" = "ts"."user_id";
	monthlySales := sqbl.From("orders").
		Select("user_id", sqbl.As(sqbl.Fn("SUM", "amount"), "total")).
		Where(sqbl.Gte("created_at", sqbl.P(1))).
		GroupBy("user_id")
	topSellers := sqbl.From("monthly_sales").
		Select("user_id").
		Where(sqbl.Gt("total", sqbl.P(2)))
	sql2 := sqbl.From(sqbl.As("users", "u")).
		With("monthly_sales", monthlySales).
		With("top_sellers", topSellers).
		Select("u.id", "u.name").
		InnerJoin(sqbl.As("top_sellers", "ts"), sqbl.Eq("u.id", "ts.user_id")).
		ToSql()
	fmt.Println(sql2)

	// Recursive CTE: walk an org chart from the root down
	// WITH RECURSIVE "org" AS (
	//   SELECT "id", "name", "manager_id" FROM "employees" WHERE "manager_id" IS NULL
	//   UNION ALL
	//   SELECT "e"."id", "e"."name", "e"."manager_id"
	//   FROM "employees" AS "e"
	//   INNER JOIN "org" ON "e"."manager_id" = "org"."id"
	// )
	// SELECT "id", "name", "manager_id" FROM "org";
	base := sqbl.From("employees").
		Select("id", "name", "manager_id").
		Where(sqbl.IsNull("manager_id"))
	recursive := sqbl.From(sqbl.As("employees", "e")).
		Select("e.id", "e.name", "e.manager_id").
		InnerJoin("org", sqbl.Eq("e.manager_id", "org.id"))
	sql3 := sqbl.From("org").
		WithRecursive("org", base.UnionAll(recursive)).
		Select("id", "name", "manager_id").
		ToSql()
	fmt.Println(sql3)
}
