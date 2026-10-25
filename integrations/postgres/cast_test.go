package postgres_test

import (
	"context"
	"slices"
	"testing"

	sqbl "github.com/kotofurumiya/sqbl/sqblpg"
)

// TestPostgres_UnnestCastWithColumnAlias verifies the typical way of passing
// a list of IDs as one array parameter and joining it as a table.
// Without the CAST, PostgreSQL cannot infer the parameter type for unnest.
func TestPostgres_UnnestCastWithColumnAlias(t *testing.T) {
	t.Parallel()
	conn := connectDB(t)
	ctx := context.Background()
	a := adapt(conn)

	if err := a.Exec(ctx, `
		CREATE TEMPORARY TABLE users (
			id   BIGINT NOT NULL PRIMARY KEY,
			name TEXT   NOT NULL
		)
	`); err != nil {
		t.Fatalf("create table: %v", err)
	}
	if err := a.Exec(ctx,
		defaultSuite.InsertInto("users").
			Columns("id", "name").
			Values(sqbl.P(1), sqbl.P(2)).
			Values(sqbl.P(3), sqbl.P(4)).
			Values(sqbl.P(5), sqbl.P(6)).
			ToSql(),
		1, "alice",
		2, "bob",
		3, "carol",
	); err != nil {
		t.Fatalf("insert users: %v", err)
	}

	ids := sqbl.AsColumns(sqbl.Fn("unnest", sqbl.Cast(sqbl.P(1), "bigint[]")), "ids", "id")
	sql := sqbl.From("users").
		Join(ids, sqbl.Eq("users.id", "ids.id")).
		Select("users.name").
		OrderBy(sqbl.Asc("users.name")).
		ToSql()

	rows, err := a.Query(ctx, sql, []int64{1, 3})
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	defer rows.Close()

	var got []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("scan: %v", err)
		}
		got = append(got, name)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows err: %v", err)
	}

	want := []string{"alice", "carol"}
	if !slices.Equal(got, want) {
		t.Errorf("got %v; want %v", got, want)
	}
}
