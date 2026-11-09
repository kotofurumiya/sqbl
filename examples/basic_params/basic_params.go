package main

import (
	"fmt"

	sqbl "github.com/kotofurumiya/sqbl/sqblpg"
)

func main() {
	// Indexed parameters: $1, $2, ... — the form pgx and lib/pq expect.
	// SELECT "id", "name" FROM "users" WHERE "id" = $1;
	sql1 := sqbl.From("users").
		Select("id", "name").
		Where(sqbl.Eq("id", sqbl.P(1))).
		MustSQL()
	fmt.Println(sql1)

	// Use indexed params in INSERT
	// INSERT INTO "users" ("name", "email") VALUES ($1, $2);
	sql2 := sqbl.InsertInto("users").
		Columns("name", "email").
		Values(sqbl.P(1), sqbl.P(2)).
		MustSQL()
	fmt.Println(sql2)

	// Named parameters: :name style, written as-is.
	// Only for drivers or tools that bind by name; pgx does not.
	// SELECT "id", "name" FROM "users" WHERE "status" = :status AND "role" = :role;
	sql3 := sqbl.From("users").
		Select("id", "name").
		Where(sqbl.And(
			sqbl.Eq("status", sqbl.P(":status")),
			sqbl.Eq("role", sqbl.P(":role")),
		)).
		MustSQL()
	fmt.Println(sql3)

	// Positional parameter: ? style.
	// Only for drivers that use ? placeholders, such as MySQL and SQLite drivers.
	// pgx rejects ?, so with PostgreSQL use indexed parameters instead.
	// SELECT "id", "name" FROM "users" WHERE "id" = ?;
	sql4 := sqbl.From("users").
		Select("id", "name").
		Where(sqbl.Eq("id", sqbl.P())).
		MustSQL()
	fmt.Println(sql4)
}
