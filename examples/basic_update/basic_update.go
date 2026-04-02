package main

import (
	"fmt"

	sqbl "github.com/kotofurumiya/sqbl/sqblpg"
)

func main() {
	// UPDATE "users" SET "name" = $1 WHERE "id" = $2;
	sql1 := sqbl.Update("users").
		Set("name", sqbl.P(1)).
		Where(sqbl.Eq("id", sqbl.P(2))).
		ToSql()
	fmt.Println(sql1)

	// Update multiple columns at once
	// UPDATE "users" SET "name" = $1, "email" = $2 WHERE "id" = $3;
	sql2 := sqbl.Update("users").
		Set("name", sqbl.P(1)).
		Set("email", sqbl.P(2)).
		Where(sqbl.Eq("id", sqbl.P(3))).
		ToSql()
	fmt.Println(sql2)

	// UPDATE ... RETURNING to get the updated row
	// UPDATE "users" SET "name" = $1 WHERE "id" = $2 RETURNING "id", "name", "updated_at";
	sql3 := sqbl.Update("users").
		Set("name", sqbl.P(1)).
		Where(sqbl.Eq("id", sqbl.P(2))).
		Returning("id", "name", "updated_at").
		ToSql()
	fmt.Println(sql3)
}
