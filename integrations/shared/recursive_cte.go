package shared

import (
	"context"
	"sort"
	"testing"

	"github.com/kotofurumiya/sqbl/syntax"
)

// RunRecursiveCTE verifies WITH RECURSIVE using an employee hierarchy scenario.
// The test builds the org chart, then retrieves all direct and indirect reports
// under a specific manager and checks the resulting ID set.
func RunRecursiveCTE(t testing.TB, conn Conn, s Suite) {
	t.Helper()
	ctx := context.Background()

	mustExec(t, conn, ctx, "create employees", `
		CREATE TEMPORARY TABLE employees (
			id         INTEGER NOT NULL,
			name       TEXT    NOT NULL,
			manager_id INTEGER
		)
	`)

	// Org chart:
	//   Alice (1) — CEO, no manager
	//     Bob (2) — reports to Alice
	//       Dave (4) — reports to Bob
	//         Eve (5) — reports to Dave
	//     Carol (3) — reports to Alice
	mustExec(t, conn, ctx, "insert employees",
		s.InsertInto("employees").
			Columns("id", "name", "manager_id").
			Values(syntax.P(1), syntax.P(2), syntax.P(3)).
			Values(syntax.P(4), syntax.P(5), syntax.P(6)).
			Values(syntax.P(7), syntax.P(8), syntax.P(9)).
			Values(syntax.P(10), syntax.P(11), syntax.P(12)).
			Values(syntax.P(13), syntax.P(14), syntax.P(15)).
			ToSql(),
		1, "Alice", nil,
		2, "Bob", 1,
		3, "Carol", 1,
		4, "Dave", 2,
		5, "Eve", 4,
	)

	// WITH RECURSIVE subordinates AS (
	//   -- anchor: direct reports of the given manager
	//   SELECT id FROM employees WHERE manager_id = $1
	//   UNION ALL
	//   -- recursive: reports of those reports
	//   SELECT e.id FROM employees AS e
	//     INNER JOIN subordinates ON e.manager_id = subordinates.id
	// )
	// SELECT id FROM subordinates ORDER BY id ASC
	//
	// For manager_id = 2 (Bob): Dave (4), Eve (5).
	anchor := s.From("employees").
		Select("id").
		Where(syntax.Eq("manager_id", syntax.P(1)))

	recursive := s.From(syntax.As("employees", "e")).
		Select("e.id").
		InnerJoin("subordinates", syntax.Eq("e.manager_id", "subordinates.id"))

	recSQL := s.From("subordinates").
		Select("id").
		WithRecursive("subordinates", anchor.UnionAll(recursive)).
		OrderBy(syntax.Asc("id")).
		ToSql()

	rows := mustQuery(t, conn, ctx, "recursive cte", recSQL, 2)
	got := collectRows(t, rows, func(r Rows) (int, error) {
		var id int
		return id, r.Scan(&id)
	})

	sort.Ints(got)
	assertSlice(t, got, []int{4, 5})
}
