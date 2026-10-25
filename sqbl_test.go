package sqbl_test

import (
	"strings"
	"testing"

	"github.com/kotofurumiya/sqbl"
	"github.com/kotofurumiya/sqbl/builder"
	"github.com/kotofurumiya/sqbl/dialect"
)

func TestDialectFree_RendersPerDialect(t *testing.T) {
	t.Parallel()
	q := sqbl.From("users").Select("id").Where(sqbl.Eq("id", sqbl.P(1)))

	tests := []struct {
		name string
		d    dialect.SqlDialect
		want string
	}{
		{name: "PostgreSQL", d: &dialect.PostgresDialect{}, want: `SELECT "id" FROM "users" WHERE "id" = $1`},
		{name: "MySQL", d: &dialect.MysqlDialect{}, want: "SELECT `id` FROM `users` WHERE `id` = ?"},
		{name: "SQLite", d: &dialect.SqliteDialect{}, want: `SELECT "id" FROM "users" WHERE "id" = ?`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := q.ToSqlWithDialect(tt.d); got != tt.want {
				t.Errorf("got %q; want %q", got, tt.want)
			}
		})
	}
}

func TestDialectFree_EntryPoints(t *testing.T) {
	t.Parallel()
	d := &dialect.PostgresDialect{}

	tests := []struct {
		name string
		b    builder.SqlBuilder
		want string
	}{
		{name: "From", b: sqbl.From("users"), want: `SELECT * FROM "users"`},
		{name: "Select", b: sqbl.Select(sqbl.Fn("NOW")), want: `SELECT NOW()`},
		{name: "InsertInto", b: sqbl.InsertInto("users").Columns("name").Values(sqbl.P(1)), want: `INSERT INTO "users" ("name") VALUES ($1)`},
		{name: "Update", b: sqbl.Update("users").Set("name", sqbl.P(1)), want: `UPDATE "users" SET "name" = $1`},
		{name: "DeleteFrom", b: sqbl.DeleteFrom("users"), want: `DELETE FROM "users"`},
		{name: "CreateTable", b: sqbl.CreateTable("users").Column("id", "BIGINT"), want: `CREATE TABLE "users" ("id" BIGINT)`},
		{name: "CreateIndex", b: sqbl.CreateIndex("idx").On("users").Columns("email"), want: `CREATE INDEX "idx" ON "users" ("email")`},
		{name: "DropTable", b: sqbl.DropTable("users"), want: `DROP TABLE "users"`},
		{name: "DropIndex", b: sqbl.DropIndex("idx"), want: `DROP INDEX "idx"`},
		{name: "AlterTable", b: sqbl.AlterTable("users").AddColumn("bio", "TEXT"), want: `ALTER TABLE "users" ADD COLUMN "bio" TEXT`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.b.ToSqlWithDialect(d); got != tt.want {
				t.Errorf("got %q; want %q", got, tt.want)
			}
		})
	}
}

func TestDialectFree_PanicsWithoutDialect(t *testing.T) {
	t.Parallel()
	builders := []struct {
		name string
		b    builder.SqlBuilder
	}{
		{name: "Select", b: sqbl.From("users")},
		{name: "Insert", b: sqbl.InsertInto("users")},
		{name: "Update", b: sqbl.Update("users")},
		{name: "Delete", b: sqbl.DeleteFrom("users")},
		{name: "CreateTable", b: sqbl.CreateTable("users")},
		{name: "CreateIndex", b: sqbl.CreateIndex("idx")},
		{name: "DropTable", b: sqbl.DropTable("users")},
		{name: "DropIndex", b: sqbl.DropIndex("idx")},
		{name: "AlterTable", b: sqbl.AlterTable("users")},
	}
	for _, tt := range builders {
		t.Run(tt.name+"/ToSql", func(t *testing.T) {
			assertDialectPanic(t, func() { tt.b.ToSql() })
		})
		t.Run(tt.name+"/ToSqlWithDialect(nil)", func(t *testing.T) {
			assertDialectPanic(t, func() { tt.b.ToSqlWithDialect(nil) })
		})
	}
}

func assertDialectPanic(t *testing.T, f func()) {
	t.Helper()
	defer func() {
		t.Helper()
		r := recover()
		msg, ok := r.(string)
		if !ok || !strings.Contains(msg, "ToSqlWithDialect") {
			t.Errorf("expected a panic mentioning ToSqlWithDialect, got %v", r)
		}
	}()
	f()
}
