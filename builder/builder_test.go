package builder_test

import (
	"testing"

	"github.com/kotofurumiya/sqbl/builder"
	"github.com/kotofurumiya/sqbl/dialect"
	"github.com/kotofurumiya/sqbl/internal/sqltesting"
)

// Builder tests render with PostgreSQL so that expected SQL reads like real SQL.
// How each dialect quotes and numbers placeholders is covered in dialect/.
var pg = &dialect.Postgres{}

func newSelect() builder.SelectBuilder { return (builder.SelectBuilder{}).Dialect(pg) }
func newInsert() builder.InsertBuilder { return (builder.InsertBuilder{}).Dialect(pg) }
func newUpdate() builder.UpdateBuilder { return (builder.UpdateBuilder{}).Dialect(pg) }
func newDelete() builder.DeleteBuilder { return (builder.DeleteBuilder{}).Dialect(pg) }

func newCreateTable() builder.CreateTableBuilder {
	return (builder.CreateTableBuilder{}).Dialect(pg)
}

func newCreateIndex() builder.CreateIndexBuilder {
	return (builder.CreateIndexBuilder{}).Dialect(pg)
}

func newAlterTable() builder.AlterTableBuilder {
	return (builder.AlterTableBuilder{}).Dialect(pg)
}

func newDropTable() builder.DropTableBuilder { return (builder.DropTableBuilder{}).Dialect(pg) }
func newDropIndex() builder.DropIndexBuilder { return (builder.DropIndexBuilder{}).Dialect(pg) }

func newCreateDatabase() builder.CreateDatabaseBuilder {
	return (builder.CreateDatabaseBuilder{}).Dialect(pg)
}

func newDropDatabase() builder.DropDatabaseBuilder {
	return (builder.DropDatabaseBuilder{}).Dialect(pg)
}

// sqlCase is the common shape of table-driven builder tests.
type sqlCase struct {
	name string
	b    builder.Statement
	want string
}

func runSQLCases(t *testing.T, tests []sqlCase) {
	t.Helper()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sqltesting.AssertSQL(t, tt.b.MustSQL(), tt.want)
		})
	}
}

// TestBuilder_RenderContract checks the two render methods every builder shares:
// ToSQL uses the builder's own dialect and ends with a semicolon, while
// ToSQLWithDialect uses the given dialect and has no semicolon.
func TestBuilder_RenderContract(t *testing.T) {
	t.Parallel()
	mysql := &dialect.MySQL{}

	tests := []struct {
		name      string
		b         builder.Statement
		wantToSQL string
		wantMySQL string
	}{
		{"select", newSelect().From("users").Select("id"), `SELECT "id" FROM "users";`, "SELECT `id` FROM `users`"},
		{"insert", newInsert().Into("users").Columns("id").Values(1), `INSERT INTO "users" ("id") VALUES (1);`, "INSERT INTO `users` (`id`) VALUES (1)"},
		{"update", newUpdate().Table("users").Set("id", 1), `UPDATE "users" SET "id" = 1;`, "UPDATE `users` SET `id` = 1"},
		{"delete", newDelete().From("users"), `DELETE FROM "users";`, "DELETE FROM `users`"},
		{"create table", newCreateTable().Table("users").Column("id", "INT"), `CREATE TABLE "users" ("id" INT);`, "CREATE TABLE `users` (`id` INT)"},
		{"create index", newCreateIndex().Name("idx").On("users").Columns("id"), `CREATE INDEX "idx" ON "users" ("id");`, "CREATE INDEX `idx` ON `users` (`id`)"},
		{"alter table", newAlterTable().Table("users").DropColumn("id"), `ALTER TABLE "users" DROP COLUMN "id";`, "ALTER TABLE `users` DROP COLUMN `id`"},
		{"drop table", newDropTable().Table("users"), `DROP TABLE "users";`, "DROP TABLE `users`"},
		{"drop index", newDropIndex().Name("idx"), `DROP INDEX "idx";`, "DROP INDEX `idx`"},
		{"create database", newCreateDatabase().Name("app"), `CREATE DATABASE "app";`, "CREATE DATABASE `app`"},
		{"drop database", newDropDatabase().Name("app"), `DROP DATABASE "app";`, "DROP DATABASE `app`"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sqltesting.AssertSQL(t, tt.b.MustSQL(), tt.wantToSQL)
			sqltesting.AssertSQL(t, tt.b.MustSQLWithDialect(mysql), tt.wantMySQL)
		})
	}
}

// TestBuilder_FailsWithoutDialect checks that a missing dialect is reported
// instead of rendering SQL with no quoting rules, and that the Must variants
// turn that report into a panic.
func TestBuilder_FailsWithoutDialect(t *testing.T) {
	t.Parallel()

	builders := []struct {
		name string
		b    builder.Statement
	}{
		{"select", builder.SelectBuilder{}},
		{"insert", builder.InsertBuilder{}},
		{"update", builder.UpdateBuilder{}},
		{"delete", builder.DeleteBuilder{}},
		{"create table", builder.CreateTableBuilder{}},
		{"create index", builder.CreateIndexBuilder{}},
		{"alter table", builder.AlterTableBuilder{}},
		{"drop table", builder.DropTableBuilder{}},
		{"drop index", builder.DropIndexBuilder{}},
		{"create database", builder.CreateDatabaseBuilder{}},
		{"drop database", builder.DropDatabaseBuilder{}},
	}
	for _, tt := range builders {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if _, err := tt.b.ToSQL(); err == nil {
				t.Error("ToSQL: want an error")
			}
			if _, err := tt.b.ToSQLWithDialect(nil); err == nil {
				t.Error("ToSQLWithDialect(nil): want an error")
			}
			assertPanics(t, func() { tt.b.MustSQL() })
			assertPanics(t, func() { tt.b.MustSQLWithDialect(nil) })
		})
	}
}

func assertPanics(t *testing.T, f func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Error("did not panic")
		}
	}()
	f()
}

// TestBuilder_CopiesCallerSlices checks that builders keep their own copy of
// slices passed to variadic methods, so reusing a slice cannot rewrite a query
// that was already built.
func TestBuilder_CopiesCallerSlices(t *testing.T) {
	t.Parallel()
	cols := []string{"a", "b"}
	vals := []any{1, 2}

	tests := []sqlCase{
		{"select distinct on", newSelect().From("t").DistinctOn(cols...), `SELECT DISTINCT ON ("a", "b") * FROM "t";`},
		{"insert columns and values", newInsert().Into("t").Columns(cols...).Values(vals...), `INSERT INTO "t" ("a", "b") VALUES (1, 2);`},
		{"insert on conflict", newInsert().Into("t").Values(1).OnConflict(cols...).DoUpdate(`"a" = 1`), `INSERT INTO "t" VALUES (1) ON CONFLICT ("a", "b") DO UPDATE SET "a" = 1;`},
		{"insert returning", newInsert().Into("t").Values(1).Returning(cols...), `INSERT INTO "t" VALUES (1) RETURNING "a", "b";`},
		{"update returning", newUpdate().Table("t").Set("a", 1).Returning(cols...), `UPDATE "t" SET "a" = 1 RETURNING "a", "b";`},
		{"delete returning", newDelete().From("t").Returning(cols...), `DELETE FROM "t" RETURNING "a", "b";`},
		{"create table column constraints", newCreateTable().Table("t").Column("a", "INT", cols...), `CREATE TABLE "t" ("a" INT a b);`},
		{
			name: "create table constraints",
			b:    newCreateTable().Table("t").Column("a", "INT").PrimaryKey(cols...).Unique(cols...).ForeignKey(cols, "r", cols, cols...),
			want: `CREATE TABLE "t" ("a" INT, PRIMARY KEY ("a", "b"), UNIQUE ("a", "b"), FOREIGN KEY ("a", "b") REFERENCES "r" ("a", "b") a b);`,
		},
	}
	cols[0], cols[1] = "x", "y"
	vals[0], vals[1] = 8, 9
	runSQLCases(t, tests)
}
