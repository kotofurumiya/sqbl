package postgres_test

import (
	"context"
	"testing"

	sqbl "github.com/kotofurumiya/sqbl/sqblpg"
)

// TestPostgres_DistinctOn verifies DISTINCT ON by retrieving each user's
// most recent event status from a status history table.
func TestPostgres_DistinctOn(t *testing.T) {
	t.Parallel()
	conn := connectDB(t)
	defer conn.Close(context.Background())

	ctx := context.Background()
	a := adapt(conn)

	if err := a.Exec(ctx, `
		CREATE TEMPORARY TABLE events (
			id         INTEGER NOT NULL,
			user_id    INTEGER NOT NULL,
			status     TEXT    NOT NULL,
			created_at INTEGER NOT NULL
		)
	`); err != nil {
		t.Fatalf("create table: %v", err)
	}

	// Seed:
	//   user 1 — three events; latest (created_at=3) is 'login'
	//   user 2 — two events;   latest (created_at=2) is 'error'
	if err := a.Exec(ctx,
		defaultSuite.InsertInto("events").
			Columns("id", "user_id", "status", "created_at").
			Values(sqbl.P(1), sqbl.P(2), sqbl.P(3), sqbl.P(4)).
			Values(sqbl.P(5), sqbl.P(6), sqbl.P(7), sqbl.P(8)).
			Values(sqbl.P(9), sqbl.P(10), sqbl.P(11), sqbl.P(12)).
			Values(sqbl.P(13), sqbl.P(14), sqbl.P(15), sqbl.P(16)).
			Values(sqbl.P(17), sqbl.P(18), sqbl.P(19), sqbl.P(20)).
			ToSql(),
		1, 1, "logout", 1,
		2, 1, "login", 2,
		3, 1, "login", 3,
		4, 2, "login", 1,
		5, 2, "error", 2,
	); err != nil {
		t.Fatalf("insert events: %v", err)
	}

	// DISTINCT ON (user_id) picks the first row per user_id after ORDER BY,
	// which is ORDER BY user_id ASC, created_at DESC → latest event per user.
	sql := sqbl.From("events").
		DistinctOn("user_id").
		Select("user_id", "status").
		OrderBy(sqbl.Asc("user_id"), sqbl.Desc("created_at")).
		ToSql()

	type row struct {
		userID int
		status string
	}
	rows, err := a.Query(ctx, sql)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	defer rows.Close()

	var got []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.userID, &r.status); err != nil {
			t.Fatalf("scan: %v", err)
		}
		got = append(got, r)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows err: %v", err)
	}

	want := []row{{1, "login"}, {2, "error"}}
	if len(got) != len(want) {
		t.Fatalf("got %d rows; want %d", len(got), len(want))
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("[%d] got %+v; want %+v", i, got[i], w)
		}
	}
}
