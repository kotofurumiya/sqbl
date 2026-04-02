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
		db.Close()
		t.Fatalf("enable foreign_keys: %v", err)
	}
	return db
}

// adapt wraps a *sql.DB as a shared.Conn.
func adapt(db *sql.DB) shared.Conn { return &sqliteConn{db} }

// defaultSuite wires the SQLite builder constructors into a shared.Suite.
var defaultSuite = shared.Suite{
	From:        sqblsqlite.From,
	InsertInto:  sqblsqlite.InsertInto,
	Update:      sqblsqlite.Update,
	DeleteFrom:  sqblsqlite.DeleteFrom,
	CreateTable: sqblsqlite.CreateTable,
	CreateIndex: sqblsqlite.CreateIndex,
	DropTable:   sqblsqlite.DropTable,
	DropIndex:   sqblsqlite.DropIndex,
	AlterTable:  sqblsqlite.AlterTable,
}

func TestSQLite_CRUD(t *testing.T) {
	t.Parallel()
	db := openDB(t)
	defer db.Close()
	shared.RunCRUD(t, adapt(db), defaultSuite)
}

func TestSQLite_Injection_ValueAsParameter(t *testing.T) {
	t.Parallel()
	db := openDB(t)
	defer db.Close()
	shared.RunInjection_ValueAsParameter(t, adapt(db), defaultSuite)
}

func TestSQLite_Injection_StoredValue(t *testing.T) {
	t.Parallel()
	db := openDB(t)
	defer db.Close()
	shared.RunInjection_StoredValue(t, adapt(db), defaultSuite)
}

func TestSQLite_Injection_LikeWildcardsInEqValue(t *testing.T) {
	t.Parallel()
	db := openDB(t)
	defer db.Close()
	shared.RunInjection_LikeWildcardsInEqValue(t, adapt(db), defaultSuite)
}

func TestSQLite_Injection_QuotedIdentifier(t *testing.T) {
	t.Parallel()
	db := openDB(t)
	defer db.Close()
	shared.RunInjection_QuotedIdentifier(t, adapt(db), defaultSuite)
}

func TestSQLite_Injection_OrderByIdentifier(t *testing.T) {
	t.Parallel()
	db := openDB(t)
	defer db.Close()
	shared.RunInjection_OrderByIdentifier(t, adapt(db), defaultSuite)
}

func TestSQLite_Injection_TableNameIdentifier(t *testing.T) {
	t.Parallel()
	db := openDB(t)
	defer db.Close()
	shared.RunInjection_TableNameIdentifier(t, adapt(db), defaultSuite)
}

func TestSQLite_Injection_SelectColumnIdentifier(t *testing.T) {
	t.Parallel()
	db := openDB(t)
	defer db.Close()
	shared.RunInjection_SelectColumnIdentifier(t, adapt(db), defaultSuite)
}

func TestSQLite_Injection_AliasIdentifier(t *testing.T) {
	t.Parallel()
	db := openDB(t)
	defer db.Close()
	shared.RunInjection_AliasIdentifier(t, adapt(db), defaultSuite)
}

func TestSQLite_Injection_CommentBasedPayloads(t *testing.T) {
	t.Parallel()
	db := openDB(t)
	defer db.Close()
	shared.RunInjection_CommentBasedPayloads(t, adapt(db), defaultSuite)
}

func TestSQLite_Injection_StackedInsertPayload(t *testing.T) {
	t.Parallel()
	db := openDB(t)
	defer db.Close()
	shared.RunInjection_StackedInsertPayload(t, adapt(db), defaultSuite)
}

func TestSQLite_JoinWithAliases(t *testing.T) {
	t.Parallel()
	db := openDB(t)
	defer db.Close()
	shared.RunJoinWithAliases(t, adapt(db), defaultSuite)
}

func TestSQLite_Subquery(t *testing.T) {
	t.Parallel()
	db := openDB(t)
	defer db.Close()
	shared.RunSubquery(t, adapt(db), defaultSuite)
}

func TestSQLite_SetOperations(t *testing.T) {
	t.Parallel()
	db := openDB(t)
	defer db.Close()
	shared.RunSetOperations(t, adapt(db), defaultSuite)
}

func TestSQLite_CTEAndAggregation(t *testing.T) {
	t.Parallel()
	db := openDB(t)
	defer db.Close()
	shared.RunCTEAndAggregation(t, adapt(db), defaultSuite)
}

func TestSQLite_Identifiers_ReservedKeywords(t *testing.T) {
	t.Parallel()
	db := openDB(t)
	defer db.Close()
	shared.RunIdentifiers_ReservedKeywords(t, adapt(db), defaultSuite)
}

func TestSQLite_Identifiers_MixedCase(t *testing.T) {
	t.Parallel()
	db := openDB(t)
	defer db.Close()
	shared.RunIdentifiers_MixedCase(t, adapt(db), defaultSuite)
}

func TestSQLite_Identifiers_EmbeddedDoubleQuote(t *testing.T) {
	t.Parallel()
	db := openDB(t)
	defer db.Close()
	shared.RunIdentifiers_EmbeddedDoubleQuote(t, adapt(db), defaultSuite)
}

func TestSQLite_CreateTable(t *testing.T) {
	t.Parallel()
	db := openDB(t)
	defer db.Close()
	shared.RunCreateTable(t, adapt(db), defaultSuite)
}

func TestSQLite_CreateTable_UniqueConstraint(t *testing.T) {
	t.Parallel()
	db := openDB(t)
	defer db.Close()
	shared.RunCreateTable_UniqueConstraint(t, adapt(db), defaultSuite)
}

func TestSQLite_CreateTable_ForeignKey(t *testing.T) {
	t.Parallel()
	db := openDB(t)
	defer db.Close()
	shared.RunCreateTable_ForeignKey(t, adapt(db), defaultSuite)
}

func TestSQLite_CreateTable_Check(t *testing.T) {
	t.Parallel()
	db := openDB(t)
	defer db.Close()
	shared.RunCreateTable_Check(t, adapt(db), defaultSuite)
}

func TestSQLite_CreateIndex(t *testing.T) {
	t.Parallel()
	db := openDB(t)
	defer db.Close()
	shared.RunCreateIndex(t, adapt(db), defaultSuite)
}

func TestSQLite_CreateIndex_Partial(t *testing.T) {
	t.Parallel()
	db := openDB(t)
	defer db.Close()
	shared.RunCreateIndex_Partial(t, adapt(db), defaultSuite)
}

func TestSQLite_FilteringAndPagination(t *testing.T) {
	t.Parallel()
	db := openDB(t)
	defer db.Close()
	shared.RunFilteringAndPagination(t, adapt(db), defaultSuite)
}

func TestSQLite_RecursiveCTE(t *testing.T) {
	t.Parallel()
	db := openDB(t)
	defer db.Close()
	shared.RunRecursiveCTE(t, adapt(db), defaultSuite)
}

func TestSQLite_SchemaEvolution(t *testing.T) {
	t.Parallel()
	db := openDB(t)
	defer db.Close()
	shared.RunSchemaEvolution(t, adapt(db), defaultSuite)
}

func TestSQLite_ReturningAndOrConflict(t *testing.T) {
	t.Parallel()
	db := openDB(t)
	defer db.Close()

	ctx := context.Background()
	a := adapt(db)

	if err := a.Exec(ctx, `
		CREATE TEMPORARY TABLE config (
			key   TEXT NOT NULL PRIMARY KEY,
			value TEXT NOT NULL
		)
	`); err != nil {
		t.Fatalf("create table: %v", err)
	}

	// Seed an initial row.
	if err := a.Exec(ctx,
		defaultSuite.InsertInto("config").
			Columns("key", "value").
			Values(sqblsqlite.P(1), sqblsqlite.P(2)).
			ToSql(),
		"theme", "light",
	); err != nil {
		t.Fatalf("seed: %v", err)
	}

	// INSERT OR IGNORE: existing key must not overwrite the existing value.
	if err := a.Exec(ctx,
		defaultSuite.InsertInto("config").
			Columns("key", "value").
			Values(sqblsqlite.P(1), sqblsqlite.P(2)).
			OrIgnore().
			ToSql(),
		"theme", "dark",
	); err != nil {
		t.Fatalf("or ignore: %v", err)
	}
	var val string
	if err := a.QueryRow(ctx,
		defaultSuite.From("config").Select("value").Where(sqblsqlite.Eq("key", sqblsqlite.P(1))).ToSql(),
		"theme",
	).Scan(&val); err != nil {
		t.Fatalf("select after ignore: %v", err)
	}
	if val != "light" {
		t.Errorf("OR IGNORE: value = %q; want %q", val, "light")
	}

	// INSERT OR REPLACE: existing key must be replaced with the new value.
	if err := a.Exec(ctx,
		defaultSuite.InsertInto("config").
			Columns("key", "value").
			Values(sqblsqlite.P(1), sqblsqlite.P(2)).
			OrReplace().
			ToSql(),
		"theme", "dark",
	); err != nil {
		t.Fatalf("or replace: %v", err)
	}
	if err := a.QueryRow(ctx,
		defaultSuite.From("config").Select("value").Where(sqblsqlite.Eq("key", sqblsqlite.P(1))).ToSql(),
		"theme",
	).Scan(&val); err != nil {
		t.Fatalf("select after replace: %v", err)
	}
	if val != "dark" {
		t.Errorf("OR REPLACE: value = %q; want %q", val, "dark")
	}

	// INSERT ... RETURNING key, value — new row, get the inserted values back.
	var retKey, retVal string
	if err := a.QueryRow(ctx,
		defaultSuite.InsertInto("config").
			Columns("key", "value").
			Values(sqblsqlite.P(1), sqblsqlite.P(2)).
			Returning("key", "value").
			ToSql(),
		"lang", "go",
	).Scan(&retKey, &retVal); err != nil {
		t.Fatalf("insert returning: %v", err)
	}
	if retKey != "lang" || retVal != "go" {
		t.Errorf("RETURNING: got (%q, %q); want (%q, %q)", retKey, retVal, "lang", "go")
	}
}
