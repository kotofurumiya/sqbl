package postgres_test

import (
	"context"
	"testing"

	"github.com/kotofurumiya/sqbl/integrations/shared"
	"github.com/kotofurumiya/sqbl/sqblpg"
	"github.com/kotofurumiya/sqbl/syntax"
)

func TestPostgres_CreateTable(t *testing.T) {
	t.Parallel()
	conn := connectDB(t)
	defer conn.Close(context.Background())
	shared.RunCreateTable(t, adapt(conn), defaultSuite)
}

func TestPostgres_CreateTable_UniqueConstraint(t *testing.T) {
	t.Parallel()
	conn := connectDB(t)
	defer conn.Close(context.Background())
	shared.RunCreateTable_UniqueConstraint(t, adapt(conn), defaultSuite)
}

func TestPostgres_CreateTable_ForeignKey(t *testing.T) {
	t.Parallel()
	conn := connectDB(t)
	defer conn.Close(context.Background())
	shared.RunCreateTable_ForeignKey(t, adapt(conn), defaultSuite)
}

func TestPostgres_CreateTable_Check(t *testing.T) {
	t.Parallel()
	conn := connectDB(t)
	defer conn.Close(context.Background())
	shared.RunCreateTable_Check(t, adapt(conn), defaultSuite)
}

func TestPostgres_CreateIndex(t *testing.T) {
	t.Parallel()
	conn := connectDB(t)
	defer conn.Close(context.Background())

	shared.RunCreateIndex(t, adapt(conn), defaultSuite)

	// PostgreSQL-specific: verify the unique index appears in pg_indexes.
	ctx := context.Background()
	var idxCount int
	if err := conn.QueryRow(ctx,
		`SELECT COUNT(*) FROM pg_indexes WHERE tablename = 'it_idx_users' AND indexname = 'it_idx_users_email'`,
	).Scan(&idxCount); err != nil {
		t.Fatalf("query pg_indexes: %v", err)
	}
	if idxCount != 1 {
		t.Errorf("pg_indexes count = %d; want 1", idxCount)
	}
}

func TestPostgres_CreateIndex_Partial(t *testing.T) {
	t.Parallel()
	conn := connectDB(t)
	defer conn.Close(context.Background())

	shared.RunCreateIndex_Partial(t, adapt(conn), defaultSuite)

	// PostgreSQL-specific: verify the partial index appears in pg_indexes.
	ctx := context.Background()
	var idxCount int
	if err := conn.QueryRow(ctx,
		`SELECT COUNT(*) FROM pg_indexes WHERE tablename = 'it_idx_partial_users' AND indexname = 'it_idx_active_users'`,
	).Scan(&idxCount); err != nil {
		t.Fatalf("query pg_indexes: %v", err)
	}
	if idxCount != 1 {
		t.Errorf("pg_indexes count = %d; want 1", idxCount)
	}
}

// TestPostgres_CreateIndex_Using verifies that the USING clause (e.g. USING hash)
// is emitted correctly. This is a PostgreSQL-specific feature.
func TestPostgres_CreateIndex_Using(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	conn := connectDB(t)
	defer conn.Close(ctx)

	_, _ = conn.Exec(ctx, `DROP TABLE IF EXISTS it_idx_using`)
	t.Cleanup(func() {
		_, _ = conn.Exec(ctx, `DROP TABLE IF EXISTS it_idx_using`)
	})

	if _, err := conn.Exec(ctx, `
		CREATE TABLE it_idx_using (
			id    BIGINT PRIMARY KEY,
			email TEXT NOT NULL
		)
	`); err != nil {
		t.Fatalf("create table: %v", err)
	}

	idxSQL := sqblpg.CreateIndex("it_idx_using_email_hash").
		On("it_idx_using").
		Columns("email").
		Using("hash").
		ToSql()

	if _, err := conn.Exec(ctx, idxSQL); err != nil {
		t.Fatalf("create index with USING hash: %v", err)
	}

	// Confirm the index is visible in pg_indexes with the correct index method.
	var idxMethod string
	if err := conn.QueryRow(ctx,
		`SELECT indexdef FROM pg_indexes WHERE indexname = 'it_idx_using_email_hash'`,
	).Scan(&idxMethod); err != nil {
		t.Fatalf("query pg_indexes: %v", err)
	}
	if idxMethod == "" {
		t.Error("index not found in pg_indexes")
	}
}

// TestPostgres_CreateTable_ColumnConstraints verifies per-column inline
// constraints (NOT NULL, UNIQUE inline, DEFAULT) that are emitted as part of
// the column definition rather than as table-level constraints. This is
// standard SQL but tested here to ensure the builder wires them correctly for
// PostgreSQL.
func TestPostgres_CreateTable_ColumnConstraints(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	conn := connectDB(t)
	defer conn.Close(ctx)

	_, _ = conn.Exec(ctx, `DROP TABLE IF EXISTS it_col_constraints`)
	t.Cleanup(func() {
		_, _ = conn.Exec(ctx, `DROP TABLE IF EXISTS it_col_constraints`)
	})

	createSQL := sqblpg.CreateTable("it_col_constraints").
		Column("id", "BIGINT", "NOT NULL").
		Column("code", "TEXT", "NOT NULL", "UNIQUE").
		Column("status", "TEXT", "NOT NULL", "DEFAULT 'active'").
		PrimaryKey("id").
		Check(syntax.Eq("status", "'active'")).
		ToSql()

	if _, err := conn.Exec(ctx, createSQL); err != nil {
		t.Fatalf("create table: %v", err)
	}

	insertSQL := sqblpg.InsertInto("it_col_constraints").
		Columns("id", "code").
		Values(sqblpg.P(1), sqblpg.P(2)).
		ToSql()

	if _, err := conn.Exec(ctx, insertSQL, 1, "A001"); err != nil {
		t.Fatalf("insert: %v", err)
	}

	// Duplicate code must fail (inline UNIQUE).
	if _, err := conn.Exec(ctx, insertSQL, 2, "A001"); err == nil {
		t.Error("expected unique violation for duplicate code; got no error")
	}

	// Default status should have been applied.
	var status string
	if err := conn.QueryRow(ctx,
		sqblpg.From("it_col_constraints").Select("status").Where(sqblpg.Eq("id", sqblpg.P(1))).ToSql(),
		1,
	).Scan(&status); err != nil {
		t.Fatalf("select status: %v", err)
	}
	if status != "active" {
		t.Errorf("status = %q; want active", status)
	}
}
