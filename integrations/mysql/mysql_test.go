package mysql_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql" // register the "mysql" driver

	"github.com/kotofurumiya/sqbl/integrations/shared"
	"github.com/kotofurumiya/sqbl/sqblmysql"
)

// mysqlConn adapts *sql.DB to shared.Conn.
type mysqlConn struct{ db *sql.DB }

// mysqlRows wraps *sql.Rows to satisfy shared.Rows.
// sql.Rows.Close() returns error; shared.Rows.Close() returns nothing.
type mysqlRows struct{ rows *sql.Rows }

func (r *mysqlRows) Next() bool             { return r.rows.Next() }
func (r *mysqlRows) Scan(dest ...any) error { return r.rows.Scan(dest...) }
func (r *mysqlRows) Err() error             { return r.rows.Err() }
func (r *mysqlRows) Close()                 { _ = r.rows.Close() }

func (a *mysqlConn) Exec(ctx context.Context, query string, args ...any) error {
	_, err := a.db.ExecContext(ctx, query, args...)
	return err
}

// mysqlRow wraps *sql.Row so that sql.ErrNoRows is mapped to shared.ErrNoRows,
// giving shared tests a single canonical sentinel to check with errors.Is.
type mysqlRow struct{ row *sql.Row }

func (r mysqlRow) Scan(dest ...any) error {
	if err := r.row.Scan(dest...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return shared.ErrNoRows
		}
		return err
	}
	return nil
}

func (a *mysqlConn) QueryRow(ctx context.Context, query string, args ...any) shared.Row {
	return mysqlRow{a.db.QueryRowContext(ctx, query, args...)}
}

func (a *mysqlConn) Query(ctx context.Context, query string, args ...any) (shared.Rows, error) {
	rows, err := a.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &mysqlRows{rows}, nil
}

// openDB opens a connection to the test MariaDB instance.
// It reads the DSN from the MYSQL_DSN environment variable, falling back to
// the default address used by the compose.yaml in this repository.
//
// SetMaxOpenConns(1) ensures all queries in the test share the same
// underlying connection, so the ANSI_QUOTES session variable set below
// remains effective for the lifetime of the test.
func openDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		dsn = "sqbl:sqbl@tcp(localhost:3306)/sqbl_test?parseTime=true"
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("MySQL not available: %v", err)
	}
	if err := db.PingContext(context.Background()); err != nil {
		_ = db.Close()
		t.Fatalf("MySQL not available: %v", err)
	}
	// Keep a single connection per test so the session variable below is stable.
	db.SetMaxOpenConns(1)
	// Enable ANSI_QUOTES so raw SQL in shared tests can use double-quote
	// identifier syntax (e.g. "order", "People") without ambiguity.
	if _, err := db.ExecContext(context.Background(),
		"SET SESSION sql_mode = CONCAT(@@sql_mode, ',ANSI_QUOTES')"); err != nil {
		_ = db.Close()
		t.Fatalf("set ANSI_QUOTES: %v", err)
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

func openConn(t *testing.T) shared.Conn { return &mysqlConn{openDB(t)} }

var suite = shared.Suite{
	From:        sqblmysql.From,
	InsertInto:  sqblmysql.InsertInto,
	Update:      sqblmysql.Update,
	DeleteFrom:  sqblmysql.DeleteFrom,
	CreateTable: sqblmysql.CreateTable,
	CreateIndex: sqblmysql.CreateIndex,
	DropTable:   sqblmysql.DropTable,
	DropIndex:   sqblmysql.DropIndex,
	AlterTable:  sqblmysql.AlterTable,

	DropIndexOnTable: true,
	PositionalParams: true,
	SetOpAll:         true,
}

func TestShared(t *testing.T) {
	t.Parallel()
	shared.RunAll(t, openConn, suite)
}

// TestOnDuplicateKeyUpdate upserts a tag counter twice, once with raw
// assignments and once with structured ones, and checks the merged row.
func TestOnDuplicateKeyUpdate(t *testing.T) {
	t.Parallel()
	conn := openConn(t)
	shared.MustExec(t, conn, `
		CREATE TEMPORARY TABLE tags (
			tag   VARCHAR(32) PRIMARY KEY,
			label VARCHAR(32) NOT NULL,
			hits  INTEGER     NOT NULL
		)
	`)
	insert := sqblmysql.InsertInto("tags").
		Columns("tag", "label", "hits").
		Values(sqblmysql.P(), sqblmysql.P(), sqblmysql.P())
	row := sqblmysql.From("tags").Select("label", "hits").ToSql()
	assertRow := func(wantLabel string, wantHits int) {
		t.Helper()
		var label string
		var hits int
		shared.MustScan(t, conn, row, nil, &label, &hits)
		if label != wantLabel || hits != wantHits {
			t.Errorf("got (%q, %d); want (%q, %d)", label, hits, wantLabel, wantHits)
		}
	}

	shared.MustExec(t, conn, insert.OnDuplicateKeyUpdate("hits = hits + 1").ToSql(), "go", "Go", 1)
	shared.MustExec(t, conn, insert.OnDuplicateKeyUpdate("hits = hits + 1").ToSql(), "go", "Ignored", 1)
	assertRow("Go", 2)

	structured := insert.
		OnDuplicateKeyUpdateSet("label", sqblmysql.Fn("VALUES", "label")).
		OnDuplicateKeyUpdateSet("hits", sqblmysql.Raw("`hits` + 10")).
		ToSql()
	shared.MustExec(t, conn, structured, "go", "Golang", 1)
	assertRow("Golang", 12)
}
