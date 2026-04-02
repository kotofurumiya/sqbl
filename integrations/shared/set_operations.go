package shared

import (
	"context"
	"testing"

	"github.com/kotofurumiya/sqbl/syntax"
)

// RunSetOperations verifies UNION and EXCEPT using two membership tables.
// The test confirms that UNION deduplicates overlapping IDs and that EXCEPT
// returns only rows present in the first operand but not the second.
func RunSetOperations(t testing.TB, conn Conn, s Suite) {
	t.Helper()
	ctx := context.Background()

	mustExec(t, conn, ctx, "create premium_members", `
		CREATE TEMPORARY TABLE premium_members (user_id INTEGER NOT NULL)
	`)
	mustExec(t, conn, ctx, "create newsletter_subscribers", `
		CREATE TEMPORARY TABLE newsletter_subscribers (user_id INTEGER NOT NULL)
	`)

	// premium: {1, 2, 3}
	mustExec(t, conn, ctx, "insert premium",
		s.InsertInto("premium_members").
			Columns("user_id").
			Values(syntax.P(1)).
			Values(syntax.P(2)).
			Values(syntax.P(3)).
			ToSql(),
		1, 2, 3,
	)

	// newsletter: {2, 3, 4, 5} — overlaps with premium on 2 and 3.
	mustExec(t, conn, ctx, "insert newsletter",
		s.InsertInto("newsletter_subscribers").
			Columns("user_id").
			Values(syntax.P(1)).
			Values(syntax.P(2)).
			Values(syntax.P(3)).
			Values(syntax.P(4)).
			ToSql(),
		2, 3, 4, 5,
	)

	// UNION deduplicates: {1,2,3} ∪ {2,3,4,5} = {1,2,3,4,5} (5 distinct IDs).
	// The result order from UNION is unspecified, so we collect into a set and
	// verify membership rather than checking positional equality.
	unionSQL := s.From("premium_members").Select("user_id").
		Union(s.From("newsletter_subscribers").Select("user_id")).
		ToSql()

	unionRows := mustQuery(t, conn, ctx, "union", unionSQL)
	defer unionRows.Close()
	unionSet := make(map[int]bool)
	for unionRows.Next() {
		var id int
		if err := unionRows.Scan(&id); err != nil {
			t.Fatalf("scan: %v", err)
		}
		unionSet[id] = true
	}
	if err := unionRows.Err(); err != nil {
		t.Fatalf("union rows err: %v", err)
	}

	if len(unionSet) != 5 {
		t.Errorf("union count = %d; want 5", len(unionSet))
	}
	for _, id := range []int{1, 2, 3, 4, 5} {
		if !unionSet[id] {
			t.Errorf("union missing user_id %d", id)
		}
	}

	// EXCEPT: {1,2,3} − {2,3,4,5} = {1}.
	// User 1 is the only premium member not subscribed to the newsletter.
	exceptSQL := s.From("premium_members").Select("user_id").
		Except(s.From("newsletter_subscribers").Select("user_id")).
		ToSql()

	var onlyPremium int
	if err := conn.QueryRow(ctx, exceptSQL).Scan(&onlyPremium); err != nil {
		t.Fatalf("except: %v", err)
	}
	if onlyPremium != 1 {
		t.Errorf("except user_id = %d; want 1", onlyPremium)
	}

	// UNION ALL preserves duplicates: {1,2,3} ∪ₐ {2,3,4,5} = 7 rows.
	// Users 2 and 3 appear twice; verify the count is 7, not 5.
	unionAllSQL := s.From("premium_members").Select("user_id").
		UnionAll(s.From("newsletter_subscribers").Select("user_id")).
		ToSql()

	unionAllRows := mustQuery(t, conn, ctx, "union all", unionAllSQL)
	unionAllGot := collectRows(t, unionAllRows, func(r Rows) (int, error) {
		var id int
		return id, r.Scan(&id)
	})

	if len(unionAllGot) != 7 {
		t.Errorf("UNION ALL count = %d; want 7 (duplicates preserved)", len(unionAllGot))
	}

	// Users 2 and 3 each appear twice; confirm by counting occurrences.
	dupCounts := make(map[int]int)
	for _, id := range unionAllGot {
		dupCounts[id]++
	}
	for _, dup := range []int{2, 3} {
		if dupCounts[dup] != 2 {
			t.Errorf("UNION ALL: user_id %d appears %d time(s); want 2", dup, dupCounts[dup])
		}
	}
}
