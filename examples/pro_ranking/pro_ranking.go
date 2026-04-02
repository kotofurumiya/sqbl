// Package main demonstrates a sales ranking report using sqbl.
//
// A typical analytics query: aggregate monthly sales per sales rep,
// rank them within their region, and return only the top 3 per region.
// Uses CTEs to separate the aggregation and ranking steps for readability.
package main

import (
	"fmt"

	sqbl "github.com/kotofurumiya/sqbl/sqblpg"
)

func main() {
	// Step 1 (CTE): Aggregate total sales per rep for the target month.
	//
	// monthly_totals AS (
	//   SELECT "rep_id", "region", SUM("amount") AS "total"
	//   FROM "sales"
	//   WHERE "sold_at" >= $1 AND "sold_at" < $2
	//   GROUP BY "rep_id", "region"
	// )
	monthlyTotals := sqbl.From("sales").
		Select(
			"rep_id",
			"region",
			sqbl.As(sqbl.Fn("SUM", "amount"), "total"),
		).
		Where(sqbl.And(
			sqbl.Gte("sold_at", sqbl.P(1)),
			sqbl.Lt("sold_at", sqbl.P(2)),
		)).
		GroupBy("rep_id", "region")

	// Step 2 (CTE): Rank reps within each region by their monthly total.
	//
	// ranked AS (
	//   SELECT "rep_id", "region", "total",
	//     RANK() OVER (PARTITION BY "region" ORDER BY "total") AS "rank"
	//   FROM "monthly_totals"
	// )
	ranked := sqbl.From("monthly_totals").
		Select(
			"rep_id",
			"region",
			"total",
			sqbl.As(
				sqbl.Over(sqbl.Fn("RANK")).PartitionBy("region").OrderBy("total"),
				"rank",
			),
		)

	// Step 3 (main query): Join with the reps table and keep only the top 3 per region.
	//
	// SELECT "r"."name", "rk"."region", "rk"."total", "rk"."rank"
	// FROM "ranked" AS "rk"
	// INNER JOIN "reps" AS "r" ON "rk"."rep_id" = "r"."id"
	// WHERE "rk"."rank" <= 3
	// ORDER BY "rk"."region" ASC, "rk"."rank" ASC;
	sql := sqbl.From(sqbl.As("ranked", "rk")).
		With("monthly_totals", monthlyTotals).
		With("ranked", ranked).
		Select("r.name", "rk.region", "rk.total", "rk.rank").
		InnerJoin(sqbl.As("reps", "r"), sqbl.Eq("rk.rep_id", "r.id")).
		Where(sqbl.Lte("rk.rank", 3)).
		OrderBy(sqbl.Asc("rk.region"), sqbl.Asc("rk.rank")).
		ToSql()
	fmt.Println(sql)
}
