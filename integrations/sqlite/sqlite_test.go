package sqlite_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	_ "modernc.org/sqlite" // register the "sqlite" driver

	"github.com/kotofurumiya/sqbl/integrations/shared"
	"github.com/kotofurumiya/sqbl/sqblsqlite"
)

// sqliteConn adapts *sql.DB to shared.Conn.
type sqliteConn struct{ db *sql.DB }

// sqliteRows wraps *sql.Rows to satisfy shared.Rows.
// sql.Rows.Close() returns error; shared.Rows.Close() returns nothing.
type sqliteRows struct{ rows *sql.Rows }

func (r *sqliteRows) Next() bool             { return r.rows.Next() }
func (r *sqliteRows) Scan(dest ...any) error { return r.rows.Scan(dest...) }
func (r *sqliteRows) Err() error             { return r.rows.Err() }
func (r *sqliteRows) Close()                 { _ = r.rows.Close() }

func (a *sqliteConn) Exec(ctx context.Context, query string, args ...any) error {
	_, err := a.db.ExecContext(ctx, query, args...)
	return err
}

// sqlRow wraps *sql.Row so that sql.ErrNoRows is mapped to shared.ErrNoRows,
// giving shared tests a single canonical sentinel to check with errors.Is.
type sqlRow struct{ row *sql.Row }

func (r sqlRow) Scan(dest ...any) error {
	if err := r.row.Scan(dest...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return shared.ErrNoRows
		}
		return err
	}
	return nil
}

func (a *sqliteConn) QueryRow(ctx context.Context, query string, args ...any) shared.Row {
	return sqlRow{a.db.QueryRowContext(ctx, query, args...)}
}

func (a *sqliteConn) Query(ctx context.Context, query string, args ...any) (shared.Rows, error) {
	rows, err := a.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &sqliteRows{rows}, nil
}

// openDB opens an in-memory SQLite database.
// Each call creates a fully isolated database that is destroyed when closed.
// Foreign key enforcement is explicitly enabled because SQLite disables it by default.
func openDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if _, err := db.ExecContext(context.Background(), "PRAGMA foreign_keys = ON"); err != nil {
		_ = db.Close()
		t.Fatalf("enable foreign_keys: %v", err)
	}
	// Close in t.Cleanup rather than a defer in each test, so that cleanups
	// registered later by the test (e.g. DROP TABLE) still run on an open
	// connection; cleanups run in reverse registration order.
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("close db: %v", err)
		}
	})
	return db
}

func openConn(t *testing.T) shared.Conn { return &sqliteConn{openDB(t)} }

var suite = shared.Suite{
	From:        sqblsqlite.From,
	InsertInto:  sqblsqlite.InsertInto,
	Update:      sqblsqlite.Update,
	DeleteFrom:  sqblsqlite.DeleteFrom,
	CreateTable: sqblsqlite.CreateTable,
	CreateIndex: sqblsqlite.CreateIndex,
	DropTable:   sqblsqlite.DropTable,
	DropIndex:   sqblsqlite.DropIndex,
	AlterTable:  sqblsqlite.AlterTable,

	PositionalParams: true,
	IndexedParams:    true,
	NamedParams:      true,
	FullJoin:         true,
	Returning:        true,
	OnConflict:       true,
	PartialIndex:     true,
}

func TestShared(t *testing.T) {
	t.Parallel()
	shared.RunAll(t, openConn, suite)
}

// TestInsertOrConflict resolves a primary-key conflict with each SQLite
// INSERT OR action.
func TestInsertOrConflict(t *testing.T) {
	t.Parallel()
	conn := openConn(t)
	shared.MustExec(t, conn, `CREATE TEMPORARY TABLE config (key TEXT PRIMARY KEY, value TEXT NOT NULL)`)
	shared.MustExec(t, conn, `INSERT INTO config (key, value) VALUES ('theme', 'light')`)

	insert := sqblsqlite.InsertInto("config").Columns("key", "value").Values(sqblsqlite.P(1), sqblsqlite.P(2))
	value := sqblsqlite.From("config").Select("value").ToSql()

	shared.MustExec(t, conn, insert.OrIgnore().ToSql(), "theme", "dark")
	shared.AssertSlice(t, shared.QueryColumn[string](t, conn, value), []string{"light"})

	shared.MustExec(t, conn, insert.OrReplace().ToSql(), "theme", "dark")
	shared.AssertSlice(t, shared.QueryColumn[string](t, conn, value), []string{"dark"})

	for _, failing := range []string{
		insert.OrAbort().ToSql(),
		insert.OrFail().ToSql(),
		insert.OrRollback().ToSql(),
	} {
		shared.MustFail(t, conn, "duplicate key", failing, "theme", "blue")
	}
	shared.AssertSlice(t, shared.QueryColumn[string](t, conn, value), []string{"dark"})
}

// TestCreateTableOptions applies table options after the column list.
// STRICT makes SQLite reject values that do not match the declared type.
func TestCreateTableOptions(t *testing.T) {
	t.Parallel()
	conn := openConn(t)
	shared.MustExec(t, conn,
		sqblsqlite.CreateTable("counters").
			Column("id", "INTEGER", "PRIMARY KEY").
			Column("hits", "INTEGER", "NOT NULL").
			Options("STRICT, WITHOUT ROWID").
			ToSql())

	insert := sqblsqlite.InsertInto("counters").Columns("id", "hits").Values(sqblsqlite.P(1), sqblsqlite.P(2)).ToSql()
	shared.MustExec(t, conn, insert, 1, 10)
	shared.MustFail(t, conn, "TEXT into a STRICT INTEGER column", insert, 2, "many")
}
