package builder_test

import (
	"testing"

	"github.com/kotofurumiya/sqbl/internal/sqltesting"
	"github.com/kotofurumiya/sqbl/syntax"
)

func TestSqlCreateIndexBuilder_Clauses(t *testing.T) {
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
			name: "columns accumulate across calls",
			b:    idx.Columns("tenant_id", "created_at"),
			want: `CREATE INDEX "idx_users_email" ON "users" ("email", "tenant_id", "created_at");`,
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

// TestSqlCreateIndexBuilder_BranchIsolation checks that deriving two indexes
// from a shared base does not leak columns between them. The base holds three
// columns so that a backing array with spare capacity would be shared if
// Columns appended in place.
func TestSqlCreateIndexBuilder_BranchIsolation(t *testing.T) {
	t.Parallel()
	base := newCreateIndex().Name("idx").On("t").Columns("a").Columns("b").Columns("c")
	left := base.Columns("l")
	right := base.Columns("r")

	sqltesting.AssertSQL(t, base.ToSql(), `CREATE INDEX "idx" ON "t" ("a", "b", "c");`)
	sqltesting.AssertSQL(t, left.ToSql(), `CREATE INDEX "idx" ON "t" ("a", "b", "c", "l");`)
	sqltesting.AssertSQL(t, right.ToSql(), `CREATE INDEX "idx" ON "t" ("a", "b", "c", "r");`)
}
