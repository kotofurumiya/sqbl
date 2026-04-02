package postgres_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kotofurumiya/sqbl/integrations/shared"
	sqbl "github.com/kotofurumiya/sqbl/sqblpg"
)

// TestPostgres_ReturningAndUpsert exercises the full INSERT/UPDATE/DELETE
// RETURNING workflow together with ON CONFLICT DO NOTHING and DO UPDATE.
// Uses a TEMPORARY articles table with a SERIAL primary key and a UNIQUE slug.
func TestPostgres_ReturningAndUpsert(t *testing.T) {
	t.Parallel()
	conn := connectDB(t)
	defer conn.Close(context.Background())

	ctx := context.Background()
	a := adapt(conn)

	if err := a.Exec(ctx, `
		CREATE TEMPORARY TABLE articles (
			id    SERIAL PRIMARY KEY,
			slug  TEXT   NOT NULL UNIQUE,
			title TEXT   NOT NULL
		)
	`); err != nil {
		t.Fatalf("create table: %v", err)
	}

	// 1. INSERT ... RETURNING id — get the auto-generated id back.
	var id1 int
	if err := a.QueryRow(ctx,
		defaultSuite.InsertInto("articles").
			Columns("slug", "title").
			Values(sqbl.P(1), sqbl.P(2)).
			Returning("id").
			ToSql(),
		"hello-world", "Hello World",
	).Scan(&id1); err != nil {
		t.Fatalf("insert returning: %v", err)
	}
	if id1 == 0 {
		t.Error("RETURNING id: expected non-zero id")
	}

	// 2. INSERT ... ON CONFLICT DO NOTHING RETURNING id — conflicting slug
	//    returns no rows; the caller must see ErrNoRows.
	var id2 int
	err := a.QueryRow(ctx,
		defaultSuite.InsertInto("articles").
			Columns("slug", "title").
			Values(sqbl.P(1), sqbl.P(2)).
			OnConflictDoNothing().
			Returning("id").
			ToSql(),
		"hello-world", "Duplicate",
	).Scan(&id2)
	if !errors.Is(err, shared.ErrNoRows) {
		t.Errorf("ON CONFLICT DO NOTHING: expected ErrNoRows, got %v", err)
	}

	// 3. INSERT ... ON CONFLICT (slug) DO UPDATE SET title = EXCLUDED.title
	//    RETURNING id, title — upsert updates the title in place.
	var id3 int
	var title3 string
	if err := a.QueryRow(ctx,
		defaultSuite.InsertInto("articles").
			Columns("slug", "title").
			Values(sqbl.P(1), sqbl.P(2)).
			OnConflict("slug").
			DoUpdate("title = EXCLUDED.title").
			Returning("id", "title").
			ToSql(),
		"hello-world", "Hello World Updated",
	).Scan(&id3, &title3); err != nil {
		t.Fatalf("upsert returning: %v", err)
	}
	if id3 != id1 {
		t.Errorf("upsert id = %d; want %d", id3, id1)
	}
	if title3 != "Hello World Updated" {
		t.Errorf("upsert title = %q; want %q", title3, "Hello World Updated")
	}

	// 4. UPDATE ... RETURNING title — retrieve the updated value directly.
	var updatedTitle string
	if err := a.QueryRow(ctx,
		defaultSuite.Update("articles").
			Set("title", sqbl.P(1)).
			Where(sqbl.Eq("id", sqbl.P(2))).
			Returning("title").
			ToSql(),
		"Hello World Final", id1,
	).Scan(&updatedTitle); err != nil {
		t.Fatalf("update returning: %v", err)
	}
	if updatedTitle != "Hello World Final" {
		t.Errorf("updated title = %q; want %q", updatedTitle, "Hello World Final")
	}

	// 5. DELETE ... RETURNING id — confirm which row was deleted.
	var deletedID int
	if err := a.QueryRow(ctx,
		defaultSuite.DeleteFrom("articles").
			Where(sqbl.Eq("id", sqbl.P(1))).
			Returning("id").
			ToSql(),
		id1,
	).Scan(&deletedID); err != nil {
		t.Fatalf("delete returning: %v", err)
	}
	if deletedID != id1 {
		t.Errorf("deleted id = %d; want %d", deletedID, id1)
	}
}
