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

// connectDB opens a connection to the test PostgreSQL instance.
// It reads the DSN from the POSTGRES_DSN environment variable, falling back to
// the default address used by the compose.yaml in this repository.
func connectDB(t *testing.T) *pgx.Conn {
	t.Helper()
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		dsn = "postgres://sqbl:sqbl@localhost:5432/sqbl_test"
	}
	conn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		t.Fatalf("PostgreSQL not available: %v", err)
	}
	return conn
}

// pgxConn adapts *pgx.Conn to the shared.Conn interface.
type pgxConn struct{ conn *pgx.Conn }

// pgxRow wraps pgx.Row so that pgx.ErrNoRows is mapped to shared.ErrNoRows,
// giving shared tests a single canonical sentinel to check with errors.Is.
type pgxRow struct{ row pgx.Row }

func (r pgxRow) Scan(dest ...any) error {
	if err := r.row.Scan(dest...); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return shared.ErrNoRows
		}
		return err
	}
	return nil
}

func (a *pgxConn) Exec(ctx context.Context, sql string, args ...any) error {
	_, err := a.conn.Exec(ctx, sql, args...)
	return err
}

func (a *pgxConn) QueryRow(ctx context.Context, sql string, args ...any) shared.Row {
	return pgxRow{a.conn.QueryRow(ctx, sql, args...)}
}

func (a *pgxConn) Query(ctx context.Context, sql string, args ...any) (shared.Rows, error) {
	return a.conn.Query(ctx, sql, args...)
}

// adapt wraps a *pgx.Conn as a shared.Conn.
func adapt(conn *pgx.Conn) shared.Conn { return &pgxConn{conn} }

// defaultSuite wires the PostgreSQL builder constructors into a shared.Suite.
var defaultSuite = shared.Suite{
	From:        sqblpg.From,
	InsertInto:  sqblpg.InsertInto,
	Update:      sqblpg.Update,
	DeleteFrom:  sqblpg.DeleteFrom,
	CreateTable: sqblpg.CreateTable,
	CreateIndex: sqblpg.CreateIndex,
	DropTable:   sqblpg.DropTable,
	DropIndex:   sqblpg.DropIndex,
	AlterTable:  sqblpg.AlterTable,
}

func TestPostgres_CRUD(t *testing.T) {
	t.Parallel()
	conn := connectDB(t)
	defer conn.Close(context.Background())
	shared.RunCRUD(t, adapt(conn), defaultSuite)
}

func TestPostgres_Injection_ValueAsParameter(t *testing.T) {
	t.Parallel()
	conn := connectDB(t)
	defer conn.Close(context.Background())
	shared.RunInjection_ValueAsParameter(t, adapt(conn), defaultSuite)
}

func TestPostgres_Injection_StoredValue(t *testing.T) {
	t.Parallel()
	conn := connectDB(t)
	defer conn.Close(context.Background())
	shared.RunInjection_StoredValue(t, adapt(conn), defaultSuite)
}

func TestPostgres_Injection_LikeWildcardsInEqValue(t *testing.T) {
	t.Parallel()
	conn := connectDB(t)
	defer conn.Close(context.Background())
	shared.RunInjection_LikeWildcardsInEqValue(t, adapt(conn), defaultSuite)
}

func TestPostgres_Injection_QuotedIdentifier(t *testing.T) {
	t.Parallel()
	conn := connectDB(t)
	defer conn.Close(context.Background())
	shared.RunInjection_QuotedIdentifier(t, adapt(conn), defaultSuite)
}

func TestPostgres_Injection_OrderByIdentifier(t *testing.T) {
	t.Parallel()
	conn := connectDB(t)
	defer conn.Close(context.Background())
	shared.RunInjection_OrderByIdentifier(t, adapt(conn), defaultSuite)
}

func TestPostgres_Injection_TableNameIdentifier(t *testing.T) {
	t.Parallel()
	conn := connectDB(t)
	defer conn.Close(context.Background())
	shared.RunInjection_TableNameIdentifier(t, adapt(conn), defaultSuite)
}

func TestPostgres_Injection_SelectColumnIdentifier(t *testing.T) {
	t.Parallel()
	conn := connectDB(t)
	defer conn.Close(context.Background())
	shared.RunInjection_SelectColumnIdentifier(t, adapt(conn), defaultSuite)
}

func TestPostgres_Injection_AliasIdentifier(t *testing.T) {
	t.Parallel()
	conn := connectDB(t)
	defer conn.Close(context.Background())
	shared.RunInjection_AliasIdentifier(t, adapt(conn), defaultSuite)
}

func TestPostgres_Injection_CommentBasedPayloads(t *testing.T) {
	t.Parallel()
	conn := connectDB(t)
	defer conn.Close(context.Background())
	shared.RunInjection_CommentBasedPayloads(t, adapt(conn), defaultSuite)
}

func TestPostgres_Injection_StackedInsertPayload(t *testing.T) {
	t.Parallel()
	conn := connectDB(t)
	defer conn.Close(context.Background())
	shared.RunInjection_StackedInsertPayload(t, adapt(conn), defaultSuite)
}

func TestPostgres_JoinWithAliases(t *testing.T) {
	t.Parallel()
	conn := connectDB(t)
	defer conn.Close(context.Background())
	shared.RunJoinWithAliases(t, adapt(conn), defaultSuite)
}

func TestPostgres_Subquery(t *testing.T) {
	t.Parallel()
	conn := connectDB(t)
	defer conn.Close(context.Background())
	shared.RunSubquery(t, adapt(conn), defaultSuite)
}

func TestPostgres_SetOperations(t *testing.T) {
	t.Parallel()
	conn := connectDB(t)
	defer conn.Close(context.Background())
	shared.RunSetOperations(t, adapt(conn), defaultSuite)
}

func TestPostgres_CTEAndAggregation(t *testing.T) {
	t.Parallel()
	conn := connectDB(t)
	defer conn.Close(context.Background())
	shared.RunCTEAndAggregation(t, adapt(conn), defaultSuite)
}

func TestPostgres_Identifiers_ReservedKeywords(t *testing.T) {
	t.Parallel()
	conn := connectDB(t)
	defer conn.Close(context.Background())
	shared.RunIdentifiers_ReservedKeywords(t, adapt(conn), defaultSuite)
}

func TestPostgres_Identifiers_MixedCase(t *testing.T) {
	t.Parallel()
	conn := connectDB(t)
	defer conn.Close(context.Background())
	shared.RunIdentifiers_MixedCase(t, adapt(conn), defaultSuite)
}

func TestPostgres_Identifiers_EmbeddedDoubleQuote(t *testing.T) {
	t.Parallel()
	conn := connectDB(t)
	defer conn.Close(context.Background())
	shared.RunIdentifiers_EmbeddedDoubleQuote(t, adapt(conn), defaultSuite)
}

func TestPostgres_FilteringAndPagination(t *testing.T) {
	t.Parallel()
	conn := connectDB(t)
	defer conn.Close(context.Background())
	shared.RunFilteringAndPagination(t, adapt(conn), defaultSuite)
}

func TestPostgres_RecursiveCTE(t *testing.T) {
	t.Parallel()
	conn := connectDB(t)
	defer conn.Close(context.Background())
	shared.RunRecursiveCTE(t, adapt(conn), defaultSuite)
}

func TestPostgres_SchemaEvolution(t *testing.T) {
	t.Parallel()
	conn := connectDB(t)
	defer conn.Close(context.Background())
	shared.RunSchemaEvolution(t, adapt(conn), defaultSuite)
}
