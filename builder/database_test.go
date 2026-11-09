package builder_test

import "testing"

func TestSqlCreateDatabaseBuilder_Clauses(t *testing.T) {
	t.Parallel()
	app := newCreateDatabase().Name("app_test")

	runSQLCases(t, []sqlCase{
		{"basic", app, `CREATE DATABASE "app_test";`},
		{"if not exists", app.IfNotExists(), `CREATE DATABASE IF NOT EXISTS "app_test";`},
		// A database name has no schema qualifier, so a dot is part of the name.
		{"dot is part of the name", newCreateDatabase().Name("app.test"), `CREATE DATABASE "app.test";`},
	})
}

func TestSqlDropDatabaseBuilder_Clauses(t *testing.T) {
	t.Parallel()
	app := newDropDatabase().Name("app_test")

	runSQLCases(t, []sqlCase{
		{"basic", app, `DROP DATABASE "app_test";`},
		{"if exists", app.IfExists(), `DROP DATABASE IF EXISTS "app_test";`},
		{"force", app.Force(), `DROP DATABASE "app_test" WITH (FORCE);`},
		{"all clauses", app.Force().IfExists(), `DROP DATABASE IF EXISTS "app_test" WITH (FORCE);`},
		{"dot is part of the name", newDropDatabase().Name("app.test"), `DROP DATABASE "app.test";`},
	})
}
