package sqblpg

import (
	"strings"
	"testing"

	"github.com/kotofurumiya/sqbl/internal/sqltesting"
	"github.com/kotofurumiya/sqbl/syntax"
)

func TestSqbl_BuildSQL(t *testing.T) {
	t.Parallel()

	t.Run("build common SQL", func(t *testing.T) {
		t.Parallel()
		expected := sqltesting.Compact(`
			SELECT
			    "u"."id",
			    "u"."name",
			    SUM("p"."amount") AS "total_spent"
			FROM
			    "users" AS "u"
			INNER JOIN
			    "purchases" AS "p" ON "u"."id" = "p"."user_id"
			WHERE
			    "p"."purchased_at" >= '2025-01-01'
			    AND ("u"."status" = 'active' OR "u"."role" = 'admin')
			GROUP BY
			    "u"."id",
			    "u"."name"
			HAVING
			    SUM("p"."amount") > 5000 AND COUNT("p"."id") < 100
			ORDER BY
			    "total_spent" DESC,
			    "u"."name" ASC
			LIMIT 10
			OFFSET 0;
		`)

		got := sqltesting.Compact(
			From(As("users", "u")).
				Select(
					"u.id",
					"u.name",
					As(Fn("SUM", "p.amount"), "total_spent"),
				).
				InnerJoin(As("purchases", "p"), Eq("u.id", "p.user_id")).
				Where(And(
					Gte("p.purchased_at", "'2025-01-01'"),
					Or(Eq("u.status", "'active'"), Eq("u.role", "'admin'")),
				)).
				GroupBy("u.id", "u.name").
				Having(And(
					Gt(Fn("SUM", "p.amount"), 5000),
					Lt(Fn("COUNT", "p.id"), 100),
				)).
				OrderBy(Desc("total_spent"), Asc("u.name")).
				Limit(10).
				Offset(0).
				ToSql(),
		)

		if got != expected {
			t.Errorf("ToSql() = %q; want %q", got, expected)
		}
	})

	t.Run("subquery in FROM", func(t *testing.T) {
		t.Parallel()
		expected := sqltesting.Compact(`
			SELECT
			    "sub"."user_id",
			    "sub"."total"
			FROM
			    (SELECT "user_id", SUM("amount") AS "total" FROM "orders" GROUP BY "user_id") AS "sub"
			WHERE
			    "sub"."total" > 1000;
		`)

		sub := From("orders").
			Select("user_id", As(Fn("SUM", "amount"), "total")).
			GroupBy("user_id")

		got := sqltesting.Compact(
			From(As(sub, "sub")).
				Select("sub.user_id", "sub.total").
				Where(Gt("sub.total", 1000)).
				ToSql(),
		)

		if got != expected {
			t.Errorf("ToSql() = %q; want %q", got, expected)
		}
	})

	t.Run("subquery in JOIN", func(t *testing.T) {
		t.Parallel()
		expected := sqltesting.Compact(`
			SELECT
			    "u"."id", "u"."name", "totals"."total"
			FROM
			    "users" AS "u"
			INNER JOIN
			    (SELECT "user_id", SUM("amount") AS "total" FROM "orders" GROUP BY "user_id") AS "totals"
			ON "u"."id" = "totals"."user_id";
		`)

		orderTotals := From("orders").
			Select("user_id", As(Fn("SUM", "amount"), "total")).
			GroupBy("user_id")

		got := sqltesting.Compact(
			From(As("users", "u")).
				Select("u.id", "u.name", "totals.total").
				InnerJoin(As(orderTotals, "totals"), Eq("u.id", "totals.user_id")).
				ToSql(),
		)

		if got != expected {
			t.Errorf("ToSql() = %q; want %q", got, expected)
		}
	})

	t.Run("locking", func(t *testing.T) {
		t.Parallel()
		tests := []struct {
			name string
			got  string
			want string
		}{
			{
				"ForUpdate",
				sqltesting.Compact(
					From("jobs").
						Select("id", "payload").
						Where(Eq("status", "'pending'")).
						OrderBy(Asc("id")).
						Limit(1).
						ForUpdate().
						ToSql(),
				),
				sqltesting.Compact(`
					SELECT
					    "id", "payload"
					FROM
					    "jobs"
					WHERE
					    "status" = 'pending'
					ORDER BY
					    "id" ASC
					LIMIT 1
					FOR UPDATE;
				`),
			},
			{
				"ForUpdate/SkipLocked",
				sqltesting.Compact(
					From("jobs").
						Select("id", "payload").
						Where(Eq("status", "'pending'")).
						OrderBy(Asc("id")).
						Limit(1).
						ForUpdate().SkipLocked().
						ToSql(),
				),
				sqltesting.Compact(`
					SELECT
					    "id", "payload"
					FROM
					    "jobs"
					WHERE
					    "status" = 'pending'
					ORDER BY
					    "id" ASC
					LIMIT 1
					FOR UPDATE SKIP LOCKED;
				`),
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// t.Parallel()
				if tt.got != tt.want {
					t.Errorf("ToSql() = %q; want %q", tt.got, tt.want)
				}
			})
		}
	})

	t.Run("CTE", func(t *testing.T) {
		t.Parallel()
		expected := sqltesting.Compact(`
			WITH recent_orders AS (SELECT "user_id", "amount" FROM "orders" WHERE "created_at" >= '2025-01-01')
			SELECT
			    "user_id", SUM("amount") AS "total"
			FROM
			    "recent_orders"
			GROUP BY
			    "user_id";
		`)

		sub := From("orders").
			Select("user_id", "amount").
			Where(Gte("created_at", "'2025-01-01'"))

		got := sqltesting.Compact(
			From("recent_orders").
				With("recent_orders", sub).
				Select("user_id", As(Fn("SUM", "amount"), "total")).
				GroupBy("user_id").
				ToSql(),
		)

		if got != expected {
			t.Errorf("ToSql() = %q; want %q", got, expected)
		}
	})

	t.Run("UNION", func(t *testing.T) {
		t.Parallel()
		tests := []struct {
			name string
			got  string
			want string
		}{
			{
				"Union",
				sqltesting.Compact(
					From("users").Select("id", "name").
						Union(From("archived_users").Select("id", "name")).
						ToSql(),
				),
				sqltesting.Compact(`
					SELECT
					    "id", "name"
					FROM
					    "users"
					UNION SELECT
					    "id", "name"
					FROM
					    "archived_users";
				`),
			},
			{
				"UnionAll",
				sqltesting.Compact(
					From("users").Select("id", "name").
						UnionAll(From("archived_users").Select("id", "name")).
						ToSql(),
				),
				sqltesting.Compact(`
					SELECT
					    "id", "name"
					FROM
					    "users"
					UNION ALL SELECT
					    "id", "name"
					FROM
					    "archived_users";
				`),
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// t.Parallel()
				if tt.got != tt.want {
					t.Errorf("ToSql() = %q; want %q", tt.got, tt.want)
				}
			})
		}
	})
}

func TestSqbl_EntryPoints(t *testing.T) {
	t.Parallel()

	t.Run("Select", func(t *testing.T) {
		t.Parallel()
		got := Select("1").ToSql()
		if !strings.Contains(got, "SELECT") {
			t.Errorf("Select: got %q", got)
		}
	})

	t.Run("InsertInto", func(t *testing.T) {
		t.Parallel()
		got := InsertInto("users").
			Columns("name").
			Values(syntax.P(1)).
			ToSql()
		if !strings.Contains(got, `INSERT INTO "users"`) {
			t.Errorf("InsertInto: got %q", got)
		}
	})

	t.Run("Update", func(t *testing.T) {
		t.Parallel()
		got := Update("users").
			Set("name", syntax.P(1)).
			Where(Eq("id", syntax.P(2))).
			ToSql()
		if !strings.Contains(got, `UPDATE "users"`) {
			t.Errorf("Update: got %q", got)
		}
	})

	t.Run("DeleteFrom", func(t *testing.T) {
		t.Parallel()
		got := DeleteFrom("users").
			Where(Eq("id", syntax.P(1))).
			ToSql()
		if !strings.Contains(got, `DELETE FROM "users"`) {
			t.Errorf("DeleteFrom: got %q", got)
		}
	})

	t.Run("CreateTable", func(t *testing.T) {
		t.Parallel()
		got := CreateTable("users").ToSql()
		want := `CREATE TABLE "users" ();`
		if got != want {
			t.Errorf("CreateTable: got %q, want %q", got, want)
		}
	})

	t.Run("CreateIndex", func(t *testing.T) {
		t.Parallel()
		got := CreateIndex("idx_users_email").On("users").Columns("email").ToSql()
		want := `CREATE INDEX "idx_users_email" ON "users" ("email");`
		if got != want {
			t.Errorf("CreateIndex: got %q, want %q", got, want)
		}
	})

	t.Run("DropTable", func(t *testing.T) {
		t.Parallel()
		got := DropTable("users").IfExists().ToSql()
		want := `DROP TABLE IF EXISTS "users";`
		if got != want {
			t.Errorf("DropTable: got %q, want %q", got, want)
		}
	})

	t.Run("DropIndex", func(t *testing.T) {
		t.Parallel()
		got := DropIndex("idx_users_email").IfExists().ToSql()
		want := `DROP INDEX IF EXISTS "idx_users_email";`
		if got != want {
			t.Errorf("DropIndex: got %q, want %q", got, want)
		}
	})

	t.Run("AlterTable", func(t *testing.T) {
		t.Parallel()
		got := AlterTable("users").AddColumn("bio", "TEXT").ToSql()
		want := `ALTER TABLE "users" ADD COLUMN "bio" TEXT;`
		if got != want {
			t.Errorf("AlterTable: got %q, want %q", got, want)
		}
	})
}
