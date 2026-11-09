package builder_test

import (
	"testing"

	"github.com/kotofurumiya/sqbl/internal/sqltesting"
	"github.com/kotofurumiya/sqbl/syntax"
)

func TestSqlSelectBuilder_Clauses(t *testing.T) {
	t.Parallel()
	runSQLCases(t, []sqlCase{
		{
			name: "no columns selects *",
			b:    newSelect().From("users"),
			want: `SELECT * FROM "users";`,
		},
		{
			name: "qualified columns",
			b:    newSelect().From("users").Select("users.id", "name", "u.*"),
			want: `SELECT "users"."id", "name", "u".* FROM "users";`,
		},
		{
			name: "select without from",
			b:    newSelect().Select(syntax.Fn("NOW")),
			want: `SELECT NOW();`,
		},
		{
			name: "call order of Select and From does not matter",
			b:    newSelect().Select("id").From("users"),
			want: `SELECT "id" FROM "users";`,
		},
		{
			name: "aliases",
			b:    newSelect().From(syntax.As("users", "u")).Select(syntax.As(syntax.Fn("COUNT", "*"), "cnt")),
			want: `SELECT COUNT(*) AS "cnt" FROM "users" AS "u";`,
		},
		{
			name: "distinct",
			b:    newSelect().From("users").Select("country").Distinct(),
			want: `SELECT DISTINCT "country" FROM "users";`,
		},
		{
			name: "distinct on takes precedence over distinct",
			b:    newSelect().From("orders").Select("user_id", "amount").Distinct().DistinctOn("user_id", "shop_id"),
			want: `SELECT DISTINCT ON ("user_id", "shop_id") "user_id", "amount" FROM "orders";`,
		},
		{
			name: "where",
			b:    newSelect().From("users").Where(syntax.Eq("id", syntax.P(1))),
			want: `SELECT * FROM "users" WHERE "id" = $1;`,
		},
		{
			name: "group by and having",
			b: newSelect().From("orders").
				Select("user_id", "shop_id").
				GroupBy("user_id", syntax.Raw(`"shop_id"`)).
				Having(syntax.Gt(syntax.Fn("SUM", "amount"), 1000)),
			want: `SELECT "user_id", "shop_id" FROM "orders" GROUP BY "user_id", "shop_id" HAVING SUM("amount") > 1000;`,
		},
		{
			name: "order by mixes plain columns and directions",
			b:    newSelect().From("users").OrderBy("id", syntax.Asc("name"), syntax.Desc("created_at")),
			want: `SELECT * FROM "users" ORDER BY "id", "name" ASC, "created_at" DESC;`,
		},
		{
			name: "zero limit and offset are still rendered",
			b:    newSelect().From("users").Limit(0).Offset(0),
			want: `SELECT * FROM "users" LIMIT 0 OFFSET 0;`,
		},
		{
			name: "offset without limit",
			b:    newSelect().From("users").Offset(20),
			want: `SELECT * FROM "users" OFFSET 20;`,
		},
		{
			name: "single-value clauses are replaced by later calls",
			b: newSelect().
				From("old").From("users").
				Select("old").Select("id").
				Where(syntax.Eq("old", 1)).Where(syntax.Eq("id", 1)).
				GroupBy("old").GroupBy("id").
				Having(syntax.Eq("old", 1)).Having(syntax.Gt("id", 0)).
				OrderBy("old").OrderBy("id").
				Limit(1).Limit(10).
				Offset(1).Offset(20),
			want: `SELECT "id" FROM "users" WHERE "id" = 1 GROUP BY "id" HAVING "id" > 0 ORDER BY "id" LIMIT 10 OFFSET 20;`,
		},
	})
}

func TestSqlSelectBuilder_Joins(t *testing.T) {
	t.Parallel()
	users := newSelect().From(syntax.As("users", "u")).Select("u.id")
	on := syntax.Eq("u.id", "o.user_id")
	latest := newSelect().From("orders").Select("amount").Where(syntax.Eq("user_id", "u.id")).Limit(1)

	runSQLCases(t, []sqlCase{
		{
			name: "join is inner join",
			b:    users.Join(syntax.As("orders", "o"), on),
			want: `SELECT "u"."id" FROM "users" AS "u" INNER JOIN "orders" AS "o" ON "u"."id" = "o"."user_id";`,
		},
		{
			name: "inner",
			b:    users.InnerJoin(syntax.As("orders", "o"), on),
			want: `SELECT "u"."id" FROM "users" AS "u" INNER JOIN "orders" AS "o" ON "u"."id" = "o"."user_id";`,
		},
		{
			name: "left",
			b:    users.LeftJoin(syntax.As("orders", "o"), on),
			want: `SELECT "u"."id" FROM "users" AS "u" LEFT JOIN "orders" AS "o" ON "u"."id" = "o"."user_id";`,
		},
		{
			name: "right",
			b:    users.RightJoin(syntax.As("orders", "o"), on),
			want: `SELECT "u"."id" FROM "users" AS "u" RIGHT JOIN "orders" AS "o" ON "u"."id" = "o"."user_id";`,
		},
		{
			name: "full outer",
			b:    users.FullJoin(syntax.As("orders", "o"), on),
			want: `SELECT "u"."id" FROM "users" AS "u" FULL OUTER JOIN "orders" AS "o" ON "u"."id" = "o"."user_id";`,
		},
		{
			name: "cross has no on",
			b:    users.CrossJoin("sizes"),
			want: `SELECT "u"."id" FROM "users" AS "u" CROSS JOIN "sizes";`,
		},
		{
			name: "left lateral joins on true",
			b:    users.LeftLateralJoin(latest, "latest"),
			want: `
				SELECT "u"."id" FROM "users" AS "u"
				LEFT LATERAL JOIN (SELECT "amount" FROM "orders" WHERE "user_id" = "u"."id" LIMIT 1) AS "latest" ON TRUE;`,
		},
		{
			name: "cross lateral has no on",
			b:    users.CrossLateralJoin(latest, "latest"),
			want: `
				SELECT "u"."id" FROM "users" AS "u"
				CROSS LATERAL JOIN (SELECT "amount" FROM "orders" WHERE "user_id" = "u"."id" LIMIT 1) AS "latest";`,
		},
		{
			name: "joins accumulate in call order",
			b: users.
				LeftJoin(syntax.As("orders", "o"), on).
				CrossJoin("sizes").
				InnerJoin(syntax.As("shops", "s"), syntax.Eq("o.shop_id", "s.id")),
			want: `
				SELECT "u"."id" FROM "users" AS "u"
				LEFT JOIN "orders" AS "o" ON "u"."id" = "o"."user_id"
				CROSS JOIN "sizes"
				INNER JOIN "shops" AS "s" ON "o"."shop_id" = "s"."id";`,
		},
	})
}

func TestSqlSelectBuilder_SetOperations(t *testing.T) {
	t.Parallel()
	active := newSelect().From("active_users").Select("id")
	premium := newSelect().From("premium_users").Select("id")

	runSQLCases(t, []sqlCase{
		{"union", active.Union(premium), `SELECT "id" FROM "active_users" UNION SELECT "id" FROM "premium_users";`},
		{"union all", active.UnionAll(premium), `SELECT "id" FROM "active_users" UNION ALL SELECT "id" FROM "premium_users";`},
		{"intersect", active.Intersect(premium), `SELECT "id" FROM "active_users" INTERSECT SELECT "id" FROM "premium_users";`},
		{"intersect all", active.IntersectAll(premium), `SELECT "id" FROM "active_users" INTERSECT ALL SELECT "id" FROM "premium_users";`},
		{"except", active.Except(premium), `SELECT "id" FROM "active_users" EXCEPT SELECT "id" FROM "premium_users";`},
		{"except all", active.ExceptAll(premium), `SELECT "id" FROM "active_users" EXCEPT ALL SELECT "id" FROM "premium_users";`},
		{
			name: "operations chain in call order with their own clauses",
			b: active.
				Union(premium.Where(syntax.Gt("id", syntax.P(1)))).
				Except(newSelect().From("banned_users").Select("id")),
			want: `
				SELECT "id" FROM "active_users"
				UNION SELECT "id" FROM "premium_users" WHERE "id" > $1
				EXCEPT SELECT "id" FROM "banned_users";`,
		},
	})
}

func TestSqlSelectBuilder_Locking(t *testing.T) {
	t.Parallel()
	q := newSelect().From("jobs").Select("id")

	runSQLCases(t, []sqlCase{
		{"for update", q.ForUpdate(), `SELECT "id" FROM "jobs" FOR UPDATE;`},
		{"for no key update", q.ForNoKeyUpdate(), `SELECT "id" FROM "jobs" FOR NO KEY UPDATE;`},
		{"for share", q.ForShare(), `SELECT "id" FROM "jobs" FOR SHARE;`},
		{"for key share", q.ForKeyShare(), `SELECT "id" FROM "jobs" FOR KEY SHARE;`},
		{"nowait", q.ForUpdate().Nowait(), `SELECT "id" FROM "jobs" FOR UPDATE NOWAIT;`},
		{"skip locked", q.ForUpdate().SkipLocked(), `SELECT "id" FROM "jobs" FOR UPDATE SKIP LOCKED;`},
		{"later lock strength wins", q.ForUpdate().ForShare().SkipLocked(), `SELECT "id" FROM "jobs" FOR SHARE SKIP LOCKED;`},
		{"option alone renders nothing", q.SkipLocked(), `SELECT "id" FROM "jobs";`},
	})
}

func TestSqlSelectBuilder_CTE(t *testing.T) {
	t.Parallel()
	recent := newSelect().From("orders").Where(syntax.Gt("created_at", syntax.P(1)))
	big := newSelect().From("recent").Where(syntax.Gt("amount", syntax.P(2)))
	tree := newSelect().From("categories").Where(syntax.IsNull("parent_id")).
		UnionAll(newSelect().From(syntax.As("categories", "c")).Select("c.*").
			InnerJoin(syntax.As("tree", "t"), syntax.Eq("c.parent_id", "t.id")))

	runSQLCases(t, []sqlCase{
		{
			name: "with",
			b:    newSelect().With("recent", recent).From("recent"),
			want: `WITH recent AS (SELECT * FROM "orders" WHERE "created_at" > $1) SELECT * FROM "recent";`,
		},
		{
			name: "multiple ctes keep call order",
			b:    newSelect().With("recent", recent).With("big", big).From("big"),
			want: `
				WITH recent AS (SELECT * FROM "orders" WHERE "created_at" > $1),
				big AS (SELECT * FROM "recent" WHERE "amount" > $2)
				SELECT * FROM "big";`,
		},
		{
			name: "with recursive",
			b:    newSelect().WithRecursive("tree", tree).From("tree"),
			want: `
				WITH RECURSIVE tree AS (SELECT * FROM "categories" WHERE "parent_id" IS NULL
				UNION ALL SELECT "c".* FROM "categories" AS "c" INNER JOIN "tree" AS "t" ON "c"."parent_id" = "t"."id")
				SELECT * FROM "tree";`,
		},
		{
			// RECURSIVE is a property of the whole WITH clause, not of a single CTE.
			name: "one recursive cte makes the whole clause recursive",
			b:    newSelect().With("recent", recent).WithRecursive("tree", tree).From("tree"),
			want: `
				WITH RECURSIVE recent AS (SELECT * FROM "orders" WHERE "created_at" > $1),
				tree AS (SELECT * FROM "categories" WHERE "parent_id" IS NULL
				UNION ALL SELECT "c".* FROM "categories" AS "c" INNER JOIN "tree" AS "t" ON "c"."parent_id" = "t"."id")
				SELECT * FROM "tree";`,
		},
	})
}

func TestSqlSelectBuilder_Subqueries(t *testing.T) {
	t.Parallel()
	totals := newSelect().From("orders").
		Select("user_id", syntax.As(syntax.Fn("SUM", "amount"), "total")).
		Where(syntax.Eq("status", syntax.P(1))).
		GroupBy("user_id")
	maxTotal := newSelect().From("orders").Select(syntax.Fn("MAX", "amount"))

	runSQLCases(t, []sqlCase{
		{
			name: "in from",
			b:    newSelect().From(syntax.As(totals, "t")).Where(syntax.Gt("t.total", syntax.P(2))),
			want: `
				SELECT * FROM
				(SELECT "user_id", SUM("amount") AS "total" FROM "orders" WHERE "status" = $1 GROUP BY "user_id") AS "t"
				WHERE "t"."total" > $2;`,
		},
		{
			name: "in join",
			b:    newSelect().From(syntax.As("users", "u")).InnerJoin(syntax.As(totals, "t"), syntax.Eq("u.id", "t.user_id")),
			want: `
				SELECT * FROM "users" AS "u" INNER JOIN
				(SELECT "user_id", SUM("amount") AS "total" FROM "orders" WHERE "status" = $1 GROUP BY "user_id") AS "t"
				ON "u"."id" = "t"."user_id";`,
		},
		{
			name: "as scalar in select and where",
			b: newSelect().From("orders").
				Select("id", syntax.As(maxTotal, "max_amount")).
				Where(syntax.Eq("amount", maxTotal)),
			want: `
				SELECT "id", (SELECT MAX("amount") FROM "orders") AS "max_amount" FROM "orders"
				WHERE "amount" = (SELECT MAX("amount") FROM "orders");`,
		},
	})
}

// TestSqlSelectBuilder_AllClauses pins the order in which clauses are rendered,
// independent of the order the methods are called in.
func TestSqlSelectBuilder_AllClauses(t *testing.T) {
	t.Parallel()
	got := newSelect().
		ForUpdate().SkipLocked().
		Offset(20).Limit(10).
		OrderBy(syntax.Desc("total")).
		Having(syntax.Gt(syntax.Fn("SUM", "o.amount"), syntax.P(2))).
		GroupBy("u.id").
		Where(syntax.Eq("u.status", syntax.P(1))).
		LeftJoin(syntax.As("orders", "o"), syntax.Eq("u.id", "o.user_id")).
		From(syntax.As("users", "u")).
		Select("u.id", syntax.As(syntax.Fn("SUM", "o.amount"), "total")).
		Distinct().
		With("recent", newSelect().From("orders")).
		UnionAll(newSelect().From("archived_totals").Select("user_id", "total")).
		ToSql()

	sqltesting.AssertSQL(t, got, `
		WITH recent AS (SELECT * FROM "orders")
		SELECT DISTINCT "u"."id", SUM("o"."amount") AS "total"
		FROM "users" AS "u"
		LEFT JOIN "orders" AS "o" ON "u"."id" = "o"."user_id"
		WHERE "u"."status" = $1
		GROUP BY "u"."id"
		HAVING SUM("o"."amount") > $2
		ORDER BY "total" DESC
		LIMIT 10 OFFSET 20
		FOR UPDATE SKIP LOCKED
		UNION ALL SELECT "user_id", "total" FROM "archived_totals";`)
}

// TestSqlSelectBuilder_BranchIsolation checks that deriving two queries from a
// shared base does not leak appended clauses between them or into the base.
// The base holds three items per list so that a backing array with spare
// capacity would be shared if a method appended in place.
func TestSqlSelectBuilder_BranchIsolation(t *testing.T) {
	t.Parallel()
	cte := newSelect().From("t")
	other := newSelect().From("s")
	base := newSelect().From("u").
		With("a", cte).With("b", cte).With("c", cte).
		CrossJoin("j1").CrossJoin("j2").CrossJoin("j3").
		Union(other).Union(other).Union(other)

	left := base.With("l", cte).CrossJoin("jl").Except(other)
	right := base.With("r", cte).CrossJoin("jr").Intersect(other)

	sqltesting.AssertSQL(t, base.ToSql(), `
		WITH a AS (SELECT * FROM "t"), b AS (SELECT * FROM "t"), c AS (SELECT * FROM "t")
		SELECT * FROM "u" CROSS JOIN "j1" CROSS JOIN "j2" CROSS JOIN "j3"
		UNION SELECT * FROM "s" UNION SELECT * FROM "s" UNION SELECT * FROM "s";`)
	sqltesting.AssertSQL(t, left.ToSql(), `
		WITH a AS (SELECT * FROM "t"), b AS (SELECT * FROM "t"), c AS (SELECT * FROM "t"), l AS (SELECT * FROM "t")
		SELECT * FROM "u" CROSS JOIN "j1" CROSS JOIN "j2" CROSS JOIN "j3" CROSS JOIN "jl"
		UNION SELECT * FROM "s" UNION SELECT * FROM "s" UNION SELECT * FROM "s" EXCEPT SELECT * FROM "s";`)
	sqltesting.AssertSQL(t, right.ToSql(), `
		WITH a AS (SELECT * FROM "t"), b AS (SELECT * FROM "t"), c AS (SELECT * FROM "t"), r AS (SELECT * FROM "t")
		SELECT * FROM "u" CROSS JOIN "j1" CROSS JOIN "j2" CROSS JOIN "j3" CROSS JOIN "jr"
		UNION SELECT * FROM "s" UNION SELECT * FROM "s" UNION SELECT * FROM "s" INTERSECT SELECT * FROM "s";`)
}
