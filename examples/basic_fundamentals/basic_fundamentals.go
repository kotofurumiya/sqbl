// Package main is a quick-reference for SQL fundamentals using sqbl.
//
// Tables used throughout:
//
//	employees  (id BIGINT, name TEXT, dept_id BIGINT, salary NUMERIC)
//	departments(id BIGINT, name TEXT)
package main

import (
	"fmt"

	sqbl "github.com/kotofurumiya/sqbl/sqblpg"
)

func main() {
	// ---------------------------------------------------------------
	// 1. SELECT * — retrieve every column
	// ---------------------------------------------------------------
	// SELECT * FROM "employees";
	sql01 := sqbl.From("employees").ToSql()
	fmt.Println(sql01)

	// ---------------------------------------------------------------
	// 2. WHERE — filter rows
	// ---------------------------------------------------------------
	// SELECT * FROM "employees" WHERE "salary" > 50000;
	sql02 := sqbl.From("employees").
		Where(sqbl.Gt("salary", 50000)).
		ToSql()
	fmt.Println(sql02)

	// ---------------------------------------------------------------
	// 3. SELECT columns + ORDER BY + LIMIT
	// ---------------------------------------------------------------
	// SELECT "name", "salary" FROM "employees" ORDER BY "salary" DESC LIMIT 5;
	sql03 := sqbl.From("employees").
		Select("name", "salary").
		OrderBy(sqbl.Desc("salary")).
		Limit(5).
		ToSql()
	fmt.Println(sql03)

	// ---------------------------------------------------------------
	// 4. DISTINCT — deduplicate results
	// ---------------------------------------------------------------
	// SELECT DISTINCT "dept_id" FROM "employees";
	sql04 := sqbl.From("employees").
		Select("dept_id").
		Distinct().
		ToSql()
	fmt.Println(sql04)

	// ---------------------------------------------------------------
	// 5. Aggregate functions — COUNT, SUM, AVG (no grouping)
	// ---------------------------------------------------------------
	// SELECT COUNT(*), SUM("salary"), AVG("salary") FROM "employees";
	sql05 := sqbl.From("employees").
		Select(
			sqbl.Fn("COUNT", "*"),
			sqbl.Fn("SUM", "salary"),
			sqbl.Fn("AVG", "salary"),
		).
		ToSql()
	fmt.Println(sql05)

	// ---------------------------------------------------------------
	// 6. GROUP BY — aggregate per group
	// ---------------------------------------------------------------
	// SELECT "dept_id", COUNT(*) AS "headcount", SUM("salary") AS "total_salary"
	// FROM "employees"
	// GROUP BY "dept_id"
	// ORDER BY "headcount" DESC;
	sql06 := sqbl.From("employees").
		Select(
			"dept_id",
			sqbl.As(sqbl.Fn("COUNT", "*"), "headcount"),
			sqbl.As(sqbl.Fn("SUM", "salary"), "total_salary"),
		).
		GroupBy("dept_id").
		OrderBy(sqbl.Desc("headcount")).
		ToSql()
	fmt.Println(sql06)

	// ---------------------------------------------------------------
	// 7. INNER JOIN — combine matching rows from two tables
	// ---------------------------------------------------------------
	// SELECT "e"."name", "d"."name" AS "dept"
	// FROM "employees" AS "e"
	// INNER JOIN "departments" AS "d" ON "e"."dept_id" = "d"."id";
	sql07 := sqbl.From(sqbl.As("employees", "e")).
		Select("e.name", sqbl.As("d.name", "dept")).
		InnerJoin(sqbl.As("departments", "d"), sqbl.Eq("e.dept_id", "d.id")).
		ToSql()
	fmt.Println(sql07)

	// ---------------------------------------------------------------
	// 8. LEFT JOIN — keep all left-side rows, NULL when no match
	// ---------------------------------------------------------------
	// SELECT "e"."name", "d"."name" AS "dept"
	// FROM "employees" AS "e"
	// LEFT JOIN "departments" AS "d" ON "e"."dept_id" = "d"."id";
	sql08 := sqbl.From(sqbl.As("employees", "e")).
		Select("e.name", sqbl.As("d.name", "dept")).
		LeftJoin(sqbl.As("departments", "d"), sqbl.Eq("e.dept_id", "d.id")).
		ToSql()
	fmt.Println(sql08)

	// ---------------------------------------------------------------
	// 9. Subquery in WHERE — e.g. employees earning above average
	// ---------------------------------------------------------------
	// SELECT "name", "salary" FROM "employees"
	// WHERE "salary" > (SELECT AVG("salary") FROM "employees");
	avgSalary := sqbl.From("employees").Select(sqbl.Fn("AVG", "salary"))
	sql09 := sqbl.From("employees").
		Select("name", "salary").
		Where(sqbl.Gt("salary", avgSalary)).
		ToSql()
	fmt.Println(sql09)

	// ---------------------------------------------------------------
	// 10. Parameters — safe placeholders instead of string literals
	// ---------------------------------------------------------------

	// Indexed ($1, $2, ...) — standard for PostgreSQL
	// SELECT "name", "salary" FROM "employees" WHERE "dept_id" = $1 AND "salary" > $2;
	sql10a := sqbl.From("employees").
		Select("name", "salary").
		Where(sqbl.And(
			sqbl.Eq("dept_id", sqbl.P(1)),
			sqbl.Gt("salary", sqbl.P(2)),
		)).
		ToSql()
	fmt.Println(sql10a)

	// Named (:param) — useful for named-parameter drivers
	// SELECT "name", "salary" FROM "employees" WHERE "dept_id" = :dept_id AND "salary" > :min_salary;
	sql10b := sqbl.From("employees").
		Select("name", "salary").
		Where(sqbl.And(
			sqbl.Eq("dept_id", sqbl.P(":dept_id")),
			sqbl.Gt("salary", sqbl.P(":min_salary")),
		)).
		ToSql()
	fmt.Println(sql10b)

	// Positional (?) — for drivers that use positional ? placeholders (e.g. MySQL, SQLite)
	// SELECT "name", "salary" FROM "employees" WHERE "dept_id" = ? AND "salary" > ?;
	sql10c := sqbl.From("employees").
		Select("name", "salary").
		Where(sqbl.And(
			sqbl.Eq("dept_id", sqbl.P()),
			sqbl.Gt("salary", sqbl.P()),
		)).
		ToSql()
	fmt.Println(sql10c)
}
