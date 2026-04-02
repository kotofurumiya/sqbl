package shared

import (
	"context"
	"testing"

	"github.com/kotofurumiya/sqbl/syntax"
)

// RunCTEAndAggregation verifies WITH (CTE), GROUP BY, and HAVING using a
// monthly sales report scenario.
// A CTE first narrows the data to January 2025, then the outer query
// aggregates per-rep totals and filters with HAVING.
// Explicit IDs are provided in INSERTs to avoid reliance on auto-increment.
func RunCTEAndAggregation(t testing.TB, conn Conn, s Suite) {
	t.Helper()
	ctx := context.Background()

	mustExec(t, conn, ctx, "create table", `
		CREATE TEMPORARY TABLE sales (
			id       INTEGER PRIMARY KEY,
			rep_name TEXT    NOT NULL,
			amount   INTEGER NOT NULL,
			sold_at  DATE    NOT NULL
		)
	`)

	// Seed data spans three months:
	//   - Alice and Bob have January records that should pass the CTE filter.
	//   - Carol's only record is in December and must be excluded by the CTE.
	//   - Bob's February record must also be excluded by the CTE.
	seedSQL := s.InsertInto("sales").
		Columns("id", "rep_name", "amount", "sold_at").
		Values(syntax.P(1), syntax.P(2), syntax.P(3), syntax.P(4)).
		Values(syntax.P(5), syntax.P(6), syntax.P(7), syntax.P(8)).
		Values(syntax.P(9), syntax.P(10), syntax.P(11), syntax.P(12)).
		Values(syntax.P(13), syntax.P(14), syntax.P(15), syntax.P(16)).
		Values(syntax.P(17), syntax.P(18), syntax.P(19), syntax.P(20)).
		ToSql()
	mustExec(t, conn, ctx, "insert", seedSQL,
		1, "Alice", 5000, "2025-01-10",
		2, "Alice", 3000, "2025-01-20",
		3, "Bob", 8000, "2025-01-15",
		4, "Carol", 1000, "2024-12-20", // prior month: excluded by CTE
		5, "Bob", 2000, "2025-02-01",   // next month: excluded by CTE
	)

	// CTE: select only January 2025 rows using BETWEEN for the date range.
	jan := s.From("sales").
		Select("rep_name", "amount").
		Where(syntax.Between("sold_at", syntax.P(1), syntax.P(2)))

	// Outer query: sum amounts per rep from the CTE, keep only those whose
	// January total is at least 5000 (HAVING).
	// Parameters $1/$2 belong to the CTE's BETWEEN; $3 belongs to HAVING.
	cteSQL := s.From("jan_sales").
		With("jan_sales", jan).
		Select("rep_name", syntax.As(syntax.Fn("SUM", "amount"), "total")).
		GroupBy("rep_name").
		Having(syntax.Gte(syntax.Fn("SUM", "amount"), syntax.P(3))).
		OrderBy(syntax.Desc("total"), syntax.Asc("rep_name")).
		ToSql()

	type repTotal struct {
		name  string
		total int
	}
	cteRows := mustQuery(t, conn, ctx, "cte query", cteSQL, "2025-01-01", "2025-01-31", 5000)
	got := collectRows(t, cteRows, func(r Rows) (repTotal, error) {
		var v repTotal
		return v, r.Scan(&v.name, &v.total)
	})

	// Bob: 8000; Alice: 5000+3000=8000. Both exceed the HAVING threshold.
	// Ordered by total DESC then rep_name ASC; totals are equal so Alice
	// comes before Bob alphabetically.
	assertSlice(t, got, []repTotal{{"Alice", 8000}, {"Bob", 8000}})
}
