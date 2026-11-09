package builder_test

import "testing"

func TestSqlDropTableBuilder_Clauses(t *testing.T) {
	t.Parallel()
	users := newDropTable().Table("users")

	runSQLCases(t, []sqlCase{
		{"basic", users, `DROP TABLE "users";`},
		{"qualified table", newDropTable().Table("auth.users"), `DROP TABLE "auth"."users";`},
		{"if exists", users.IfExists(), `DROP TABLE IF EXISTS "users";`},
		{"cascade", users.Cascade(), `DROP TABLE "users" CASCADE;`},
		{"all clauses", users.Cascade().IfExists(), `DROP TABLE IF EXISTS "users" CASCADE;`},
	})
}
