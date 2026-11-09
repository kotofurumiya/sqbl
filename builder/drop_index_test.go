package builder_test

import "testing"

func TestDropIndexBuilder_Clauses(t *testing.T) {
	t.Parallel()
	idx := newDropIndex().Name("idx_users_email")

	runSQLCases(t, []sqlCase{
		{"basic", idx, `DROP INDEX "idx_users_email";`},
		{"if exists", idx.IfExists(), `DROP INDEX IF EXISTS "idx_users_email";`},
		{"on table", idx.On("users"), `DROP INDEX "idx_users_email" ON "users";`},
		{"all clauses", idx.On("users").IfExists(), `DROP INDEX IF EXISTS "idx_users_email" ON "users";`},
	})
}
