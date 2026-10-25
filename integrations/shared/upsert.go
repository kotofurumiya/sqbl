package shared

import (
	"context"
	"errors"
	"testing"

	"github.com/kotofurumiya/sqbl/syntax"
)

// RunUpsert_DoUpdateWhere verifies that ON CONFLICT ... DO UPDATE ... WHERE
// overwrites an existing row only when the condition holds.
// The scenario accepts a write only when it carries a newer version.
// Requires ON CONFLICT support (PostgreSQL, SQLite 3.24+).
func RunUpsert_DoUpdateWhere(t testing.TB, conn Conn, s Suite) {
	t.Helper()
	ctx := context.Background()

	dropIfExists(conn, ctx, "it_upsert_versioned_users")
	t.Cleanup(func() { dropIfExists(conn, ctx, "it_upsert_versioned_users") })

	mustExec(t, conn, ctx, "create table", `
		CREATE TABLE it_upsert_versioned_users (
			id      BIGINT  PRIMARY KEY,
			name    TEXT    NOT NULL,
			version INTEGER NOT NULL
		)
	`)

	upsertSQL := s.InsertInto("it_upsert_versioned_users").
		Columns("id", "name", "version").
		Values(syntax.P(1), syntax.P(2), syntax.P(3)).
		OnConflict("id").
		DoUpdate(`"name" = excluded."name"`, `"version" = excluded."version"`).
		DoUpdateWhere(syntax.Lt("it_upsert_versioned_users.version", "excluded.version")).
		Returning("version").
		ToSql()

	var version int
	if err := conn.QueryRow(ctx, upsertSQL, 1, "Alice", 2).Scan(&version); err != nil {
		t.Fatalf("initial insert: %v", err)
	}

	// An older version does not satisfy the condition, so no row is updated or returned.
	err := conn.QueryRow(ctx, upsertSQL, 1, "Old Alice", 1).Scan(&version)
	if !errors.Is(err, ErrNoRows) {
		t.Errorf("older upsert: expected ErrNoRows, got %v", err)
	}

	if err := conn.QueryRow(ctx, upsertSQL, 1, "New Alice", 3).Scan(&version); err != nil {
		t.Fatalf("newer upsert: %v", err)
	}
	if version != 3 {
		t.Errorf("newer upsert: version = %d; want 3", version)
	}

	var name string
	if err := conn.QueryRow(ctx,
		s.From("it_upsert_versioned_users").Select("name").Where(syntax.Eq("id", syntax.P(1))).ToSql(),
		1,
	).Scan(&name); err != nil {
		t.Fatalf("select: %v", err)
	}
	if name != "New Alice" {
		t.Errorf("name = %q; want %q", name, "New Alice")
	}
}

// RunUpsert_OnConflictWhere verifies that ON CONFLICT (cols) WHERE pred can
// target a partial unique index.
// The scenario keeps email unique among active users only.
// Requires partial index and ON CONFLICT support (PostgreSQL, SQLite 3.24+).
func RunUpsert_OnConflictWhere(t testing.TB, conn Conn, s Suite) {
	t.Helper()
	ctx := context.Background()

	dropIfExists(conn, ctx, "it_upsert_active_users")
	t.Cleanup(func() { dropIfExists(conn, ctx, "it_upsert_active_users") })

	mustExec(t, conn, ctx, "create table", `
		CREATE TABLE it_upsert_active_users (
			id     BIGINT  PRIMARY KEY,
			email  TEXT    NOT NULL,
			name   TEXT    NOT NULL,
			active INTEGER NOT NULL
		)
	`)
	mustExec(t, conn, ctx, "create partial unique index", `
		CREATE UNIQUE INDEX it_upsert_active_users_email
		ON it_upsert_active_users (email) WHERE active = 1
	`)

	upsertSQL := s.InsertInto("it_upsert_active_users").
		Columns("id", "email", "name", "active").
		Values(syntax.P(1), syntax.P(2), syntax.P(3), syntax.P(4)).
		OnConflict("email").
		OnConflictWhere(syntax.Eq("active", 1)).
		DoUpdate(`"name" = excluded."name"`).
		ToSql()

	mustExec(t, conn, ctx, "insert active user", upsertSQL, 1, "alice@example.com", "Alice", 1)
	// Conflicts with the active user, so the name is updated instead of inserting a row.
	mustExec(t, conn, ctx, "upsert active user", upsertSQL, 2, "alice@example.com", "Alice Smith", 1)

	mustExec(t, conn, ctx, "deactivate user",
		s.Update("it_upsert_active_users").Set("active", 0).Where(syntax.Eq("id", syntax.P(1))).ToSql(),
		1,
	)
	// The inactive user is outside the partial index, so the same email is inserted as a new row.
	mustExec(t, conn, ctx, "insert new active user", upsertSQL, 3, "alice@example.com", "Alice Jones", 1)

	type user struct {
		id     int
		name   string
		active int
	}
	rows := mustQuery(t, conn, ctx, "select users",
		s.From("it_upsert_active_users").Select("id", "name", "active").OrderBy("id").ToSql(),
	)
	got := collectRows(t, rows, func(r Rows) (user, error) {
		var u user
		err := r.Scan(&u.id, &u.name, &u.active)
		return u, err
	})
	assertSlice(t, got, []user{
		{1, "Alice Smith", 0},
		{3, "Alice Jones", 1},
	})
}
