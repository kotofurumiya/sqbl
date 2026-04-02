package main

import (
	"fmt"

	sqbl "github.com/kotofurumiya/sqbl/sqblpg"
)

func main() {
	// DELETE FROM "users" WHERE "id" = $1;
	sql1 := sqbl.DeleteFrom("users").
		Where(sqbl.Eq("id", sqbl.P(1))).
		ToSql()
	fmt.Println(sql1)

	// Delete with a compound condition
	// DELETE FROM "sessions" WHERE "user_id" = $1 AND "expires_at" < $2;
	sql2 := sqbl.DeleteFrom("sessions").
		Where(sqbl.And(
			sqbl.Eq("user_id", sqbl.P(1)),
			sqbl.Lt("expires_at", sqbl.P(2)),
		)).
		ToSql()
	fmt.Println(sql2)

	// DELETE ... RETURNING to get the deleted row
	// DELETE FROM "users" WHERE "id" = $1 RETURNING "id", "email";
	sql3 := sqbl.DeleteFrom("users").
		Where(sqbl.Eq("id", sqbl.P(1))).
		Returning("id", "email").
		ToSql()
	fmt.Println(sql3)
}
