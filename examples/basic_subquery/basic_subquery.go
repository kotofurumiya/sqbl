package main

import (
	"fmt"

	sqbl "github.com/kotofurumiya/sqbl/sqblpg"
)

func main() {
	// Subquery in FROM
	// SELECT "sub"."user_id", "sub"."total"
	// FROM (
	//   SELECT "user_id", SUM("amount") AS "total" FROM "orders" GROUP BY "user_id"
	// ) AS "sub"
	// WHERE "sub"."total" > $1;
	sub := sqbl.From("orders").
		Select("user_id", sqbl.As(sqbl.Fn("SUM", "amount"), "total")).
		GroupBy("user_id")
	sql1 := sqbl.From(sqbl.As(sub, "sub")).
		Select("sub.user_id", "sub.total").
		Where(sqbl.Gt("sub.total", sqbl.P(1))).
		ToSql()
	fmt.Println(sql1)

	// Subquery in WHERE with IN
	// SELECT "id", "name" FROM "users"
	// WHERE "id" IN (SELECT "user_id" FROM "orders" WHERE "status" = 'paid');
	paidUsersSub := sqbl.From("orders").
		Select("user_id").
		Where(sqbl.Eq("status", "'paid'"))
	sql2 := sqbl.From("users").
		Select("id", "name").
		Where(sqbl.In("id", paidUsersSub)).
		ToSql()
	fmt.Println(sql2)

	// UNION: combine results from two queries
	// SELECT "id", "name" FROM "users" WHERE "role" = 'admin'
	// UNION
	// SELECT "id", "name" FROM "users" WHERE "created_at" > $1;
	admins := sqbl.From("users").Select("id", "name").Where(sqbl.Eq("role", "'admin'"))
	newUsers := sqbl.From("users").Select("id", "name").Where(sqbl.Gt("created_at", sqbl.P(1)))
	sql3 := admins.Union(newUsers).ToSql()
	fmt.Println(sql3)
}
