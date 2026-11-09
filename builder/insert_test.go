package builder_test

import (
	"testing"

	"github.com/kotofurumiya/sqbl/internal/sqltesting"
	"github.com/kotofurumiya/sqbl/syntax"
)

func TestSqlInsertBuilder_Values(t *testing.T) {
	t.Parallel()
	users := newInsert().Into("users").Columns("name", "email")

	runSQLCases(t, []sqlCase{
		{
			name: "single row",
			b:    users.Values(syntax.P(1), syntax.P(2)),
			want: `INSERT INTO "users" ("name", "email") VALUES ($1, $2);`,
		},
		{
			name: "rows accumulate in call order",
			b:    users.Values(syntax.P(1), syntax.P(2)).Values(syntax.P(3), syntax.P(4)),
			want: `INSERT INTO "users" ("name", "email") VALUES ($1, $2), ($3, $4);`,
		},
		{
			name: "no column list",
			b:    newInsert().Into("users").Values(syntax.P(1), syntax.P(2)),
			want: `INSERT INTO "users" VALUES ($1, $2);`,
		},
		{
			name: "columns are replaced by a later call",
			b:    users.Columns("id").Values(syntax.P(1)),
			want: `INSERT INTO "users" ("id") VALUES ($1);`,
		},
		{
			// A plain string is an identifier even in a value position, so a
			// string literal has to be passed with Raw or as a parameter.
			name: "value kinds",
			b: newInsert().Into("t").Values(
				syntax.P(), syntax.P(":name"), 42, 1.5, true, false, nil,
				"other_col", syntax.Raw("'text'"), syntax.Raw("DEFAULT"), syntax.Fn("NOW"),
			),
			want: `INSERT INTO "t" VALUES (?, :name, 42, 1.5, TRUE, FALSE, NULL, "other_col", 'text', DEFAULT, NOW());`,
		},
		{
			name: "no rows renders no VALUES clause",
			b:    newInsert().Into("users").Columns("name"),
			want: `INSERT INTO "users" ("name");`,
		},
		{
			name: "returning",
			b:    users.Values(syntax.P(1), syntax.P(2)).Returning("id", "created_at"),
			want: `INSERT INTO "users" ("name", "email") VALUES ($1, $2) RETURNING "id", "created_at";`,
		},
	})
}

func TestSqlInsertBuilder_ConflictHandling(t *testing.T) {
	t.Parallel()
	tags := newInsert().Into("tags").Columns("name", "count").Values(syntax.P(1), syntax.P(2))

	runSQLCases(t, []sqlCase{
		// ON CONFLICT (PostgreSQL / SQLite)
		{
			name: "do nothing without target",
			b:    tags.OnConflictDoNothing(),
			want: `INSERT INTO "tags" ("name", "count") VALUES ($1, $2) ON CONFLICT DO NOTHING;`,
		},
		{
			name: "do nothing with target and index predicate",
			b:    tags.OnConflict("name").OnConflictWhere(syntax.Eq("active", true)).OnConflictDoNothing(),
			want: `INSERT INTO "tags" ("name", "count") VALUES ($1, $2) ON CONFLICT ("name") WHERE "active" = TRUE DO NOTHING;`,
		},
		{
			// An index predicate without target columns is invalid SQL.
			name: "index predicate is dropped without target",
			b:    tags.OnConflictWhere(syntax.Eq("active", true)).OnConflictDoNothing(),
			want: `INSERT INTO "tags" ("name", "count") VALUES ($1, $2) ON CONFLICT DO NOTHING;`,
		},
		{
			name: "target alone renders nothing",
			b:    tags.OnConflict("name"),
			want: `INSERT INTO "tags" ("name", "count") VALUES ($1, $2);`,
		},
		{
			name: "do update with raw and structured assignments in call order",
			b: tags.OnConflict("name", "kind").
				DoUpdate(`"count" = "tags"."count" + 1`, `"seen" = TRUE`).
				DoUpdateSet("name", "excluded.name").
				DoUpdateSet("note", syntax.P(3)),
			want: `
				INSERT INTO "tags" ("name", "count") VALUES ($1, $2)
				ON CONFLICT ("name", "kind") DO UPDATE SET
				"count" = "tags"."count" + 1, "seen" = TRUE, "name" = "excluded"."name", "note" = $3;`,
		},
		{
			name: "do update with index predicate and update condition",
			b: tags.OnConflict("name").
				OnConflictWhere(syntax.Eq("active", true)).
				DoUpdateSet("count", "excluded.count").
				DoUpdateWhere(syntax.Lt("tags.count", "excluded.count")),
			want: `
				INSERT INTO "tags" ("name", "count") VALUES ($1, $2)
				ON CONFLICT ("name") WHERE "active" = TRUE
				DO UPDATE SET "count" = "excluded"."count"
				WHERE "tags"."count" < "excluded"."count";`,
		},
		{
			name: "update condition alone renders nothing",
			b:    tags.OnConflict("name").DoUpdateWhere(syntax.Eq("a", 1)),
			want: `INSERT INTO "tags" ("name", "count") VALUES ($1, $2);`,
		},

		// ON DUPLICATE KEY UPDATE (MySQL)
		{
			name: "on duplicate key update with raw and structured assignments in call order",
			b: tags.
				OnDuplicateKeyUpdate(`"count" = "count" + 1`).
				OnDuplicateKeyUpdateSet("name", syntax.Fn("VALUES", "name")),
			want: `
				INSERT INTO "tags" ("name", "count") VALUES ($1, $2)
				ON DUPLICATE KEY UPDATE "count" = "count" + 1, "name" = VALUES("name");`,
		},

		// INSERT OR ... (SQLite)
		{"or replace", tags.OrReplace(), `INSERT OR REPLACE INTO "tags" ("name", "count") VALUES ($1, $2);`},
		{"or ignore", tags.OrIgnore(), `INSERT OR IGNORE INTO "tags" ("name", "count") VALUES ($1, $2);`},
		{"or abort", tags.OrAbort(), `INSERT OR ABORT INTO "tags" ("name", "count") VALUES ($1, $2);`},
		{"or fail", tags.OrFail(), `INSERT OR FAIL INTO "tags" ("name", "count") VALUES ($1, $2);`},
		{"or rollback", tags.OrRollback(), `INSERT OR ROLLBACK INTO "tags" ("name", "count") VALUES ($1, $2);`},
		{"later algorithm wins", tags.OrReplace().OrIgnore(), `INSERT OR IGNORE INTO "tags" ("name", "count") VALUES ($1, $2);`},
	})
}

// TestSqlInsertBuilder_AllClauses pins the clause order. sqbl does not validate
// combinations, so clauses from different databases are rendered side by side.
func TestSqlInsertBuilder_AllClauses(t *testing.T) {
	t.Parallel()
	got := newInsert().
		Returning("id").
		OnDuplicateKeyUpdateSet("count", 0).
		DoUpdateWhere(syntax.Gt("count", 0)).
		DoUpdateSet("count", 1).
		OnConflictWhere(syntax.Eq("active", true)).
		OnConflict("name").
		Values(syntax.P(1), syntax.P(2)).
		Columns("name", "count").
		Into("tags").
		OrIgnore().
		ToSql()

	sqltesting.AssertSQL(t, got, `
		INSERT OR IGNORE INTO "tags" ("name", "count") VALUES ($1, $2)
		ON CONFLICT ("name") WHERE "active" = TRUE DO UPDATE SET "count" = 1 WHERE "count" > 0
		ON DUPLICATE KEY UPDATE "count" = 0
		RETURNING "id";`)
}

// TestSqlInsertBuilder_BranchIsolation checks that deriving two statements from
// a shared base does not leak appended rows or assignments between them.
// The base holds three items per list so that a backing array with spare
// capacity would be shared if a method appended in place.
func TestSqlInsertBuilder_BranchIsolation(t *testing.T) {
	t.Parallel()
	base := newInsert().Into("t").
		Values(1).Values(2).Values(3).
		DoUpdateSet("a", 1).DoUpdate(`"b" = 2`).DoUpdateSet("c", 3).
		OnDuplicateKeyUpdateSet("a", 1).OnDuplicateKeyUpdate(`"b" = 2`).OnDuplicateKeyUpdateSet("c", 3)

	left := base.Values(4).DoUpdateSet("l", 4).OnDuplicateKeyUpdate(`"l" = 4`)
	right := base.Values(5).DoUpdate(`"r" = 5`).OnDuplicateKeyUpdateSet("r", 5)

	sqltesting.AssertSQL(t, base.ToSql(), `
		INSERT INTO "t" VALUES (1), (2), (3)
		ON CONFLICT DO UPDATE SET "a" = 1, "b" = 2, "c" = 3
		ON DUPLICATE KEY UPDATE "a" = 1, "b" = 2, "c" = 3;`)
	sqltesting.AssertSQL(t, left.ToSql(), `
		INSERT INTO "t" VALUES (1), (2), (3), (4)
		ON CONFLICT DO UPDATE SET "a" = 1, "b" = 2, "c" = 3, "l" = 4
		ON DUPLICATE KEY UPDATE "a" = 1, "b" = 2, "c" = 3, "l" = 4;`)
	sqltesting.AssertSQL(t, right.ToSql(), `
		INSERT INTO "t" VALUES (1), (2), (3), (5)
		ON CONFLICT DO UPDATE SET "a" = 1, "b" = 2, "c" = 3, "r" = 5
		ON DUPLICATE KEY UPDATE "a" = 1, "b" = 2, "c" = 3, "r" = 5;`)
}
