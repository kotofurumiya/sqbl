package builder_test

import (
	"testing"

	"github.com/kotofurumiya/sqbl/internal/sqltesting"
	"github.com/kotofurumiya/sqbl/syntax"
)

func TestCreateIndexBuilder_Clauses(t *testing.T) {
	t.Parallel()
	idx := newCreateIndex().Name("idx_users_email").On("users").Columns("email")

	runSQLCases(t, []sqlCase{
		{
			name: "basic",
			b:    idx,
			want: `CREATE INDEX "idx_users_email" ON "users" ("email");`,
		},
		{
			name: "unique if not exists",
			b:    idx.Unique().IfNotExists(),
			want: `CREATE UNIQUE INDEX IF NOT EXISTS "idx_users_email" ON "users" ("email");`,
		},
		{
			name: "columns are replaced by a later call",
			b:    idx.Columns("tenant_id", "created_at"),
			want: `CREATE INDEX "idx_users_email" ON "users" ("tenant_id", "created_at");`,
		},
		{
			name: "columns accept directions and expressions",
			b:    idx.Columns(syntax.Desc("created_at"), syntax.Raw(`(LOWER("email"))`)),
			want: `CREATE INDEX "idx_users_email" ON "users" ("created_at" DESC, (LOWER("email")));`,
		},
		{
			name: "partial index predicates accumulate with and",
			b:    idx.Where(syntax.IsNull("deleted_at")).Where(nil).Where(syntax.Eq("active", true)),
			want: `CREATE INDEX "idx_users_email" ON "users" ("email") WHERE "deleted_at" IS NULL AND "active" = TRUE;`,
		},
		{
			name: "using method is written verbatim",
			b:    idx.Using("btree"),
			want: `CREATE INDEX "idx_users_email" ON "users" USING btree ("email");`,
		},
		{
			name: "partial index",
			b:    idx.Where(syntax.IsNull("deleted_at")),
			want: `CREATE INDEX "idx_users_email" ON "users" ("email") WHERE "deleted_at" IS NULL;`,
		},
		{
			name: "all clauses",
			b:    idx.Where(syntax.IsNull("deleted_at")).Using("hash").IfNotExists().Unique(),
			want: `CREATE UNIQUE INDEX IF NOT EXISTS "idx_users_email" ON "users" USING hash ("email") WHERE "deleted_at" IS NULL;`,
		},
	})
}

// TestCreateIndexBuilder_BranchIsolation checks that deriving two indexes
// from a shared base does not leak predicates between them. The base holds
// three predicates so that a backing array with spare capacity would be shared
// if Where appended in place.
func TestCreateIndexBuilder_BranchIsolation(t *testing.T) {
	t.Parallel()
	base := newCreateIndex().Name("idx").On("t").Columns("a").
		Where(syntax.Eq("w", 1)).Where(syntax.Eq("w", 2)).Where(syntax.Eq("w", 3))
	left := base.Where(syntax.Eq("l", 1))
	right := base.Where(syntax.Eq("r", 1))

	sqltesting.AssertSQL(t, base.MustSQL(), `CREATE INDEX "idx" ON "t" ("a") WHERE "w" = 1 AND "w" = 2 AND "w" = 3;`)
	sqltesting.AssertSQL(t, left.MustSQL(), `CREATE INDEX "idx" ON "t" ("a") WHERE "w" = 1 AND "w" = 2 AND "w" = 3 AND "l" = 1;`)
	sqltesting.AssertSQL(t, right.MustSQL(), `CREATE INDEX "idx" ON "t" ("a") WHERE "w" = 1 AND "w" = 2 AND "w" = 3 AND "r" = 1;`)
}
