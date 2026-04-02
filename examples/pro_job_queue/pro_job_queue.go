// Package main demonstrates a background job queue pattern using sqbl.
//
// A common pattern for reliable job processing is to atomically claim one
// pending job with SELECT FOR UPDATE SKIP LOCKED, then update its status.
// SKIP LOCKED lets multiple workers run concurrently without stepping on
// each other — rows already locked by another worker are silently skipped.
package main

import (
	"fmt"

	sqbl "github.com/kotofurumiya/sqbl/sqblpg"
)

func main() {
	// Step 1: Claim the next available job.
	// Selects the oldest pending job, locks it, and skips any rows already
	// locked by another worker. Run inside a transaction.
	//
	// SELECT "id", "type", "payload"
	// FROM "jobs"
	// WHERE "status" = 'pending'
	// ORDER BY "id" ASC
	// LIMIT 1
	// FOR UPDATE SKIP LOCKED;
	claimJob := sqbl.From("jobs").
		Select("id", "type", "payload").
		Where(sqbl.Eq("status", "'pending'")).
		OrderBy(sqbl.Asc("id")).
		Limit(1).
		ForUpdate().SkipLocked().
		ToSql()
	fmt.Println(claimJob)

	// Step 2: Mark the claimed job as processing.
	// UPDATE "jobs" SET "status" = 'processing', "started_at" = NOW()
	// WHERE "id" = $1;
	startJob := sqbl.Update("jobs").
		Set("status", "'processing'").
		Set("started_at", "NOW()").
		Where(sqbl.Eq("id", sqbl.P(1))).
		ToSql()
	fmt.Println(startJob)

	// Step 3a: Mark the job as done.
	// UPDATE "jobs" SET "status" = 'done', "finished_at" = NOW()
	// WHERE "id" = $1
	// RETURNING "id", "type";
	completeJob := sqbl.Update("jobs").
		Set("status", "'done'").
		Set("finished_at", "NOW()").
		Where(sqbl.Eq("id", sqbl.P(1))).
		Returning("id", "type").
		ToSql()
	fmt.Println(completeJob)

	// Step 3b: Mark the job as failed and record the error message.
	// UPDATE "jobs" SET "status" = 'failed', "error" = $1, "finished_at" = NOW()
	// WHERE "id" = $2
	// RETURNING "id", "type";
	failJob := sqbl.Update("jobs").
		Set("status", "'failed'").
		Set("error", sqbl.P(1)).
		Set("finished_at", "NOW()").
		Where(sqbl.Eq("id", sqbl.P(2))).
		Returning("id", "type").
		ToSql()
	fmt.Println(failJob)
}
