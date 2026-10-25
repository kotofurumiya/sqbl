package postgres_test

import (
	"context"
	"testing"

	sqbl "github.com/kotofurumiya/sqbl/sqblpg"
)

// TestPostgres_CreateAndDropDatabase verifies the lifecycle of a throwaway database.
func TestPostgres_CreateAndDropDatabase(t *testing.T) {
	t.Parallel()
	conn := connectDB(t)
	ctx := context.Background()
	a := adapt(conn)
	const name = "sqbl_it_throwaway"

	exists := func() bool {
		t.Helper()
		var n int
		if err := conn.QueryRow(ctx, `SELECT COUNT(*) FROM pg_database WHERE datname = $1`, name).Scan(&n); err != nil {
			t.Fatalf("query pg_database: %v", err)
		}
		return n == 1
	}

	if err := a.Exec(ctx, sqbl.DropDatabase(name).IfExists().ToSql()); err != nil {
		t.Fatalf("drop leftover database: %v", err)
	}
	if err := a.Exec(ctx, sqbl.CreateDatabase(name).ToSql()); err != nil {
		t.Fatalf("create database: %v", err)
	}
	if !exists() {
		t.Fatal("database does not exist after CREATE DATABASE")
	}

	if err := a.Exec(ctx, sqbl.DropDatabase(name).IfExists().Force().ToSql()); err != nil {
		t.Fatalf("drop database with force: %v", err)
	}
	if exists() {
		t.Error("database still exists after DROP DATABASE")
	}
}
