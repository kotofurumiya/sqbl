package builder_test

import (
	"strings"
	"testing"

	"github.com/kotofurumiya/sqbl/builder"

	"github.com/kotofurumiya/sqbl/internal/sqltesting"
	"github.com/kotofurumiya/sqbl/syntax"
)

func TestSelectBuilder_Clauses(t *testing.T) {
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
			name: "list and single-value clauses are replaced by later calls",
			b: newSelect().
				From("old").From("users").
				Select("old").Select("id").
				GroupBy("old").GroupBy("id").
				OrderBy("old").OrderBy("id").
				DistinctOn("old").DistinctOn("id").
				Limit(1).Limit(10).
				Offset(1).Offset(20),
			want: `SELECT DISTINCT ON ("id") "id" FROM "users" GROUP BY "id" ORDER BY "id" LIMIT 10 OFFSET 20;`,
		},
		{
			// Each call narrows the query, so filters added to a shared base are never lost.
			name: "where and having conditions accumulate with and",
			b: newSelect().From("orders").GroupBy("user_id").
				Where(syntax.Eq("tenant_id", syntax.P(1))).
				Where(syntax.Or(syntax.Eq("status", syntax.P(2)), syntax.Eq("status", syntax.P(3)))).
				Having(syntax.Gt(syntax.Fn("COUNT", "*"), 1)).
				Having(syntax.Lt(syntax.Fn("SUM", "amount"), 100)),
			want: `
				SELECT * FROM "orders"
				WHERE "tenant_id" = $1 AND ("status" = $2 OR "status" = $3)
				GROUP BY "user_id"
				HAVING COUNT(*) > 1 AND SUM("amount") < 100;`,
		},
		{
			name: "nil conditions are ignored",
			b:    newSelect().From("users").Where(nil).Where(syntax.Eq("id", 1)).Where(nil).Having(nil),
			want: `SELECT * FROM "users" WHERE "id" = 1;`,
		},
		{
			name: "distinct on accepts expressions",
			b:    newSelect().From("users").DistinctOn(syntax.Fn("LOWER", "email")).OrderBy(syntax.Fn("LOWER", "email")),
			want: `SELECT DISTINCT ON (LOWER("email")) * FROM "users" ORDER BY LOWER("email");`,
		},
	})
}

func TestSelectBuilder_Joins(t *testing.T) {
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
				LEFT JOIN LATERAL (SELECT "amount" FROM "orders" WHERE "user_id" = "u"."id" LIMIT 1) AS "latest" ON TRUE;`,
		},
		{
			name: "cross lateral has no on",
			b:    users.CrossLateralJoin(latest, "latest"),
			want: `
				SELECT "u"."id" FROM "users" AS "u"
				CROSS JOIN LATERAL (SELECT "amount" FROM "orders" WHERE "user_id" = "u"."id" LIMIT 1) AS "latest";`,
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

func TestSelectBuilder_InvalidSources(t *testing.T) {
	t.Parallel()
	users := newSelect().From("users")

	runErrCases(t, []errCase{
		{"join without on", users.Join("orders", nil), "Join needs an ON condition"},
		{"inner join without on", users.InnerJoin("orders", nil), "InnerJoin needs an ON condition"},
		{"left join without on", users.LeftJoin("orders", nil), "LeftJoin needs an ON condition"},
		{"right join without on", users.RightJoin("orders", nil), "RightJoin needs an ON condition"},
		{"full join without on", users.FullJoin("orders", nil), "FullJoin needs an ON condition"},
		{"nil join table", users.CrossJoin(nil), "CrossJoin got nil"},
		{"nil from", newSelect().From(nil), "From got nil"},
	})
}

func TestSelectBuilder_SetOperations(t *testing.T) {
	t.Parallel()
	a := newSelect().From("a").Select("id")
	b := newSelect().From("b").Select("id")
	c := newSelect().From("c").Select("id")

	runSQLCases(t, []sqlCase{
		{"union", a.Union(b), `SELECT "id" FROM "a" UNION SELECT "id" FROM "b";`},
		{"union all", a.UnionAll(b), `SELECT "id" FROM "a" UNION ALL SELECT "id" FROM "b";`},
		{"intersect", a.Intersect(b), `SELECT "id" FROM "a" INTERSECT SELECT "id" FROM "b";`},
		{"intersect all", a.IntersectAll(b), `SELECT "id" FROM "a" INTERSECT ALL SELECT "id" FROM "b";`},
		{"except", a.Except(b), `SELECT "id" FROM "a" EXCEPT SELECT "id" FROM "b";`},
		{"except all", a.ExceptAll(b), `SELECT "id" FROM "a" EXCEPT ALL SELECT "id" FROM "b";`},
		{
			name: "operations chain left to right without parentheses",
			b:    a.Union(b.Where(syntax.Gt("id", syntax.P(1)))).Except(c).UnionAll(c),
			want: `
				SELECT "id" FROM "a"
				UNION SELECT "id" FROM "b" WHERE "id" > $1
				EXCEPT SELECT "id" FROM "c"
				UNION ALL SELECT "id" FROM "c";`,
		},
		{
			name: "order by, limit and offset apply to the combined rows",
			b:    a.Union(b).OrderBy(syntax.Desc("id")).Limit(10).Offset(5),
			want: `SELECT "id" FROM "a" UNION SELECT "id" FROM "b" ORDER BY "id" DESC LIMIT 10 OFFSET 5;`,
		},
		{
			name: "operands with their own trailing clauses are parenthesized",
			b:    a.OrderBy("id").Limit(1).Union(b.ForUpdate()).Union(c.Offset(2)),
			want: `
				(SELECT "id" FROM "a" ORDER BY "id" LIMIT 1)
				UNION (SELECT "id" FROM "b" FOR UPDATE)
				UNION (SELECT "id" FROM "c" OFFSET 2);`,
		},
		{
			// A WITH on the leftmost operand already scopes over the whole compound.
			name: "with is parenthesized only on the right",
			b:    a.With("x", c).Union(b.With("y", c)),
			want: `
				WITH "x" AS (SELECT "id" FROM "c") SELECT "id" FROM "a"
				UNION (WITH "y" AS (SELECT "id" FROM "c") SELECT "id" FROM "b");`,
		},
		{
			// Flattening would read as (a EXCEPT b) UNION c.
			name: "compound on the right is parenthesized",
			b:    a.Except(b.Union(c)),
			want: `SELECT "id" FROM "a" EXCEPT (SELECT "id" FROM "b" UNION SELECT "id" FROM "c");`,
		},
		{
			// INTERSECT binds tighter than UNION, so a flat chain would read as a UNION (b INTERSECT c).
			name: "intersect after union groups the earlier operations",
			b:    a.Union(b).Intersect(c),
			want: `(SELECT "id" FROM "a" UNION SELECT "id" FROM "b") INTERSECT SELECT "id" FROM "c";`,
		},
		{
			name: "intersect after intersect stays flat",
			b:    a.Intersect(b).Intersect(c),
			want: `SELECT "id" FROM "a" INTERSECT SELECT "id" FROM "b" INTERSECT SELECT "id" FROM "c";`,
		},
		{
			name: "continuing a limited compound groups it first",
			b:    a.Union(b).Limit(1).Union(c),
			want: `(SELECT "id" FROM "a" UNION SELECT "id" FROM "b" LIMIT 1) UNION SELECT "id" FROM "c";`,
		},
		{
			name: "compound as a source, an in operand and a cte",
			b: newSelect().
				With("ids", a.Union(b)).
				From(syntax.As(a.Union(b), "u")).
				Where(syntax.In("u.id", a.Intersect(b))),
			want: `
				WITH "ids" AS (SELECT "id" FROM "a" UNION SELECT "id" FROM "b")
				SELECT * FROM (SELECT "id" FROM "a" UNION SELECT "id" FROM "b") AS "u"
				WHERE "u"."id" IN (SELECT "id" FROM "a" INTERSECT SELECT "id" FROM "b");`,
		},
	})
}

func TestSelectBuilder_Locking(t *testing.T) {
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
		{"later option wins", q.ForUpdate().SkipLocked().Nowait(), `SELECT "id" FROM "jobs" FOR UPDATE NOWAIT;`},
	})
	runErrCases(t, []errCase{
		{"nowait without lock", q.Nowait(), "Nowait needs a locking clause"},
		{"skip locked without lock", q.SkipLocked(), "SkipLocked needs a locking clause"},
	})
}

func TestSelectBuilder_CTE(t *testing.T) {
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
			want: `WITH "recent" AS (SELECT * FROM "orders" WHERE "created_at" > $1) SELECT * FROM "recent";`,
		},
		{
			name: "multiple ctes keep call order",
			b:    newSelect().With("recent", recent).With("big", big).From("big"),
			want: `
				WITH "recent" AS (SELECT * FROM "orders" WHERE "created_at" > $1),
				"big" AS (SELECT * FROM "recent" WHERE "amount" > $2)
				SELECT * FROM "big";`,
		},
		{
			name: "with recursive",
			b:    newSelect().WithRecursive("tree", tree).From("tree"),
			want: `
				WITH RECURSIVE "tree" AS (SELECT * FROM "categories" WHERE "parent_id" IS NULL
				UNION ALL SELECT "c".* FROM "categories" AS "c" INNER JOIN "tree" AS "t" ON "c"."parent_id" = "t"."id")
				SELECT * FROM "tree";`,
		},
		{
			// RECURSIVE is a property of the whole WITH clause, not of a single CTE.
			name: "one recursive cte makes the whole clause recursive",
			b:    newSelect().With("recent", recent).WithRecursive("tree", tree).From("tree"),
			want: `
				WITH RECURSIVE "recent" AS (SELECT * FROM "orders" WHERE "created_at" > $1),
				"tree" AS (SELECT * FROM "categories" WHERE "parent_id" IS NULL
				UNION ALL SELECT "c".* FROM "categories" AS "c" INNER JOIN "tree" AS "t" ON "c"."parent_id" = "t"."id")
				SELECT * FROM "tree";`,
		},
		{
			// The name is quoted like any identifier, so reserved words work as CTE names.
			name: "name is quoted",
			b:    newSelect().With("order", recent).From("order"),
			want: `WITH "order" AS (SELECT * FROM "orders" WHERE "created_at" > $1) SELECT * FROM "order";`,
		},
	})
}

func TestSelectBuilder_Subqueries(t *testing.T) {
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
		{
			// A lone subquery already brings its own parentheses; doubling them would
			// make PostgreSQL treat it as a scalar subquery.
			name: "as the only operand of in and not in",
			b: newSelect().From("orders").Where(syntax.And(
				syntax.In("user_id", newSelect().From("vips").Select("id")),
				syntax.NotIn("id", newSelect().From("refunds").Select("order_id")),
			)),
			want: `
				SELECT * FROM "orders"
				WHERE "user_id" IN (SELECT "id" FROM "vips")
				AND "id" NOT IN (SELECT "order_id" FROM "refunds");`,
		},
		{
			name: "among other in values",
			b:    newSelect().From("orders").Where(syntax.In("amount", 0, maxTotal)),
			want: `SELECT * FROM "orders" WHERE "amount" IN (0, (SELECT MAX("amount") FROM "orders"));`,
		},
	})
}

// TestSelectBuilder_AllClauses pins the order in which clauses are rendered,
// independent of the order the methods are called in.
func TestSelectBuilder_AllClauses(t *testing.T) {
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
		MustSQL()

	sqltesting.AssertSQL(t, got, `
		WITH "recent" AS (SELECT * FROM "orders")
		SELECT DISTINCT "u"."id", SUM("o"."amount") AS "total"
		FROM "users" AS "u"
		LEFT JOIN "orders" AS "o" ON "u"."id" = "o"."user_id"
		WHERE "u"."status" = $1
		GROUP BY "u"."id"
		HAVING SUM("o"."amount") > $2
		ORDER BY "total" DESC
		LIMIT 10 OFFSET 20
		FOR UPDATE SKIP LOCKED;`)
}

// TestSelectBuilder_BranchIsolation checks that deriving two queries from a
// shared base does not leak appended clauses between them or into the base.
// The base holds three items per list so that a backing array with spare
// capacity would be shared if a method appended in place.
func TestSelectBuilder_BranchIsolation(t *testing.T) {
	t.Parallel()
	cte := newSelect().From("t")
	base := newSelect().From("u").
		With("a", cte).With("b", cte).With("c", cte).
		CrossJoin("j1").CrossJoin("j2").CrossJoin("j3").
		Where(syntax.Eq("w", 1)).Where(syntax.Eq("w", 2)).Where(syntax.Eq("w", 3))

	left := base.With("l", cte).CrossJoin("jl").Where(syntax.Eq("l", 1))
	right := base.With("r", cte).CrossJoin("jr").Where(syntax.Eq("r", 1))

	sqltesting.AssertSQL(t, base.MustSQL(), `
		WITH "a" AS (SELECT * FROM "t"), "b" AS (SELECT * FROM "t"), "c" AS (SELECT * FROM "t")
		SELECT * FROM "u" CROSS JOIN "j1" CROSS JOIN "j2" CROSS JOIN "j3"
		WHERE "w" = 1 AND "w" = 2 AND "w" = 3;`)
	sqltesting.AssertSQL(t, left.MustSQL(), `
		WITH "a" AS (SELECT * FROM "t"), "b" AS (SELECT * FROM "t"), "c" AS (SELECT * FROM "t"), "l" AS (SELECT * FROM "t")
		SELECT * FROM "u" CROSS JOIN "j1" CROSS JOIN "j2" CROSS JOIN "j3" CROSS JOIN "jl"
		WHERE "w" = 1 AND "w" = 2 AND "w" = 3 AND "l" = 1;`)
	sqltesting.AssertSQL(t, right.MustSQL(), `
		WITH "a" AS (SELECT * FROM "t"), "b" AS (SELECT * FROM "t"), "c" AS (SELECT * FROM "t"), "r" AS (SELECT * FROM "t")
		SELECT * FROM "u" CROSS JOIN "j1" CROSS JOIN "j2" CROSS JOIN "j3" CROSS JOIN "jr"
		WHERE "w" = 1 AND "w" = 2 AND "w" = 3 AND "r" = 1;`)
}

// TestCompoundBuilder_BranchIsolation does the same for set operations.
func TestCompoundBuilder_BranchIsolation(t *testing.T) {
	t.Parallel()
	s := newSelect().From("s")
	base := newSelect().From("u").Union(s).Union(s).Union(s)
	left := base.Except(s)
	right := base.UnionAll(s)

	sqltesting.AssertSQL(t, base.MustSQL(), `
		SELECT * FROM "u" UNION SELECT * FROM "s" UNION SELECT * FROM "s" UNION SELECT * FROM "s";`)
	sqltesting.AssertSQL(t, left.MustSQL(), `
		SELECT * FROM "u" UNION SELECT * FROM "s" UNION SELECT * FROM "s" UNION SELECT * FROM "s" EXCEPT SELECT * FROM "s";`)
	sqltesting.AssertSQL(t, right.MustSQL(), `
		SELECT * FROM "u" UNION SELECT * FROM "s" UNION SELECT * FROM "s" UNION SELECT * FROM "s" UNION ALL SELECT * FROM "s";`)
}

func TestCompoundBuilder_ZeroValue(t *testing.T) {
	t.Parallel()
	_, err := builder.CompoundBuilder{}.ToSQLWithDialect(pg)
	if err == nil || !strings.Contains(err.Error(), "needs a left operand") {
		t.Errorf("want an error about the missing left operand, got %v", err)
	}
}
