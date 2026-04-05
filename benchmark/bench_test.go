package benchmark_test

import (
	"testing"

	"github.com/kotofurumiya/sqbl/sqblpg"
)

// BenchmarkSimple: FROM only, equivalent to SELECT *.
// No slice operations. Baseline for the minimum library overhead.
func BenchmarkSimple(b *testing.B) {
	for b.Loop() {
		sqblpg.From("users").ToSql()
	}
}

// BenchmarkNoSliceCopy: WHERE + LIMIT only, using methods that never touch slices.
// From (SqlFragment pointer), Where (SqlFragment pointer), and Limit (*int) are all scalar assignments.
// Eq produces a ComparisonExpr with no internal slice.
func BenchmarkNoSliceCopy(b *testing.B) {
	for b.Loop() {
		sqblpg.From("users").
			Where(sqblpg.Eq("id", sqblpg.P(1))).
			Limit(1).
			ToSql()
	}
}

// BenchmarkWithSliceCopy: SELECT columns + 2 JOINs + WHERE + ORDER BY.
// Multiple slice copies occur during builder construction:
//   - Select(...)    → make([]T, 3) + loop
//   - InnerJoin(...) → append(append(nil, old...), new)
//   - LeftJoin(...)  → append(append(nil, old...), new)
//   - OrderBy(...)   → make([]T, 1) + loop
func BenchmarkWithSliceCopy(b *testing.B) {
	for b.Loop() {
		sqblpg.From(sqblpg.As("users", "u")).
			Select("u.id", "u.name", "o.amount").
			InnerJoin(sqblpg.As("orders", "o"), sqblpg.Eq("u.id", "o.user_id")).
			LeftJoin(sqblpg.As("products", "p"), sqblpg.Eq("o.product_id", "p.id")).
			Where(sqblpg.Eq("u.active", true)).
			OrderBy(sqblpg.Asc("u.name")).
			ToSql()
	}
}

// BenchmarkComplex: a realistic query exercising all major features.
// CTE + 2 JOINs + AND(3 conditions) + GROUP BY + HAVING + ORDER BY + LIMIT.
// Builder phase: slice copies in With, Select, InnerJoin, LeftJoin, GroupBy, OrderBy.
// Render phase: nested CTE subquery, multiple JOIN ON clauses, GROUP BY/HAVING string joins.
func BenchmarkComplex(b *testing.B) {
	for b.Loop() {
		activeUsers := sqblpg.From("users").
			Select("id").
			Where(sqblpg.Eq("active", true))

		sqblpg.From(sqblpg.As("orders", "o")).
			With("active_users", activeUsers).
			Select(
				"o.user_id",
				sqblpg.As(sqblpg.Fn("COUNT", "*"), "order_count"),
				sqblpg.As(sqblpg.Fn("SUM", "o.amount"), "total"),
			).
			InnerJoin(sqblpg.As("active_users", "au"), sqblpg.Eq("o.user_id", "au.id")).
			LeftJoin(sqblpg.As("users", "u"), sqblpg.Eq("o.user_id", "u.id")).
			Where(sqblpg.And(
				sqblpg.Gte("o.created_at", sqblpg.P(1)),
				sqblpg.Lt("o.created_at", sqblpg.P(2)),
				sqblpg.Eq("o.status", sqblpg.P(3)),
			)).
			GroupBy("o.user_id").
			Having(sqblpg.Gt(sqblpg.Fn("COUNT", "*"), 5)).
			OrderBy(sqblpg.Desc("total")).
			Limit(100).
			ToSql()
	}
}
