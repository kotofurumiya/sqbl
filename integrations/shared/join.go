package shared

import (
	"context"
	"testing"

	"github.com/kotofurumiya/sqbl/syntax"
)

// RunJoinWithAliases verifies INNER JOIN and LEFT JOIN using a
// customers-and-orders scenario where one customer (Carol) has no orders.
// Explicit IDs are provided in INSERTs to avoid reliance on auto-increment.
func RunJoinWithAliases(t testing.TB, conn Conn, s Suite) {
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

	// Alice (id=1) has two orders; Bob (id=2) has one; Carol (id=3) has none.
	mustExec(t, conn, ctx, "insert orders",
		s.InsertInto("orders").
			Columns("id", "customer_id", "amount").
			Values(syntax.P(1), syntax.P(2), syntax.P(3)).
			Values(syntax.P(4), syntax.P(5), syntax.P(6)).
			Values(syntax.P(7), syntax.P(8), syntax.P(9)).
			ToSql(),
		1, 1, 1000, 2, 1, 2000, 3, 2, 500,
	)

	// INNER JOIN: only customers who have at least one order should appear.
	// Carol must not appear; Alice appears twice (one row per order).
	innerSQL := s.From(syntax.As("customers", "c")).
		Select("c.name", "o.amount").
		InnerJoin(
			syntax.As("orders", "o"),
			syntax.Eq("c.id", "o.customer_id"),
		).
		OrderBy(syntax.Asc("c.name"), syntax.Asc("o.amount")).
		ToSql()

	type nameAmount struct {
		name   string
		amount int
	}
	innerRows := mustQuery(t, conn, ctx, "inner join", innerSQL)
	innerGot := collectRows(t, innerRows, func(r Rows) (nameAmount, error) {
		var v nameAmount
		return v, r.Scan(&v.name, &v.amount)
	})
	assertSlice(t, innerGot, []nameAmount{{"Alice", 1000}, {"Alice", 2000}, {"Bob", 500}})

	// LEFT JOIN + GROUP BY: all customers appear, Carol with count=0.
	// COUNT(o.id) returns 0 instead of NULL for unmatched rows because
	// COUNT ignores NULLs and o.id is NULL for Carol after the outer join.
	leftSQL := s.From(syntax.As("customers", "c")).
		Select("c.name", syntax.As(syntax.Fn("COUNT", "o.id"), "order_count")).
		LeftJoin(
			syntax.As("orders", "o"),
			syntax.Eq("c.id", "o.customer_id"),
		).
		GroupBy("c.id", "c.name").
		OrderBy(syntax.Asc("c.name")).
		ToSql()

	type nameCount struct {
		name  string
		count int
	}
	leftRows := mustQuery(t, conn, ctx, "left join", leftSQL)
	leftGot := collectRows(t, leftRows, func(r Rows) (nameCount, error) {
		var v nameCount
		return v, r.Scan(&v.name, &v.count)
	})
	assertSlice(t, leftGot, []nameCount{{"Alice", 2}, {"Bob", 1}, {"Carol", 0}})
}
