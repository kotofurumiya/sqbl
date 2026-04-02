package shared

import (
	"context"
	"testing"

	"github.com/kotofurumiya/sqbl/syntax"
)

// RunSchemaEvolution simulates an incremental migration workflow:
// create a table, seed data, add a column, update and query the new column,
// create then drop an index, and finally drop the table and verify it is gone.
func RunSchemaEvolution(t testing.TB, conn Conn, s Suite) {
	t.Helper()
	ctx := context.Background()

	const table = "it_evo_users"
	const index = "it_evo_users_name_idx"

	dropIfExists(conn, ctx, table)
	t.Cleanup(func() { dropIfExists(conn, ctx, table) })

	// 1. Create table.
	mustExec(t, conn, ctx, "create table",
		s.CreateTable(table).
			Column("id", "BIGINT", "NOT NULL").
			Column("name", "TEXT", "NOT NULL").
			PrimaryKey("id").
			ToSql(),
	)

	// 2. Seed two rows.
	mustExec(t, conn, ctx, "insert",
		s.InsertInto(table).
			Columns("id", "name").
			Values(syntax.P(1), syntax.P(2)).
			Values(syntax.P(3), syntax.P(4)).
			ToSql(),
		1, "Alice", 2, "Bob",
	)

	// 3. Add a nullable bio column via ALTER TABLE.
	mustExec(t, conn, ctx, "alter table add bio",
		s.AlterTable(table).AddColumn("bio", "TEXT").ToSql(),
	)

	// 4. Set bio for id=1, then read it back.
	mustExec(t, conn, ctx, "update bio",
		s.Update(table).
			Set("bio", syntax.P(1)).
			Where(syntax.Eq("id", syntax.P(2))).
			ToSql(),
		"hello", 1,
	)

	var bio string
	if err := conn.QueryRow(ctx,
		s.From(table).Select("bio").Where(syntax.Eq("id", syntax.P(1))).ToSql(),
		1,
	).Scan(&bio); err != nil {
		t.Fatalf("select bio: %v", err)
	}
	if bio != "hello" {
		t.Errorf("bio = %q; want %q", bio, "hello")
	}

	// 5. Create an index on name.
	// On(table) is required by MySQL and silently ignored by PostgreSQL/SQLite.
	mustExec(t, conn, ctx, "create index",
		s.CreateIndex(index).On(table).Columns("name").ToSql(),
	)

	// 6. Drop the index.
	// IfExists() is applied for PostgreSQL/SQLite; MySQL ignores it.
	// On(table) is required by MySQL.
	mustExec(t, conn, ctx, "drop index",
		s.DropIndex(index).On(table).IfExists().ToSql(),
	)

	// 7. Drop the table itself.
	mustExec(t, conn, ctx, "drop table",
		s.DropTable(table).ToSql(),
	)

	// 8. Verify the table is gone: a SELECT must now fail.
	selectSQL := s.From(table).Select("id").ToSql()
	rows, err := conn.Query(ctx, selectSQL)
	if err == nil {
		rows.Close()
		t.Error("expected error querying dropped table; got nil")
	}
}
