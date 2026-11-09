package main

import (
	"fmt"

	sqbl "github.com/kotofurumiya/sqbl/sqblpg"
)

func main() {
	// Indexed parameters: $1, $2, ... (PostgreSQL style)
	// SELECT "id", "name" FROM "users" WHERE "id" = $1;
	sql1 := sqbl.From("users").
		Select("id", "name").
		Where(sqbl.Eq("id", sqbl.P(1))).
		ToSQL()
	fmt.Println(sql1)

	// Use indexed params in INSERT
	// INSERT INTO "users" ("name", "email") VALUES ($1, $2);
	sql2 := sqbl.InsertInto("users").
		Columns("name", "email").
		Values(sqbl.P(1), sqbl.P(2)).
		ToSQL()
	fmt.Println(sql2)

	// Named parameters: :name style
	// SELECT "id", "name" FROM "users" WHERE "status" = :status AND "role" = :role;
	sql3 := sqbl.From("users").
		Select("id", "name").
		Where(sqbl.And(
			sqbl.Eq("status", sqbl.P(":status")),
			sqbl.Eq("role", sqbl.P(":role")),
		)).
		ToSQL()
	fmt.Println(sql3)

	// Positional parameter: ? style (useful for MySQL or generic drivers)
	// SELECT "id", "name" FROM "users" WHERE "id" = ?;
	sql4 := sqbl.From("users").
		Select("id", "name").
		Where(sqbl.Eq("id", sqbl.P())).
		ToSQL()
	fmt.Println(sql4)
}
