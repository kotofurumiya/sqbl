package builder_test

import (
	"testing"

	"github.com/kotofurumiya/sqbl/builder"
	"github.com/kotofurumiya/sqbl/dialect"
	"github.com/kotofurumiya/sqbl/internal/sqltesting"
	"github.com/kotofurumiya/sqbl/syntax"
)

func newBuilder() builder.SqlSelectBuilder {
	return (builder.SqlSelectBuilder{}).Dialect(&dialect.SimpleDialect{})
}

func TestSqlSelectBuilder_SelectAll(t *testing.T) {
	t.Parallel()
	got := newBuilder().From("users").ToSql()
	want := `SELECT * FROM "users";`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlSelectBuilder_SelectColumns(t *testing.T) {
	t.Parallel()
	got := sqltesting.Compact(
		newBuilder().
			From("users").
			Select("id", "name").
			ToSql(),
	)
	want := sqltesting.Compact(`
		SELECT
		    "id", "name"
		FROM
		    "users";
	`)
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlSelectBuilder_SelectThenFrom(t *testing.T) {
	t.Parallel()
	// Select().From() and From().Select() should produce the same result.
	fromFirst := newBuilder().From("users").Select("id").ToSql()
	selectFirst := newBuilder().Select("id").From("users").ToSql()
	if fromFirst != selectFirst {
		t.Errorf("From().Select() = %q; Select().From() = %q", fromFirst, selectFirst)
	}
}

func TestSqlSelectBuilder_SelectOnly(t *testing.T) {
	t.Parallel()
	// SELECT without FROM is valid (e.g. SELECT NOW()).
	// Expressions containing '(' are left unquoted by isSimpleIdentifier.
	got := newBuilder().Select("NOW()").ToSql()
	want := `SELECT NOW();`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlSelectBuilder_Distinct(t *testing.T) {
	t.Parallel()
	got := sqltesting.Compact(
		newBuilder().
			From("users").
			Select("country").
			Distinct().
			ToSql(),
	)
	want := sqltesting.Compact(`
		SELECT DISTINCT
		    "country"
		FROM
		    "users";
	`)
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlSelectBuilder_Alias(t *testing.T) {
	t.Parallel()
	got := sqltesting.Compact(
		newBuilder().
			From(syntax.As("users", "u")).
			Select(syntax.As("COUNT(*)", "cnt")).
			ToSql(),
	)
	want := sqltesting.Compact(`
		SELECT
		    COUNT(*) AS "cnt"
		FROM
		    "users" AS "u";
	`)
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlSelectBuilder_Where(t *testing.T) {
	t.Parallel()
	got := sqltesting.Compact(
		newBuilder().
			From("users").
			Select("id").
			Where(syntax.Eq("status", "'active'")).
			ToSql(),
	)
	want := sqltesting.Compact(`
		SELECT
		    "id"
		FROM
		    "users"
		WHERE
		    "status" = 'active';
	`)
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlSelectBuilder_WhereAllComparisons(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		expr syntax.SqlFragment
		want string
	}{
		{"Eq", syntax.Eq("a", 1), `"a" = 1`},
		{"Ne", syntax.Ne("a", 1), `"a" <> 1`},
		{"Lt", syntax.Lt("a", 1), `"a" < 1`},
		{"Lte", syntax.Lte("a", 1), `"a" <= 1`},
		{"Gt", syntax.Gt("a", 1), `"a" > 1`},
		{"Gte", syntax.Gte("a", 1), `"a" >= 1`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// t.Parallel()
			got := newBuilder().From("t").Select("id").Where(tt.expr).ToSql()
			want := `SELECT "id" FROM "t" WHERE ` + tt.want + `;`
			if got != want {
				t.Errorf("got %q; want %q", got, want)
			}
		})
	}
}

func TestSqlSelectBuilder_WhereAndOr(t *testing.T) {
	t.Parallel()
	got := sqltesting.Compact(
		newBuilder().
			From("users").
			Select("id").
			Where(syntax.And(
				syntax.Eq("status", "'active'"),
				syntax.Or(syntax.Eq("role", "'admin'"), syntax.Eq("role", "'mod'")),
			)).
			ToSql(),
	)
	want := sqltesting.Compact(`
		SELECT
		    "id"
		FROM
		    "users"
		WHERE
		    "status" = 'active' AND ("role" = 'admin' OR "role" = 'mod');
	`)
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlSelectBuilder_GroupByHaving(t *testing.T) {
	t.Parallel()
	got := sqltesting.Compact(
		newBuilder().
			From("orders").
			Select("user_id", syntax.As(syntax.Fn("SUM", "amount"), "total")).
			GroupBy("user_id").
			Having(syntax.Gt(syntax.Fn("SUM", "amount"), 1000)).
			ToSql(),
	)
	want := sqltesting.Compact(`
		SELECT
		    "user_id", SUM("amount") AS "total"
		FROM
		    "orders"
		GROUP BY
		    "user_id"
		HAVING
		    SUM("amount") > 1000;
	`)
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlSelectBuilder_OrderBy(t *testing.T) {
	t.Parallel()
	got := sqltesting.Compact(
		newBuilder().
			From("users").
			Select("id", "name").
			OrderBy(syntax.Asc("name"), syntax.Desc("created_at")).
			ToSql(),
	)
	want := sqltesting.Compact(`
		SELECT
		    "id", "name"
		FROM
		    "users"
		ORDER BY
		    "name" ASC,
		    "created_at" DESC;
	`)
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlSelectBuilder_LimitOffset(t *testing.T) {
	t.Parallel()
	got := sqltesting.Compact(
		newBuilder().
			From("users").
			Select("id").
			Limit(10).
			Offset(20).
			ToSql(),
	)
	want := sqltesting.Compact(`
		SELECT
		    "id"
		FROM
		    "users"
		LIMIT 10
		OFFSET 20;
	`)
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlSelectBuilder_InnerJoin(t *testing.T) {
	t.Parallel()
	got := sqltesting.Compact(
		newBuilder().
			From(syntax.As("users", "u")).
			Select("u.id", "o.total").
			InnerJoin(syntax.As("orders", "o"), syntax.Eq("u.id", "o.user_id")).
			ToSql(),
	)
	want := sqltesting.Compact(`
		SELECT
		    "u.id", "o.total"
		FROM
		    "users" AS "u"
		INNER JOIN
		    "orders" AS "o" ON "u.id" = "o.user_id";
	`)
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlSelectBuilder_JoinIsInnerJoin(t *testing.T) {
	t.Parallel()
	// Join() is an alias for InnerJoin() and must produce the same output.
	inner := newBuilder().
		From(syntax.As("users", "u")).
		Select("u.id").
		InnerJoin(syntax.As("orders", "o"), syntax.Eq("u.id", "o.user_id")).
		ToSql()
	join := newBuilder().
		From(syntax.As("users", "u")).
		Select("u.id").
		Join(syntax.As("orders", "o"), syntax.Eq("u.id", "o.user_id")).
		ToSql()
	if inner != join {
		t.Errorf("Join() and InnerJoin() differ:\n  Join()      = %q\n  InnerJoin() = %q", join, inner)
	}
}

func TestSqlSelectBuilder_LeftJoin(t *testing.T) {
	t.Parallel()
	got := sqltesting.Compact(
		newBuilder().
			From(syntax.As("users", "u")).
			Select("u.id", "p.bio").
			LeftJoin(syntax.As("profiles", "p"), syntax.Eq("u.id", "p.user_id")).
			ToSql(),
	)
	want := sqltesting.Compact(`
		SELECT
		    "u.id", "p.bio"
		FROM
		    "users" AS "u"
		LEFT JOIN
		    "profiles" AS "p" ON "u.id" = "p.user_id";
	`)
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlSelectBuilder_CrossJoin(t *testing.T) {
	t.Parallel()
	got := sqltesting.Compact(
		newBuilder().
			From("colors").
			Select("*").
			CrossJoin("sizes").
			ToSql(),
	)
	want := sqltesting.Compact(`
		SELECT
		    *
		FROM
		    "colors"
		CROSS JOIN "sizes";
	`)
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlSelectBuilder_Locking(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		got  string
		want string
	}{
		{
			"ForUpdate",
			newBuilder().From("users").Select("id").ForUpdate().ToSql(),
			`SELECT "id" FROM "users" FOR UPDATE;`,
		},
		{
			"ForUpdate/Nowait",
			newBuilder().From("users").Select("id").ForUpdate().Nowait().ToSql(),
			`SELECT "id" FROM "users" FOR UPDATE NOWAIT;`,
		},
		{
			"ForUpdate/SkipLocked",
			newBuilder().From("users").Select("id").ForUpdate().SkipLocked().ToSql(),
			`SELECT "id" FROM "users" FOR UPDATE SKIP LOCKED;`,
		},
		{
			"ForShare",
			newBuilder().From("users").Select("id").ForShare().ToSql(),
			`SELECT "id" FROM "users" FOR SHARE;`,
		},
		{
			"ForNoKeyUpdate",
			newBuilder().From("users").Select("id").ForNoKeyUpdate().ToSql(),
			`SELECT "id" FROM "users" FOR NO KEY UPDATE;`,
		},
		{
			"ForKeyShare",
			newBuilder().From("users").Select("id").ForKeyShare().ToSql(),
			`SELECT "id" FROM "users" FOR KEY SHARE;`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// t.Parallel()
			if tt.got != tt.want {
				t.Errorf("got %q; want %q", tt.got, tt.want)
			}
		})
	}
}

func TestSqlSelectBuilder_With(t *testing.T) {
	t.Parallel()
	recent := newBuilder().
		From("orders").
		Select("user_id", "amount").
		Where(syntax.Gte("created_at", "'2025-01-01'"))
	got := sqltesting.Compact(
		newBuilder().
			With("recent_orders", recent).
			From("recent_orders").
			Select("user_id", syntax.As("SUM(amount)", "total")).
			GroupBy("user_id").
			ToSql(),
	)
	want := sqltesting.Compact(`
		WITH recent_orders AS (SELECT "user_id", "amount" FROM "orders" WHERE "created_at" >= '2025-01-01')
		SELECT
		    "user_id", SUM(amount) AS "total"
		FROM
		    "recent_orders"
		GROUP BY
		    "user_id";
	`)
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlSelectBuilder_Union(t *testing.T) {
	t.Parallel()
	q1 := newBuilder().From("users").Select("id", "name")
	q2 := newBuilder().From("archived_users").Select("id", "name")
	got := sqltesting.Compact(q1.Union(q2).ToSql())
	want := sqltesting.Compact(`
		SELECT
		    "id", "name"
		FROM
		    "users"
		UNION SELECT
		    "id", "name"
		FROM
		    "archived_users";
	`)
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlSelectBuilder_UnionAll(t *testing.T) {
	t.Parallel()
	q1 := newBuilder().From("users").Select("id", "name")
	q2 := newBuilder().From("archived_users").Select("id", "name")
	got := sqltesting.Compact(q1.UnionAll(q2).ToSql())
	want := sqltesting.Compact(`
		SELECT
		    "id", "name"
		FROM
		    "users"
		UNION ALL SELECT
		    "id", "name"
		FROM
		    "archived_users";
	`)
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlSelectBuilder_RightJoin(t *testing.T) {
	t.Parallel()
	got := sqltesting.Compact(
		newBuilder().
			From(syntax.As("orders", "o")).
			Select("o.id", "u.name").
			RightJoin(syntax.As("users", "u"), syntax.Eq("o.user_id", "u.id")).
			ToSql(),
	)
	want := sqltesting.Compact(`
		SELECT
		    "o.id", "u.name"
		FROM
		    "orders" AS "o"
		RIGHT JOIN
		    "users" AS "u" ON "o.user_id" = "u.id";
	`)
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlSelectBuilder_FullJoin(t *testing.T) {
	t.Parallel()
	got := sqltesting.Compact(
		newBuilder().
			From(syntax.As("employees", "e")).
			Select("e.name", "d.name").
			FullJoin(syntax.As("departments", "d"), syntax.Eq("e.dept_id", "d.id")).
			ToSql(),
	)
	want := sqltesting.Compact(`
		SELECT
		    "e.name", "d.name"
		FROM
		    "employees" AS "e"
		FULL OUTER JOIN
		    "departments" AS "d" ON "e.dept_id" = "d.id";
	`)
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlSelectBuilder_Intersect(t *testing.T) {
	t.Parallel()
	q1 := newBuilder().From("active_users").Select("id")
	q2 := newBuilder().From("premium_users").Select("id")
	got := sqltesting.Compact(q1.Intersect(q2).ToSql())
	want := sqltesting.Compact(`
		SELECT "id" FROM "active_users"
		INTERSECT SELECT "id" FROM "premium_users";
	`)
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlSelectBuilder_IntersectAll(t *testing.T) {
	t.Parallel()
	q1 := newBuilder().From("active_users").Select("id")
	q2 := newBuilder().From("premium_users").Select("id")
	got := sqltesting.Compact(q1.IntersectAll(q2).ToSql())
	want := sqltesting.Compact(`
		SELECT "id" FROM "active_users"
		INTERSECT ALL SELECT "id" FROM "premium_users";
	`)
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlSelectBuilder_Except(t *testing.T) {
	t.Parallel()
	q1 := newBuilder().From("all_users").Select("id")
	q2 := newBuilder().From("banned_users").Select("id")
	got := sqltesting.Compact(q1.Except(q2).ToSql())
	want := sqltesting.Compact(`
		SELECT "id" FROM "all_users"
		EXCEPT SELECT "id" FROM "banned_users";
	`)
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlSelectBuilder_ExceptAll(t *testing.T) {
	t.Parallel()
	q1 := newBuilder().From("all_users").Select("id")
	q2 := newBuilder().From("banned_users").Select("id")
	got := sqltesting.Compact(q1.ExceptAll(q2).ToSql())
	want := sqltesting.Compact(`
		SELECT "id" FROM "all_users"
		EXCEPT ALL SELECT "id" FROM "banned_users";
	`)
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlSelectBuilder_WhereNewExpressions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		expr syntax.SqlFragment
		want string
	}{
		{"Not", syntax.Not(syntax.Eq("active", false)), `NOT ("active" = FALSE)`},
		{"In", syntax.In("status", "'active'", "'pending'"), `"status" IN ('active', 'pending')`},
		{"NotIn", syntax.NotIn("role", "'guest'"), `"role" NOT IN ('guest')`},
		{"IsNull", syntax.IsNull("deleted_at"), `"deleted_at" IS NULL`},
		{"IsNotNull", syntax.IsNotNull("email"), `"email" IS NOT NULL`},
		{"Between", syntax.Between("age", 18, 65), `"age" BETWEEN 18 AND 65`},
		{"Like", syntax.Like("name", "'%foo%'"), `"name" LIKE '%foo%'`},
		{"ILike", syntax.ILike("name", "'%foo%'"), `"name" ILIKE '%foo%'`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// t.Parallel()
			got := newBuilder().From("t").Select("id").Where(tt.expr).ToSql()
			want := `SELECT "id" FROM "t" WHERE ` + tt.want + `;`
			if got != want {
				t.Errorf("got %q; want %q", got, want)
			}
		})
	}
}

func TestSqlSelectBuilder_SubqueryInFrom(t *testing.T) {
	t.Parallel()
	sub := newBuilder().
		From("orders").
		Select("user_id", syntax.As("SUM(amount)", "total")).
		GroupBy("user_id")
	got := sqltesting.Compact(
		newBuilder().
			From(syntax.As(sub, "sub")).
			Select("sub.user_id", "sub.total").
			Where(syntax.Gt("sub.total", 1000)).
			ToSql(),
	)
	want := sqltesting.Compact(`
		SELECT
		    "sub.user_id", "sub.total"
		FROM
		    (SELECT "user_id", SUM(amount) AS "total" FROM "orders" GROUP BY "user_id") AS "sub"
		WHERE
		    "sub.total" > 1000;
	`)
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlSelectBuilder_SubqueryInJoin(t *testing.T) {
	t.Parallel()
	sub := newBuilder().
		From("orders").
		Select("user_id", syntax.As("SUM(amount)", "total")).
		GroupBy("user_id")
	got := sqltesting.Compact(
		newBuilder().
			From(syntax.As("users", "u")).
			Select("u.id", "totals.total").
			InnerJoin(syntax.As(sub, "totals"), syntax.Eq("u.id", "totals.user_id")).
			ToSql(),
	)
	want := sqltesting.Compact(`
		SELECT
		    "u.id", "totals.total"
		FROM
		    "users" AS "u"
		INNER JOIN
		    (SELECT "user_id", SUM(amount) AS "total" FROM "orders" GROUP BY "user_id") AS "totals"
		    ON "u.id" = "totals.user_id";
	`)
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlSelectBuilder_DistinctOn(t *testing.T) {
	t.Parallel()
	got := sqltesting.Compact(
		(&builder.SqlSelectBuilder{}).
			Dialect(&dialect.PostgresDialect{}).
			From("orders").
			Select("user_id", "amount").
			DistinctOn("user_id").
			OrderBy(syntax.Desc("amount")).
			ToSql(),
	)
	want := sqltesting.Compact(`
		SELECT DISTINCT ON ("user_id")
		    "user_id", "amount"
		FROM
		    "orders"
		ORDER BY
		    "amount" DESC;
	`)
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlSelectBuilder_DistinctOn_NonPostgres_Fallback(t *testing.T) {
	t.Parallel()
	// Non-PostgreSQL dialects ignore DistinctOn and use plain SELECT.
	got := newBuilder().
		From("orders").
		Select("user_id").
		DistinctOn("user_id").
		ToSql()
	want := `SELECT "user_id" FROM "orders";`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlSelectBuilder_WithRecursive(t *testing.T) {
	t.Parallel()
	base := newBuilder().From("categories").Select("id", "parent_id", "name").Where(syntax.IsNull("parent_id"))
	recursive := newBuilder().From("categories").
		Select("categories.id", "categories.parent_id", "categories.name").
		InnerJoin(syntax.As("tree", "t"), syntax.Eq("categories.parent_id", "t.id"))
	tree := base.Union(recursive)

	got := sqltesting.Compact(
		newBuilder().
			From("tree").
			WithRecursive("tree", tree).
			Select("id", "name").
			ToSql(),
	)
	want := sqltesting.Compact(`
		WITH RECURSIVE tree AS (SELECT "id", "parent_id", "name" FROM "categories" WHERE "parent_id" IS NULL UNION SELECT "categories.id", "categories.parent_id", "categories.name" FROM "categories" INNER JOIN "tree" AS "t" ON "categories.parent_id" = "t.id")
		SELECT "id", "name" FROM "tree";
	`)
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlSelectBuilder_LeftLateralJoin(t *testing.T) {
	t.Parallel()
	latest := (&builder.SqlSelectBuilder{}).
		Dialect(&dialect.PostgresDialect{}).
		From("orders").
		Select("amount").
		Where(syntax.Eq("user_id", "u.id")).
		OrderBy(syntax.Desc("created_at")).
		Limit(1)

	got := sqltesting.Compact(
		(&builder.SqlSelectBuilder{}).
			Dialect(&dialect.PostgresDialect{}).
			From(syntax.As("users", "u")).
			Select("u.id", "latest.amount").
			LeftLateralJoin(latest, "latest").
			ToSql(),
	)
	want := sqltesting.Compact(`
		SELECT "u"."id", "latest"."amount"
		FROM "users" AS "u"
		LEFT LATERAL JOIN (SELECT "amount" FROM "orders" WHERE "user_id" = "u"."id" ORDER BY "created_at" DESC LIMIT 1) AS "latest" ON TRUE;
	`)
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlSelectBuilder_CrossLateralJoin(t *testing.T) {
	t.Parallel()
	vals := newBuilder().Select("1", "2", "3")

	got := sqltesting.Compact(
		newBuilder().
			From("users").
			Select("id", "n").
			CrossLateralJoin(vals, "nums").
			ToSql(),
	)
	want := sqltesting.Compact(`
		SELECT "id", "n"
		FROM "users"
		CROSS LATERAL JOIN (SELECT "1", "2", "3") AS "nums";
	`)
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}
