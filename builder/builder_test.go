package builder_test

import (
	"testing"

	"github.com/kotofurumiya/sqbl/builder"
	"github.com/kotofurumiya/sqbl/dialect"
	"github.com/kotofurumiya/sqbl/internal/sqltesting"
)

// Builder tests render with PostgreSQL so that expected SQL reads like real SQL.
// How each dialect quotes and numbers placeholders is covered in dialect/.
var pg = &dialect.PostgresDialect{}

func newSelect() builder.SqlSelectBuilder { return (builder.SqlSelectBuilder{}).Dialect(pg) }
func newInsert() builder.SqlInsertBuilder { return (builder.SqlInsertBuilder{}).Dialect(pg) }
func newUpdate() builder.SqlUpdateBuilder { return (builder.SqlUpdateBuilder{}).Dialect(pg) }
func newDelete() builder.SqlDeleteBuilder { return (builder.SqlDeleteBuilder{}).Dialect(pg) }

func newCreateTable() builder.SqlCreateTableBuilder {
	return (builder.SqlCreateTableBuilder{}).Dialect(pg)
}

func newCreateIndex() builder.SqlCreateIndexBuilder {
	return (builder.SqlCreateIndexBuilder{}).Dialect(pg)
}

func newAlterTable() builder.SqlAlterTableBuilder {
	return (builder.SqlAlterTableBuilder{}).Dialect(pg)
}

func newDropTable() builder.SqlDropTableBuilder { return (builder.SqlDropTableBuilder{}).Dialect(pg) }
func newDropIndex() builder.SqlDropIndexBuilder { return (builder.SqlDropIndexBuilder{}).Dialect(pg) }

func newCreateDatabase() builder.SqlCreateDatabaseBuilder {
	return (builder.SqlCreateDatabaseBuilder{}).Dialect(pg)
}

func newDropDatabase() builder.SqlDropDatabaseBuilder {
	return (builder.SqlDropDatabaseBuilder{}).Dialect(pg)
}

// sqlCase is the common shape of table-driven builder tests.
type sqlCase struct {
	name string
	b    builder.SqlBuilder
	want string
}

func runSQLCases(t *testing.T, tests []sqlCase) {
	t.Helper()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sqltesting.AssertSQL(t, tt.b.ToSql(), tt.want)
		})
	}
}

// TestSqlBuilder_RenderContract checks the two render methods every builder shares:
// ToSql uses the builder's own dialect and ends with a semicolon, while
// ToSqlWithDialect uses the given dialect and has no semicolon.
func TestSqlBuilder_RenderContract(t *testing.T) {
	t.Parallel()
	mysql := &dialect.MysqlDialect{}

	tests := []struct {
		name      string
		b         builder.SqlBuilder
		wantToSql string
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
			sqltesting.AssertSQL(t, tt.b.ToSql(), tt.wantToSql)
			sqltesting.AssertSQL(t, tt.b.ToSqlWithDialect(mysql), tt.wantMySQL)
		})
	}
}

// TestSqlBuilder_PanicsWithoutDialect checks that a missing dialect fails loudly
// instead of rendering SQL with no quoting rules.
func TestSqlBuilder_PanicsWithoutDialect(t *testing.T) {
	t.Parallel()

	builders := []struct {
		name string
		b    builder.SqlBuilder
	}{
		{"select", builder.SqlSelectBuilder{}},
		{"insert", builder.SqlInsertBuilder{}},
		{"update", builder.SqlUpdateBuilder{}},
		{"delete", builder.SqlDeleteBuilder{}},
		{"create table", builder.SqlCreateTableBuilder{}},
		{"create index", builder.SqlCreateIndexBuilder{}},
		{"alter table", builder.SqlAlterTableBuilder{}},
		{"drop table", builder.SqlDropTableBuilder{}},
		{"drop index", builder.SqlDropIndexBuilder{}},
		{"create database", builder.SqlCreateDatabaseBuilder{}},
		{"drop database", builder.SqlDropDatabaseBuilder{}},
	}
	for _, tt := range builders {
		t.Run(tt.name+"/ToSql", func(t *testing.T) {
			t.Parallel()
			assertPanics(t, func() { tt.b.ToSql() })
		})
		t.Run(tt.name+"/ToSqlWithDialect(nil)", func(t *testing.T) {
			t.Parallel()
			assertPanics(t, func() { tt.b.ToSqlWithDialect(nil) })
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

// TestSqlBuilder_CopiesCallerSlices checks that builders keep their own copy of
// slices passed to variadic methods, so reusing a slice cannot rewrite a query
// that was already built.
func TestSqlBuilder_CopiesCallerSlices(t *testing.T) {
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
