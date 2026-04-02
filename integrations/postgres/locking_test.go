package postgres_test

import (
	"context"
	"testing"

	"github.com/kotofurumiya/sqbl/sqblpg"
)

// TestPostgres_ForUpdateSkipLocked verifies the job-queue dequeue pattern:
// two concurrent workers each select the next available job using
// FOR UPDATE SKIP LOCKED, and must not receive the same row.
//
// This test requires a shared (non-temporary) table because TEMP TABLEs are
// session-scoped and invisible to other connections. The table is created with
// IF NOT EXISTS, truncated before seeding, and dropped in t.Cleanup.
func TestPostgres_ForUpdateSkipLocked(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	conn := connectDB(t)
	defer conn.Close(ctx)

	// A second independent connection simulates a concurrent worker process.
	conn2 := connectDB(t)
	defer conn2.Close(ctx)

	// Create the shared table if it does not exist yet, then wipe any leftover
	// rows from a previous run so the test starts with a clean slate.
	if _, err := conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS jobs_test (
			id      SERIAL PRIMARY KEY,
			payload TEXT NOT NULL,
			status  TEXT NOT NULL DEFAULT 'pending'
		)
	`); err != nil {
		t.Fatalf("create table: %v", err)
	}

	if _, err := conn.Exec(ctx, `TRUNCATE jobs_test`); err != nil {
		t.Fatalf("truncate: %v", err)
	}

	// Drop the table once the test finishes so it does not leak into other runs.
	t.Cleanup(func() {
		conn.Exec(ctx, `DROP TABLE IF EXISTS jobs_test`)
	})

	// Insert three pending jobs. Having at least two rows is essential:
	// worker1 will lock the first row, so worker2 must be able to skip it and
	// pick up a different one.
	seedSQL := sqblpg.InsertInto("jobs_test").
		Columns("payload").
		Values(sqblpg.P(1)).
		Values(sqblpg.P(2)).
		Values(sqblpg.P(3)).
		ToSql()
	if _, err := conn.Exec(ctx, seedSQL, "job-A", "job-B", "job-C"); err != nil {
		t.Fatalf("insert jobs: %v", err)
	}

	// Build the dequeue query once and reuse it for both workers.
	// FOR UPDATE acquires a row-level lock; SKIP LOCKED bypasses already-locked
	// rows instead of waiting, which is the standard pattern for job queues.
	dequeueSQL := sqblpg.From("jobs_test").
		Select("id", "payload").
		Where(sqblpg.Eq("status", sqblpg.P(1))).
		OrderBy(sqblpg.Asc("id")).
		Limit(1).
		ForUpdate().SkipLocked().
		ToSql()

	// worker1 opens a transaction and locks the first pending job without
	// committing, so the lock remains active when worker2 runs.
	tx1, err := conn.Begin(ctx)
	if err != nil {
		t.Fatalf("begin tx1: %v", err)
	}
	defer tx1.Rollback(ctx)

	var id1 int
	var payload1 string
	if err := tx1.QueryRow(ctx, dequeueSQL, "pending").Scan(&id1, &payload1); err != nil {
		t.Fatalf("dequeue tx1: %v", err)
	}

	// worker2 runs the same query on a separate connection. Because id1 is
	// still locked by tx1, SKIP LOCKED causes PostgreSQL to skip that row and
	// return the next available one instead.
	tx2, err := conn2.Begin(ctx)
	if err != nil {
		t.Fatalf("begin tx2: %v", err)
	}
	defer tx2.Rollback(ctx)

	var id2 int
	var payload2 string
	if err := tx2.QueryRow(ctx, dequeueSQL, "pending").Scan(&id2, &payload2); err != nil {
		t.Fatalf("dequeue tx2: %v", err)
	}

	if id1 == id2 {
		t.Errorf("SKIP LOCKED failed: both workers got same job id=%d", id1)
	}
	t.Logf("worker1 got id=%d (%s), worker2 got id=%d (%s)", id1, payload1, id2, payload2)
}
