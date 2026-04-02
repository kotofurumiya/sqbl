package main

import (
	"fmt"

	sqbl "github.com/kotofurumiya/sqbl/sqblpg"
)

func main() {
	// INSERT INTO "users" ("name", "email") VALUES ($1, $2);
	sql1 := sqbl.InsertInto("users").
		Columns("name", "email").
		Values(sqbl.P(1), sqbl.P(2)).
		ToSql()
	fmt.Println(sql1)

	// Multiple rows in a single INSERT
	// INSERT INTO "users" ("name", "email") VALUES ($1, $2), ($3, $4);
	sql2 := sqbl.InsertInto("users").
		Columns("name", "email").
		Values(sqbl.P(1), sqbl.P(2)).
		Values(sqbl.P(3), sqbl.P(4)).
		ToSql()
	fmt.Println(sql2)

	// INSERT ... RETURNING to get the generated id
	// INSERT INTO "users" ("name", "email") VALUES ($1, $2) RETURNING "id", "created_at";
	sql3 := sqbl.InsertInto("users").
		Columns("name", "email").
		Values(sqbl.P(1), sqbl.P(2)).
		Returning("id", "created_at").
		ToSql()
	fmt.Println(sql3)

	// Upsert: ON CONFLICT DO NOTHING
	// INSERT INTO "tags" ("name") VALUES ($1) ON CONFLICT DO NOTHING;
	sql4 := sqbl.InsertInto("tags").
		Columns("name").
		Values(sqbl.P(1)).
		OnConflictDoNothing().
		ToSql()
	fmt.Println(sql4)

	// Upsert: ON CONFLICT DO UPDATE (update on duplicate key)
	// INSERT INTO "settings" ("user_id", "key", "value") VALUES ($1, $2, $3)
	// ON CONFLICT ("user_id", "key") DO UPDATE SET "value" = EXCLUDED."value";
	sql5 := sqbl.InsertInto("settings").
		Columns("user_id", "key", "value").
		Values(sqbl.P(1), sqbl.P(2), sqbl.P(3)).
		OnConflict("user_id", "key").DoUpdate(`"value" = EXCLUDED."value"`).
		ToSql()
	fmt.Println(sql5)
}
