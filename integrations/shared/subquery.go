package shared

import (
	"context"
	"testing"

	"github.com/kotofurumiya/sqbl/syntax"
)

// RunSubquery verifies that a SqlSelectBuilder can be embedded as a subquery
// in both the FROM clause and a JOIN clause.
// Explicit IDs are provided in INSERTs to avoid reliance on auto-increment.
func RunSubquery(t testing.TB, conn Conn, s Suite) {
	t.Helper()
	ctx := context.Background()

	mustExec(t, conn, ctx, "create customers", `
		CREATE TEMPORARY TABLE customers (
			id   INTEGER PRIMARY KEY,
			name TEXT NOT NULL
		)
	`)
	mustExec(t, conn, ctx, "create orders", `
		CREATE TEMPORARY TABLE orders (
			id          INTEGER PRIMARY KEY,
			customer_id INTEGER NOT NULL,
			amount      INTEGER NOT NULL
		)
	`)

	mustExec(t, conn, ctx, "insert customers",
		s.InsertInto("customers").
			Columns("id", "name").
			Values(syntax.P(1), syntax.P(2)).
			Values(syntax.P(3), syntax.P(4)).
			Values(syntax.P(5), syntax.P(6)).
			ToSql(),
		1, "Alice", 2, "Bob", 3, "Carol",
	)

	// Alice total: 3500, Bob total: 800, Carol: no orders.
	mustExec(t, conn, ctx, "insert orders",
		s.InsertInto("orders").
			Columns("id", "customer_id", "amount").
			Values(syntax.P(1), syntax.P(2), syntax.P(3)).
			Values(syntax.P(4), syntax.P(5), syntax.P(6)).
			Values(syntax.P(7), syntax.P(8), syntax.P(9)).
			ToSql(),
		1, 1, 1000, 2, 1, 2500, 3, 2, 800,
	)

	// Subquery: per-customer order totals. Reused as both a FROM and JOIN subquery.
	totals := s.From("orders").
		Select("customer_id", syntax.As(syntax.Fn("SUM", "amount"), "total")).
		GroupBy("customer_id")

	// FROM subquery: wrap the aggregation as a derived table and filter by total.
	// Only Alice (3500) exceeds the threshold of 1000.
	fromSubSQL := s.From(syntax.As(totals, "t")).
		Select("t.customer_id", "t.total").
		Where(syntax.Gt("t.total", syntax.P(1))).
		OrderBy(syntax.Desc("t.total")).
		ToSql()

	var custID, total int
	if err := conn.QueryRow(ctx, fromSubSQL, 1000).Scan(&custID, &total); err != nil {
		t.Fatalf("from subquery: %v", err)
	}
	if custID != 1 || total != 3500 {
		t.Errorf("from subquery = (customer_id=%d, total=%d); want (1, 3500)", custID, total)
	}

	// JOIN subquery: LEFT JOIN customers against the same aggregation so that
	// Carol, who has no orders, receives a total of 0 via COALESCE.
	joinSubSQL := s.From(syntax.As("customers", "c")).
		Select("c.name", syntax.As(syntax.Fn("COALESCE", "s.total", 0), "total")).
		LeftJoin(
			syntax.As(totals, "s"),
			syntax.Eq("c.id", "s.customer_id"),
		).
		OrderBy(syntax.Desc("total"), syntax.Asc("c.name")).
		ToSql()

	type nameTotal struct {
		name  string
		total int
	}
	joinRows := mustQuery(t, conn, ctx, "join subquery", joinSubSQL)
	joinGot := collectRows(t, joinRows, func(r Rows) (nameTotal, error) {
		var v nameTotal
		return v, r.Scan(&v.name, &v.total)
	})
	assertSlice(t, joinGot, []nameTotal{{"Alice", 3500}, {"Bob", 800}, {"Carol", 0}})
}
