package shared

import (
	"context"
	"testing"

	"github.com/kotofurumiya/sqbl/syntax"
)

// RunCreateTable verifies that the CREATE TABLE builder generates SQL that the
// database accepts, and that the resulting table is usable for INSERT and SELECT.
func RunCreateTable(t testing.TB, conn Conn, s Suite) {
	t.Helper()
	ctx := context.Background()

	dropIfExists(conn, ctx, "it_create_users")
	t.Cleanup(func() { dropIfExists(conn, ctx, "it_create_users") })

	createSQL := s.CreateTable("it_create_users").
		IfNotExists().
		Column("id", "BIGINT", "NOT NULL").
		Column("name", "TEXT", "NOT NULL").
		Column("email", "TEXT").
		PrimaryKey("id").
		ToSql()
	mustExec(t, conn, ctx, "create table", createSQL)

	mustExec(t, conn, ctx, "insert",
		s.InsertInto("it_create_users").
			Columns("id", "name", "email").
			Values(syntax.P(1), syntax.P(2), syntax.P(3)).
			Values(syntax.P(4), syntax.P(5), syntax.P(6)).
			ToSql(),
		1, "Alice", "alice@example.com", 2, "Bob", nil,
	)

	selectSQL := s.From("it_create_users").
		Select("name").
		Where(syntax.Eq("id", syntax.P(1))).
		ToSql()

	var name string
	if err := conn.QueryRow(ctx, selectSQL, 1).Scan(&name); err != nil {
		t.Fatalf("select: %v", err)
	}
	if name != "Alice" {
		t.Errorf("name = %q; want Alice", name)
	}

	// Idempotency: running the same CREATE TABLE IF NOT EXISTS again must not error.
	mustExec(t, conn, ctx, "second create (IF NOT EXISTS)", createSQL)
}

// RunCreateTable_UniqueConstraint verifies that a UNIQUE table constraint is
// generated correctly and enforced by the database.
func RunCreateTable_UniqueConstraint(t testing.TB, conn Conn, s Suite) {
	t.Helper()
	ctx := context.Background()

	dropIfExists(conn, ctx, "it_create_tags")
	t.Cleanup(func() { dropIfExists(conn, ctx, "it_create_tags") })

	mustExec(t, conn, ctx, "create table",
		s.CreateTable("it_create_tags").
			Column("id", "BIGINT", "NOT NULL").
			Column("name", "TEXT", "NOT NULL").
			PrimaryKey("id").
			Unique("name").
			ToSql(),
	)

	insertSQL := s.InsertInto("it_create_tags").
		Columns("id", "name").
		Values(syntax.P(1), syntax.P(2)).
		ToSql()
	mustExec(t, conn, ctx, "insert", insertSQL, 1, "go")

	// Inserting a duplicate name must fail due to the UNIQUE constraint.
	if err := conn.Exec(ctx, insertSQL, 2, "go"); err == nil {
		t.Error("expected unique constraint violation; got no error")
	}
}

// RunCreateTable_ForeignKey verifies that a FOREIGN KEY constraint is accepted
// and enforced: inserting a row that references a non-existent parent must fail,
// and ON DELETE CASCADE must remove child rows when the parent is deleted.
func RunCreateTable_ForeignKey(t testing.TB, conn Conn, s Suite) {
	t.Helper()
	ctx := context.Background()

	// Drop in reverse dependency order to handle leftover tables.
	dropIfExists(conn, ctx, "it_fk_orders", "it_fk_users")
	t.Cleanup(func() { dropIfExists(conn, ctx, "it_fk_orders", "it_fk_users") })

	mustExec(t, conn, ctx, "create users table",
		s.CreateTable("it_fk_users").
			Column("id", "BIGINT", "NOT NULL").
			PrimaryKey("id").
			ToSql(),
	)
	mustExec(t, conn, ctx, "create orders table",
		s.CreateTable("it_fk_orders").
			Column("id", "BIGINT", "NOT NULL").
			Column("user_id", "BIGINT", "NOT NULL").
			PrimaryKey("id").
			ForeignKey([]string{"user_id"}, "it_fk_users", []string{"id"}, "ON DELETE CASCADE").
			ToSql(),
	)

	// Inserting an order that references a non-existent user must fail.
	badInsert := s.InsertInto("it_fk_orders").
		Columns("id", "user_id").
		Values(syntax.P(1), syntax.P(2)).
		ToSql()
	if err := conn.Exec(ctx, badInsert, 1, 999); err == nil {
		t.Error("expected foreign key violation for non-existent user; got no error")
	}

	// Insert the parent user and a child order.
	mustExec(t, conn, ctx, "insert user",
		s.InsertInto("it_fk_users").Columns("id").Values(syntax.P(1)).ToSql(), 1)
	mustExec(t, conn, ctx, "insert order",
		s.InsertInto("it_fk_orders").Columns("id", "user_id").Values(syntax.P(1), syntax.P(2)).ToSql(), 1, 1)

	// Deleting the parent user must cascade and remove the child order.
	mustExec(t, conn, ctx, "delete user",
		s.DeleteFrom("it_fk_users").Where(syntax.Eq("id", syntax.P(1))).ToSql(), 1)

	var count int
	if err := conn.QueryRow(ctx, s.From("it_fk_orders").Select("COUNT(*)").ToSql()).Scan(&count); err != nil {
		t.Fatalf("count orders: %v", err)
	}
	if count != 0 {
		t.Errorf("orders after cascade delete = %d; want 0", count)
	}
}

// RunCreateTable_Check verifies that a CHECK constraint built with
// syntax.Expression is accepted and enforced by the database.
func RunCreateTable_Check(t testing.TB, conn Conn, s Suite) {
	t.Helper()
	ctx := context.Background()

	dropIfExists(conn, ctx, "it_check_products")
	t.Cleanup(func() { dropIfExists(conn, ctx, "it_check_products") })

	mustExec(t, conn, ctx, "create table",
		s.CreateTable("it_check_products").
			Column("id", "BIGINT", "NOT NULL").
			Column("price", "INTEGER", "NOT NULL").
			PrimaryKey("id").
			Check(syntax.Gt("price", 0)).
			ToSql(),
	)

	insertSQL := s.InsertInto("it_check_products").
		Columns("id", "price").
		Values(syntax.P(1), syntax.P(2)).
		ToSql()

	// A valid row must succeed.
	mustExec(t, conn, ctx, "insert valid row", insertSQL, 1, 100)

	// Rows violating the CHECK (price <= 0) must fail.
	if err := conn.Exec(ctx, insertSQL, 2, 0); err == nil {
		t.Error("expected check constraint violation for price=0; got no error")
	}
	if err := conn.Exec(ctx, insertSQL, 3, -1); err == nil {
		t.Error("expected check constraint violation for price=-1; got no error")
	}
}

// RunCreateIndex verifies that the CREATE INDEX builder generates SQL accepted
// by the database and that the UNIQUE index is enforced.
// The active column uses INTEGER (1=active) instead of BOOLEAN to remain
// portable across dialects.
func RunCreateIndex(t testing.TB, conn Conn, s Suite) {
	t.Helper()
	ctx := context.Background()

	dropIfExists(conn, ctx, "it_idx_users")
	t.Cleanup(func() { dropIfExists(conn, ctx, "it_idx_users") })

	mustExec(t, conn, ctx, "create table", `
		CREATE TABLE it_idx_users (
			id     BIGINT PRIMARY KEY,
			email  TEXT NOT NULL,
			active INTEGER NOT NULL DEFAULT 1
		)
	`)

	// Build and apply a UNIQUE index on email.
	mustExec(t, conn, ctx, "create unique index",
		s.CreateIndex("it_idx_users_email").
			On("it_idx_users").
			Columns("email").
			Unique().
			ToSql(),
	)

	insertSQL := s.InsertInto("it_idx_users").
		Columns("id", "email", "active").
		Values(syntax.P(1), syntax.P(2), syntax.P(3)).
		ToSql()
	mustExec(t, conn, ctx, "insert", insertSQL, 1, "alice@example.com", 1)

	// A duplicate email must be rejected by the unique index.
	if err := conn.Exec(ctx, insertSQL, 2, "alice@example.com", 1); err == nil {
		t.Error("expected unique index violation for duplicate email; got no error")
	}
}

// RunCreateIndex_Partial verifies that a partial index (CREATE INDEX ... WHERE)
// is accepted by the database. MySQL/MariaDB do not support partial indexes;
// call this function only for PostgreSQL and SQLite test suites.
func RunCreateIndex_Partial(t testing.TB, conn Conn, s Suite) {
	t.Helper()
	ctx := context.Background()

	dropIfExists(conn, ctx, "it_idx_partial_users")
	t.Cleanup(func() { dropIfExists(conn, ctx, "it_idx_partial_users") })

	mustExec(t, conn, ctx, "create table", `
		CREATE TABLE it_idx_partial_users (
			id     BIGINT PRIMARY KEY,
			email  TEXT NOT NULL,
			active INTEGER NOT NULL DEFAULT 1
		)
	`)

	// Build and apply a partial index covering only active users.
	mustExec(t, conn, ctx, "create partial index",
		s.CreateIndex("it_idx_active_users").
			On("it_idx_partial_users").
			Columns("email").
			Where(syntax.Eq("active", 1)).
			IfNotExists().
			ToSql(),
	)
}
