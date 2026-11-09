package postgres_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/kotofurumiya/sqbl/integrations/shared"
	"github.com/kotofurumiya/sqbl/sqblpg"
)

// connectDB opens a dedicated connection to the test PostgreSQL instance.
// POSTGRES_DSN overrides the default, which matches compose.yaml.
func connectDB(t *testing.T) *pgx.Conn {
	t.Helper()
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		dsn = "postgres://sqbl:sqbl@localhost:5432/sqbl_test"
	}
	conn, err := pgx.Connect(t.Context(), dsn)
	if err != nil {
		t.Fatalf("PostgreSQL not available: %v", err)
	}
	// Closing in a cleanup rather than a defer keeps the connection open for
	// cleanups registered later by the test, since cleanups run in reverse.
	// t.Context() is already canceled when cleanups run.
	t.Cleanup(func() {
		if err := conn.Close(context.Background()); err != nil {
			t.Errorf("close conn: %v", err)
		}
	})
	return conn
}

// pgxConn adapts *pgx.Conn to shared.Conn.
type pgxConn struct{ conn *pgx.Conn }

// pgxRow maps pgx.ErrNoRows to shared.ErrNoRows.
type pgxRow struct{ row pgx.Row }

func (r pgxRow) Scan(dest ...any) error {
	err := r.row.Scan(dest...)
	if errors.Is(err, pgx.ErrNoRows) {
		return shared.ErrNoRows
	}
	return err
}

func (a pgxConn) Exec(ctx context.Context, sql string, args ...any) error {
	_, err := a.conn.Exec(ctx, sql, args...)
	return err
}

func (a pgxConn) QueryRow(ctx context.Context, sql string, args ...any) shared.Row {
	return pgxRow{a.conn.QueryRow(ctx, sql, args...)}
}

func (a pgxConn) Query(ctx context.Context, sql string, args ...any) (shared.Rows, error) {
	return a.conn.Query(ctx, sql, args...)
}

func openConn(t *testing.T) shared.Conn { return pgxConn{connectDB(t)} }

var suite = shared.Suite{
	From:        sqblpg.From,
	InsertInto:  sqblpg.InsertInto,
	Update:      sqblpg.Update,
	DeleteFrom:  sqblpg.DeleteFrom,
	CreateTable: sqblpg.CreateTable,
	CreateIndex: sqblpg.CreateIndex,
	DropTable:   sqblpg.DropTable,
	DropIndex:   sqblpg.DropIndex,
	AlterTable:  sqblpg.AlterTable,

	IndexedParams: true,
	FullJoin:      true,
	SetOpAll:      true,
	Returning:     true,
	OnConflict:    true,
	PartialIndex:  true,
}

func TestShared(t *testing.T) {
	t.Parallel()
	shared.RunAll(t, openConn, suite)
}
