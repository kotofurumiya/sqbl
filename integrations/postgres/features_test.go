package postgres_test

import (
	"context"
	"strings"
	"testing"

	"github.com/kotofurumiya/sqbl/integrations/shared"
	sqbl "github.com/kotofurumiya/sqbl/sqblpg"
	"github.com/kotofurumiya/sqbl/syntax"
)

// seedUsersAndOrders creates users 1-3 (alice, bob, carol) and orders where
// alice has amounts 100 then 300, bob has 200, and carol has none.
func seedUsersAndOrders(t *testing.T, conn shared.Conn) {
	t.Helper()
	shared.MustExec(t, conn, `CREATE TEMPORARY TABLE users (id BIGINT PRIMARY KEY, name TEXT NOT NULL)`)
	shared.MustExec(t, conn, `
		CREATE TEMPORARY TABLE orders (
			id      BIGINT  PRIMARY KEY,
			user_id BIGINT  NOT NULL,
			amount  INTEGER NOT NULL
		)
	`)
	shared.MustExec(t, conn,
		sqbl.InsertInto("users").Columns("id", "name").
			Values(sqbl.P(1), sqbl.P(2)).
			Values(sqbl.P(3), sqbl.P(4)).
			Values(sqbl.P(5), sqbl.P(6)).
			ToSql(),
		1, "alice", 2, "bob", 3, "carol")
	shared.MustExec(t, conn,
		sqbl.InsertInto("orders").Columns("id", "user_id", "amount").
			Values(sqbl.P(1), sqbl.P(2), sqbl.P(3)).
			Values(sqbl.P(4), sqbl.P(5), sqbl.P(6)).
			Values(sqbl.P(7), sqbl.P(8), sqbl.P(9)).
			ToSql(),
		1, 1, 100, 2, 1, 300, 3, 2, 200)
}

type nameAmount struct {
	name   string
	amount int
}

func scanNameAmount(r shared.Rows) (nameAmount, error) {
	var v nameAmount
	return v, r.Scan(&v.name, &v.amount)
}

// TestDistinctOn keeps one row per user: the first after ORDER BY, which here
// is each user's largest order.
func TestDistinctOn(t *testing.T) {
	t.Parallel()
	conn := openConn(t)
	seedUsersAndOrders(t, conn)

	sql := sqbl.From(sqbl.As("orders", "o")).
		Join(sqbl.As("users", "u"), sqbl.Eq("u.id", "o.user_id")).
		DistinctOn("o.user_id").
		Select("u.name", "o.amount").
		OrderBy(sqbl.Asc("o.user_id"), sqbl.Desc("o.amount")).
		ToSql()
	shared.AssertSlice(t, shared.QueryRows(t, conn, scanNameAmount, sql),
		[]nameAmount{{"alice", 300}, {"bob", 200}})
}

// TestILike matches case-insensitively, unlike LIKE.
func TestILike(t *testing.T) {
	t.Parallel()
	conn := openConn(t)
	seedUsersAndOrders(t, conn)

	query := func(cond syntax.SqlFragment) []string {
		return shared.QueryColumn[string](t, conn, sqbl.From("users").Select("name").Where(cond).ToSql(), "ALI%")
	}
	shared.AssertSlice(t, query(sqbl.Like("name", sqbl.P(1))), nil)
	shared.AssertSlice(t, query(sqbl.ILike("name", sqbl.P(1))), []string{"alice"})
}

// TestUnnestArrayParameter joins a Go slice bound as one array parameter.
// The CAST is required because PostgreSQL cannot infer the type of a bare
// parameter passed to unnest, and AsColumns names the produced column.
func TestUnnestArrayParameter(t *testing.T) {
	t.Parallel()
	conn := openConn(t)
	seedUsersAndOrders(t, conn)

	ids := sqbl.AsColumns(sqbl.Fn("unnest", sqbl.Cast(sqbl.P(1), "bigint[]")), "ids", "id")
	sql := sqbl.From("users").
		Join(ids, sqbl.Eq("users.id", "ids.id")).
		Select("users.name").
		OrderBy(sqbl.Asc("users.id")).
		ToSql()
	shared.AssertSlice(t, shared.QueryColumn[string](t, conn, sql, []int64{1, 3}), []string{"alice", "carol"})
}

// TestLateralJoins fetches each user's latest order through a correlated
// subquery; the LEFT form keeps users without orders, the CROSS form drops them.
func TestLateralJoins(t *testing.T) {
	t.Parallel()
	conn := openConn(t)
	seedUsersAndOrders(t, conn)

	latest := sqbl.From("orders").
		Select("amount").
		Where(sqbl.Eq("orders.user_id", "u.id")).
		OrderBy(sqbl.Desc("orders.id")).
		Limit(1)
	base := sqbl.From(sqbl.As("users", "u")).
		Select("u.name", sqbl.As(sqbl.Fn("COALESCE", "latest.amount", 0), "amount")).
		OrderBy(sqbl.Asc("u.id"))

	t.Run("Left", func(t *testing.T) {
		sql := base.LeftLateralJoin(latest, "latest").ToSql()
		shared.AssertSlice(t, shared.QueryRows(t, conn, scanNameAmount, sql),
			[]nameAmount{{"alice", 300}, {"bob", 200}, {"carol", 0}})
	})
	t.Run("Cross", func(t *testing.T) {
		sql := base.CrossLateralJoin(latest, "latest").ToSql()
		shared.AssertSlice(t, shared.QueryRows(t, conn, scanNameAmount, sql),
			[]nameAmount{{"alice", 300}, {"bob", 200}})
	})
}

// TestRowLocking runs the job-queue dequeue pattern across two connections.
// Locks are only visible across sessions, so the table is persistent.
func TestRowLocking(t *testing.T) {
	t.Parallel()
	conn1, conn2 := connectDB(t), connectDB(t)
	shared.UseTables(t, pgxConn{conn1}, "it_jobs")
	shared.MustExec(t, pgxConn{conn1}, `CREATE TABLE it_jobs (id INTEGER PRIMARY KEY)`)
	shared.MustExec(t, pgxConn{conn1}, `INSERT INTO it_jobs (id) VALUES (1), (2)`)

	next := sqbl.From("it_jobs").Select("id").OrderBy(sqbl.Asc("id")).Limit(1)

	// Worker 1 holds a lock on job 1 until its transaction ends.
	tx1, err := conn1.Begin(t.Context())
	if err != nil {
		t.Fatalf("begin tx1: %v", err)
	}
	t.Cleanup(func() { _ = tx1.Rollback(context.Background()) })
	var locked int
	if err := tx1.QueryRow(t.Context(), next.ForUpdate().ToSql()).Scan(&locked); err != nil || locked != 1 {
		t.Fatalf("worker 1 locked (%d, %v); want job 1", locked, err)
	}

	tests := []struct {
		name    string
		sql     string
		wantID  int
		wantErr bool
	}{
		{"SkipLocked passes over the locked row", next.ForUpdate().SkipLocked().ToSql(), 2, false},
		{"Nowait fails instead of waiting", next.ForUpdate().Nowait().ToSql(), 0, true},
		// FOR UPDATE conflicts with every row-lock mode, so each mode fails
		// under NOWAIT only if it really tried to lock the row.
		{"ForShare conflicts with ForUpdate", next.ForShare().Nowait().ToSql(), 0, true},
		{"ForNoKeyUpdate conflicts with ForUpdate", next.ForNoKeyUpdate().Nowait().ToSql(), 0, true},
		{"ForKeyShare conflicts with ForUpdate", next.ForKeyShare().Nowait().ToSql(), 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx2, err := conn2.Begin(t.Context())
			if err != nil {
				t.Fatalf("begin tx2: %v", err)
			}
			defer func() { _ = tx2.Rollback(t.Context()) }()
			var id int
			err = tx2.QueryRow(t.Context(), tt.sql).Scan(&id)
			if gotErr := err != nil; gotErr != tt.wantErr || id != tt.wantID {
				t.Errorf("got (id=%d, err=%v); want id=%d, error=%v", id, err, tt.wantID, tt.wantErr)
			}
		})
	}
}

// TestCreateIndexUsing builds an index with a non-default access method.
func TestCreateIndexUsing(t *testing.T) {
	t.Parallel()
	conn := openConn(t)
	shared.UseTables(t, conn, "it_idx_using")
	shared.MustExec(t, conn, `CREATE TABLE it_idx_using (id BIGINT PRIMARY KEY, email TEXT NOT NULL)`)
	shared.MustExec(t, conn,
		sqbl.CreateIndex("it_idx_using_email").On("it_idx_using").Columns("email").Using("hash").ToSql())

	defs := shared.QueryColumn[string](t, conn,
		`SELECT indexdef FROM pg_indexes WHERE indexname = 'it_idx_using_email'`)
	if len(defs) != 1 || !strings.Contains(defs[0], "USING hash") {
		t.Errorf("indexdef = %q; want one definition USING hash", defs)
	}
}

// TestDropTableCascade drops a table that another table references, which
// only CASCADE allows; the referencing table survives without its foreign key.
func TestDropTableCascade(t *testing.T) {
	t.Parallel()
	conn := openConn(t)
	shared.UseTables(t, conn, "it_cascade_child", "it_cascade_parent")
	shared.MustExec(t, conn, `CREATE TABLE it_cascade_parent (id BIGINT PRIMARY KEY)`)
	shared.MustExec(t, conn, `CREATE TABLE it_cascade_child (parent_id BIGINT REFERENCES it_cascade_parent (id))`)

	shared.MustFail(t, conn, "referenced table without CASCADE", sqbl.DropTable("it_cascade_parent").ToSql())
	shared.MustExec(t, conn, sqbl.DropTable("it_cascade_parent").Cascade().ToSql())
	shared.MustExec(t, conn, `INSERT INTO it_cascade_child (parent_id) VALUES (999)`)
}

// TestCreateAndDropDatabase runs the lifecycle of a throwaway database.
func TestCreateAndDropDatabase(t *testing.T) {
	t.Parallel()
	conn := openConn(t)
	const name = "sqbl_it_throwaway"
	count := func() []int {
		return shared.QueryColumn[int](t, conn, `SELECT COUNT(*) FROM pg_database WHERE datname = $1`, name)
	}

	shared.MustExec(t, conn, sqbl.DropDatabase(name).IfExists().ToSql())
	shared.MustExec(t, conn, sqbl.CreateDatabase(name).ToSql())
	shared.AssertSlice(t, count(), []int{1})
	shared.MustExec(t, conn, sqbl.DropDatabase(name).IfExists().Force().ToSql())
	shared.AssertSlice(t, count(), []int{0})
}
