// Package main demonstrates an audit log pattern using sqbl.
//
// A common requirement is recording every change to sensitive data.
// This example shows how to build the queries for:
//  1. Upserting a user record (INSERT ... ON CONFLICT DO UPDATE)
//  2. Writing an audit log entry and retrieving its generated id via RETURNING
//  3. Querying the audit trail for a specific resource
package main

import (
	"fmt"

	sqbl "github.com/kotofurumiya/sqbl/sqblpg"
)

func main() {
	// Step 1: Upsert a user — insert or update on email conflict.
	// INSERT INTO "users" ("email", "name", "updated_at")
	// VALUES ($1, $2, NOW())
	// ON CONFLICT ("email") DO UPDATE SET
	//   "name" = EXCLUDED."name", "updated_at" = NOW()
	// RETURNING "id", "email";
	upsertUser := sqbl.InsertInto("users").
		Columns("email", "name", "updated_at").
		Values(sqbl.P(1), sqbl.P(2), "NOW()").
		OnConflict("email").DoUpdate(
		`"name" = EXCLUDED."name"`,
		`"updated_at" = NOW()`,
	).
		Returning("id", "email").
		ToSql()
	fmt.Println(upsertUser)

	// Step 2: Write an audit log entry for the change.
	// The caller passes the user id obtained from step 1.
	//
	// INSERT INTO "audit_logs" ("user_id", "action", "payload", "created_at")
	// VALUES ($1, $2, $3, NOW())
	// RETURNING "id";
	writeAuditLog := sqbl.InsertInto("audit_logs").
		Columns("user_id", "action", "payload", "created_at").
		Values(sqbl.P(1), sqbl.P(2), sqbl.P(3), "NOW()").
		Returning("id").
		ToSql()
	fmt.Println(writeAuditLog)

	// Step 3: Fetch the full audit trail for a user, newest first.
	//
	// SELECT "id", "action", "payload", "created_at"
	// FROM "audit_logs"
	// WHERE "user_id" = $1
	// ORDER BY "created_at" DESC
	// LIMIT 50;
	auditTrail := sqbl.From("audit_logs").
		Select("id", "action", "payload", "created_at").
		Where(sqbl.Eq("user_id", sqbl.P(1))).
		OrderBy(sqbl.Desc("created_at")).
		Limit(50).
		ToSql()
	fmt.Println(auditTrail)
}
