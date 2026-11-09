package builder_test

import (
	"testing"

	"github.com/kotofurumiya/sqbl/internal/sqltesting"
	"github.com/kotofurumiya/sqbl/syntax"
)

func TestCreateTableBuilder_Clauses(t *testing.T) {
	t.Parallel()
	users := newCreateTable().Table("users").Column("id", "BIGINT")

	runSQLCases(t, []sqlCase{
		{
			name: "columns with and without constraints",
			b:    newCreateTable().Table("users").Column("id", "BIGINT", "NOT NULL", "DEFAULT 0").Column("bio", "TEXT"),
			want: `CREATE TABLE "users" ("id" BIGINT NOT NULL DEFAULT 0, "bio" TEXT);`,
		},
		{
			name: "if not exists",
			b:    users.IfNotExists(),
			want: `CREATE TABLE IF NOT EXISTS "users" ("id" BIGINT);`,
		},
		{
			name: "primary key and unique, single and composite",
			b:    users.PrimaryKey("id").PrimaryKey("tenant_id", "id").Unique("email").Unique("tenant_id", "email"),
			want: `
				CREATE TABLE "users" ("id" BIGINT,
				PRIMARY KEY ("id"), PRIMARY KEY ("tenant_id", "id"), UNIQUE ("email"), UNIQUE ("tenant_id", "email"));`,
		},
		{
			name: "foreign keys with and without actions",
			b: users.
				ForeignKey([]string{"org_id"}, "orgs", []string{"id"}).
				ForeignKey([]string{"tenant_id", "team_id"}, "auth.teams", []string{"tenant_id", "id"}, "ON DELETE CASCADE", "ON UPDATE RESTRICT"),
			want: `
				CREATE TABLE "users" ("id" BIGINT,
				FOREIGN KEY ("org_id") REFERENCES "orgs" ("id"),
				FOREIGN KEY ("tenant_id", "team_id") REFERENCES "auth"."teams" ("tenant_id", "id") ON DELETE CASCADE ON UPDATE RESTRICT);`,
		},
		{
			name: "check",
			b:    users.Check(syntax.Gt("id", 0)),
			want: `CREATE TABLE "users" ("id" BIGINT, CHECK ("id" > 0));`,
		},
		{
			// Column definitions must precede table constraints in SQL,
			// so they are grouped regardless of call order.
			name: "table constraints follow all columns",
			b:    newCreateTable().Table("users").PrimaryKey("id").Column("id", "BIGINT").Check(syntax.Gt("id", 0)).Column("bio", "TEXT"),
			want: `CREATE TABLE "users" ("id" BIGINT, "bio" TEXT, PRIMARY KEY ("id"), CHECK ("id" > 0));`,
		},
		{
			name: "constraints alone",
			b:    newCreateTable().Table("users").PrimaryKey("id"),
			want: `CREATE TABLE "users" (PRIMARY KEY ("id"));`,
		},
		{
			name: "options are written verbatim and accumulate",
			b:    users.Options("WITH (fillfactor = 70)", "TABLESPACE fast").Options("STRICT, WITHOUT ROWID"),
			want: `CREATE TABLE "users" ("id" BIGINT) WITH (fillfactor = 70) TABLESPACE fast STRICT, WITHOUT ROWID;`,
		},
	})
}

// TestCreateTableBuilder_BranchIsolation checks that deriving two statements
// from a shared base does not leak columns, constraints, or options between them.
// The base holds three items per list so that a backing array with spare
// capacity would be shared if a method appended in place.
func TestCreateTableBuilder_BranchIsolation(t *testing.T) {
	t.Parallel()
	base := newCreateTable().Table("t").
		Column("a", "INT").Column("b", "INT").Column("c", "INT").
		Unique("a").Unique("b").Unique("c").
		Options("A").Options("B").Options("C")

	left := base.Column("l", "INT").PrimaryKey("l").Options("L")
	right := base.Column("r", "INT").Check(syntax.Gt("r", 0)).Options("R")

	sqltesting.AssertSQL(t, base.MustSQL(), `
		CREATE TABLE "t" ("a" INT, "b" INT, "c" INT, UNIQUE ("a"), UNIQUE ("b"), UNIQUE ("c")) A B C;`)
	sqltesting.AssertSQL(t, left.MustSQL(), `
		CREATE TABLE "t" ("a" INT, "b" INT, "c" INT, "l" INT,
		UNIQUE ("a"), UNIQUE ("b"), UNIQUE ("c"), PRIMARY KEY ("l")) A B C L;`)
	sqltesting.AssertSQL(t, right.MustSQL(), `
		CREATE TABLE "t" ("a" INT, "b" INT, "c" INT, "r" INT,
		UNIQUE ("a"), UNIQUE ("b"), UNIQUE ("c"), CHECK ("r" > 0)) A B C R;`)
}
