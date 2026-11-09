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
