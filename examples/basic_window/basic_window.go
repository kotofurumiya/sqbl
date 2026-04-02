package main

import (
	"fmt"

	sqbl "github.com/kotofurumiya/sqbl/sqblpg"
)

func main() {
	// ROW_NUMBER: assign a sequential number within each partition, ordered by salary
	// SELECT "id", "dept", "salary",
	//   ROW_NUMBER() OVER (PARTITION BY "dept" ORDER BY "salary") AS "row_num"
	// FROM "employees";
	sql1 := sqbl.From("employees").
		Select(
			"id", "dept", "salary",
			sqbl.As(
				sqbl.Over(sqbl.Fn("ROW_NUMBER")).PartitionBy("dept").OrderBy("salary"),
				"row_num",
			),
		).
		ToSql()
	fmt.Println(sql1)

	// RANK vs DENSE_RANK: ties get the same rank; DENSE_RANK leaves no gaps
	// SELECT "id", "dept", "salary",
	//   RANK() OVER (PARTITION BY "dept" ORDER BY "salary") AS "rank",
	//   DENSE_RANK() OVER (PARTITION BY "dept" ORDER BY "salary") AS "dense_rank"
	// FROM "employees";
	window := sqbl.Over(sqbl.Fn("RANK")).PartitionBy("dept").OrderBy("salary")
	denseWindow := sqbl.Over(sqbl.Fn("DENSE_RANK")).PartitionBy("dept").OrderBy("salary")
	sql2 := sqbl.From("employees").
		Select(
			"id", "dept", "salary",
			sqbl.As(window, "rank"),
			sqbl.As(denseWindow, "dense_rank"),
		).
		ToSql()
	fmt.Println(sql2)

	// Running total: SUM() OVER with ORDER BY accumulates the sum row by row
	// SELECT "id", "user_id", "amount",
	//   SUM("amount") OVER (PARTITION BY "user_id" ORDER BY "created_at") AS "running_total"
	// FROM "orders";
	sql3 := sqbl.From("orders").
		Select(
			"id", "user_id", "amount",
			sqbl.As(
				sqbl.Over(sqbl.Fn("SUM", "amount")).PartitionBy("user_id").OrderBy("created_at"),
				"running_total",
			),
		).
		ToSql()
	fmt.Println(sql3)
}
