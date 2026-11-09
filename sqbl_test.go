package sqbl_test

import (
	"strings"
	"testing"

	"github.com/kotofurumiya/sqbl"
	"github.com/kotofurumiya/sqbl/builder"
	"github.com/kotofurumiya/sqbl/dialect"
	"github.com/kotofurumiya/sqbl/internal/sqltesting"
)

// One dialect-free builder must render each dialect's quoting, placeholders,
// and boolean literals when given that dialect.
func TestRendersPerDialect(t *testing.T) {
	t.Parallel()
	q := sqbl.From("users").
		Select("u.id").
		Where(sqbl.And(sqbl.Eq("u.active", true), sqbl.Eq("u.id", sqbl.P(1))))

	tests := []struct {
		name string
		d    dialect.SqlDialect
		want string
	}{
		{"PostgreSQL", &dialect.PostgresDialect{}, `SELECT "u"."id" FROM "users" WHERE "u"."active" = TRUE AND "u"."id" = $1`},
		{"MySQL", &dialect.MysqlDialect{}, "SELECT `u`.`id` FROM `users` WHERE `u`.`active` = TRUE AND `u`.`id` = ?"},
		{"SQLite", &dialect.SqliteDialect{}, `SELECT "u"."id" FROM "users" WHERE "u"."active" = 1 AND "u"."id" = ?1`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sqltesting.AssertSQL(t, q.ToSqlWithDialect(tt.d), tt.want)
		})
	}
}

func TestEntryPoints(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		b    builder.SqlBuilder
		want string
	}{
		{"From", sqbl.From("users"), `SELECT * FROM "users"`},
		{"Select", sqbl.Select("id"), `SELECT "id"`},
		{"InsertInto", sqbl.InsertInto("users").Columns("name").Values(sqbl.P(1)), `INSERT INTO "users" ("name") VALUES ($1)`},
		{"Update", sqbl.Update("users").Set("name", sqbl.P(1)), `UPDATE "users" SET "name" = $1`},
		{"DeleteFrom", sqbl.DeleteFrom("users"), `DELETE FROM "users"`},
		{"CreateTable", sqbl.CreateTable("users").Column("id", "BIGINT"), `CREATE TABLE "users" ("id" BIGINT)`},
		{"CreateIndex", sqbl.CreateIndex("idx").On("users").Columns("email"), `CREATE INDEX "idx" ON "users" ("email")`},
		{"DropTable", sqbl.DropTable("users"), `DROP TABLE "users"`},
		{"DropIndex", sqbl.DropIndex("idx"), `DROP INDEX "idx"`},
		{"AlterTable", sqbl.AlterTable("users").AddColumn("bio", "TEXT"), `ALTER TABLE "users" ADD COLUMN "bio" TEXT`},
		{"CreateDatabase", sqbl.CreateDatabase("app"), `CREATE DATABASE "app"`},
		{"DropDatabase", sqbl.DropDatabase("app"), `DROP DATABASE "app"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sqltesting.AssertSQL(t, tt.b.ToSqlWithDialect(&dialect.PostgresDialect{}), tt.want)
		})
	}
}

// Rendering without a dialect is a programming error, so every builder must
// panic with a hint pointing at ToSqlWithDialect instead of emitting broken SQL.
func TestPanicsWithoutDialect(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		b    builder.SqlBuilder
	}{
		{"From", sqbl.From("users")},
		{"InsertInto", sqbl.InsertInto("users")},
		{"Update", sqbl.Update("users")},
		{"DeleteFrom", sqbl.DeleteFrom("users")},
		{"CreateTable", sqbl.CreateTable("users")},
		{"CreateIndex", sqbl.CreateIndex("idx")},
		{"DropTable", sqbl.DropTable("users")},
		{"DropIndex", sqbl.DropIndex("idx")},
		{"AlterTable", sqbl.AlterTable("users")},
		{"CreateDatabase", sqbl.CreateDatabase("app")},
		{"DropDatabase", sqbl.DropDatabase("app")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertDialectPanic(t, "ToSql", func() { tt.b.ToSql() })
			assertDialectPanic(t, "ToSqlWithDialect(nil)", func() { tt.b.ToSqlWithDialect(nil) })
		})
	}
}

func assertDialectPanic(t *testing.T, call string, f func()) {
	t.Helper()
	defer func() {
		t.Helper()
		msg, ok := recover().(string)
		if !ok || !strings.Contains(msg, "ToSqlWithDialect") {
			t.Errorf("%s: want a panic mentioning ToSqlWithDialect, got %q", call, msg)
		}
	}()
	f()
}
