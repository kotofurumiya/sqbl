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
	return db
}

// adapt wraps a *sql.DB as a shared.Conn.
func adapt(db *sql.DB) shared.Conn { return &mysqlConn{db} }

// defaultSuite wires the MySQL builder constructors into a shared.Suite.
var defaultSuite = shared.Suite{
	From:        sqblmysql.From,
	InsertInto:  sqblmysql.InsertInto,
	Update:      sqblmysql.Update,
	DeleteFrom:  sqblmysql.DeleteFrom,
	CreateTable: sqblmysql.CreateTable,
	CreateIndex: sqblmysql.CreateIndex,
	DropTable:   sqblmysql.DropTable,
	DropIndex:   sqblmysql.DropIndex,
	AlterTable:  sqblmysql.AlterTable,
}

func TestMySQL_CRUD(t *testing.T) {
	t.Parallel()
	db := openDB(t)
	defer db.Close()
	shared.RunCRUD(t, adapt(db), defaultSuite)
}

func TestMySQL_Injection_ValueAsParameter(t *testing.T) {
	t.Parallel()
	db := openDB(t)
	defer db.Close()
	shared.RunInjection_ValueAsParameter(t, adapt(db), defaultSuite)
}

func TestMySQL_Injection_StoredValue(t *testing.T) {
	t.Parallel()
	db := openDB(t)
	defer db.Close()
	shared.RunInjection_StoredValue(t, adapt(db), defaultSuite)
}

func TestMySQL_Injection_LikeWildcardsInEqValue(t *testing.T) {
	t.Parallel()
	db := openDB(t)
	defer db.Close()
	shared.RunInjection_LikeWildcardsInEqValue(t, adapt(db), defaultSuite)
}

func TestMySQL_Injection_QuotedIdentifier(t *testing.T) {
	t.Parallel()
	db := openDB(t)
	defer db.Close()
	shared.RunInjection_QuotedIdentifier(t, adapt(db), defaultSuite)
}

func TestMySQL_Injection_OrderByIdentifier(t *testing.T) {
	t.Parallel()
	db := openDB(t)
	defer db.Close()
	shared.RunInjection_OrderByIdentifier(t, adapt(db), defaultSuite)
}

func TestMySQL_Injection_TableNameIdentifier(t *testing.T) {
	t.Parallel()
	db := openDB(t)
	defer db.Close()
	shared.RunInjection_TableNameIdentifier(t, adapt(db), defaultSuite)
}

func TestMySQL_Injection_SelectColumnIdentifier(t *testing.T) {
	t.Parallel()
	db := openDB(t)
	defer db.Close()
	shared.RunInjection_SelectColumnIdentifier(t, adapt(db), defaultSuite)
}

func TestMySQL_Injection_AliasIdentifier(t *testing.T) {
	t.Parallel()
	db := openDB(t)
	defer db.Close()
	shared.RunInjection_AliasIdentifier(t, adapt(db), defaultSuite)
}

func TestMySQL_Injection_CommentBasedPayloads(t *testing.T) {
	t.Parallel()
	db := openDB(t)
	defer db.Close()
	shared.RunInjection_CommentBasedPayloads(t, adapt(db), defaultSuite)
}

func TestMySQL_Injection_StackedInsertPayload(t *testing.T) {
	t.Parallel()
	db := openDB(t)
	defer db.Close()
	shared.RunInjection_StackedInsertPayload(t, adapt(db), defaultSuite)
}

func TestMySQL_JoinWithAliases(t *testing.T) {
	t.Parallel()
	db := openDB(t)
	defer db.Close()
	shared.RunJoinWithAliases(t, adapt(db), defaultSuite)
}

func TestMySQL_Subquery(t *testing.T) {
	t.Parallel()
	db := openDB(t)
	defer db.Close()
	shared.RunSubquery(t, adapt(db), defaultSuite)
}

func TestMySQL_SetOperations(t *testing.T) {
	t.Parallel()
	db := openDB(t)
	defer db.Close()
	shared.RunSetOperations(t, adapt(db), defaultSuite)
}

func TestMySQL_CTEAndAggregation(t *testing.T) {
	t.Parallel()
	db := openDB(t)
	defer db.Close()
	shared.RunCTEAndAggregation(t, adapt(db), defaultSuite)
}

func TestMySQL_Identifiers_ReservedKeywords(t *testing.T) {
	t.Parallel()
	db := openDB(t)
	defer db.Close()
	shared.RunIdentifiers_ReservedKeywords(t, adapt(db), defaultSuite)
}

func TestMySQL_Identifiers_MixedCase(t *testing.T) {
	t.Parallel()
	db := openDB(t)
	defer db.Close()
	shared.RunIdentifiers_MixedCase(t, adapt(db), defaultSuite)
}

func TestMySQL_Identifiers_EmbeddedDoubleQuote(t *testing.T) {
	t.Parallel()
	db := openDB(t)
	defer db.Close()
	shared.RunIdentifiers_EmbeddedDoubleQuote(t, adapt(db), defaultSuite)
}

func TestMySQL_CreateTable(t *testing.T) {
	t.Parallel()
	db := openDB(t)
	defer db.Close()
	shared.RunCreateTable(t, adapt(db), defaultSuite)
}

func TestMySQL_CreateTable_UniqueConstraint(t *testing.T) {
	t.Parallel()
	db := openDB(t)
	defer db.Close()
	shared.RunCreateTable_UniqueConstraint(t, adapt(db), defaultSuite)
}

func TestMySQL_CreateTable_ForeignKey(t *testing.T) {
	t.Parallel()
	db := openDB(t)
	defer db.Close()
	shared.RunCreateTable_ForeignKey(t, adapt(db), defaultSuite)
}

func TestMySQL_CreateTable_Check(t *testing.T) {
	t.Parallel()
	db := openDB(t)
	defer db.Close()
	shared.RunCreateTable_Check(t, adapt(db), defaultSuite)
}

func TestMySQL_CreateIndex(t *testing.T) {
	t.Parallel()
	db := openDB(t)
	defer db.Close()
	shared.RunCreateIndex(t, adapt(db), defaultSuite)
}

// TestMySQL_CreateIndex_Partial is intentionally absent: MySQL/MariaDB do not
// support partial indexes (WHERE clause in CREATE INDEX).

func TestMySQL_FilteringAndPagination(t *testing.T) {
	t.Parallel()
	db := openDB(t)
	defer db.Close()
	shared.RunFilteringAndPagination(t, adapt(db), defaultSuite)
}

func TestMySQL_RecursiveCTE(t *testing.T) {
	t.Parallel()
	db := openDB(t)
	defer db.Close()
	shared.RunRecursiveCTE(t, adapt(db), defaultSuite)
}

func TestMySQL_SchemaEvolution(t *testing.T) {
	t.Parallel()
	db := openDB(t)
	defer db.Close()
	shared.RunSchemaEvolution(t, adapt(db), defaultSuite)
}

func TestMySQL_OnDuplicateKeyUpdate(t *testing.T) {
	t.Parallel()
	db := openDB(t)
	defer db.Close()

	ctx := context.Background()
	a := adapt(db)

	if err := a.Exec(ctx, `
		CREATE TEMPORARY TABLE tag_counts (
			tag   TEXT        NOT NULL,
			count INTEGER     NOT NULL DEFAULT 0,
			UNIQUE KEY (tag(255))
		)
	`); err != nil {
		t.Fatalf("create table: %v", err)
	}

	// First INSERT: new tag with count 1.
	insertSQL := defaultSuite.InsertInto("tag_counts").
		Columns("tag", "count").
		Values(sqblmysql.P(1), sqblmysql.P(2)).
		OnDuplicateKeyUpdate("count = count + 1").
		ToSql()

	if err := a.Exec(ctx, insertSQL, "go", 1); err != nil {
		t.Fatalf("first insert: %v", err)
	}

	// Second INSERT with same tag: ON DUPLICATE KEY UPDATE increments count.
	if err := a.Exec(ctx, insertSQL, "go", 1); err != nil {
		t.Fatalf("second insert (duplicate): %v", err)
	}

	var count int
	if err := a.QueryRow(ctx,
		defaultSuite.From("tag_counts").
			Select("count").
			Where(sqblmysql.Eq("tag", sqblmysql.P(1))).
			ToSql(),
		"go",
	).Scan(&count); err != nil {
		t.Fatalf("select count: %v", err)
	}
	if count != 2 {
		t.Errorf("count = %d; want 2", count)
	}
}
