package sqblsqlite_test

import (
	"testing"

	"github.com/kotofurumiya/sqbl/builder"
	"github.com/kotofurumiya/sqbl/internal/sqltesting"
	"github.com/kotofurumiya/sqbl/sqblsqlite"
)

func TestEntryPoints(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		b    builder.SqlBuilder
		want string
	}{
		{"From", sqblsqlite.From("users"), `SELECT * FROM "users";`},
		{"Select", sqblsqlite.Select("id"), `SELECT "id";`},
		{"InsertInto", sqblsqlite.InsertInto("users").Columns("name").Values(sqblsqlite.P(1)), `INSERT INTO "users" ("name") VALUES (?1);`},
		{"Update", sqblsqlite.Update("users").Set("name", sqblsqlite.P(1)), `UPDATE "users" SET "name" = ?1;`},
		{"DeleteFrom", sqblsqlite.DeleteFrom("users"), `DELETE FROM "users";`},
		{"CreateTable", sqblsqlite.CreateTable("users").Column("id", "INTEGER"), `CREATE TABLE "users" ("id" INTEGER);`},
		{"CreateIndex", sqblsqlite.CreateIndex("idx").On("users").Columns("email"), `CREATE INDEX "idx" ON "users" ("email");`},
		{"DropTable", sqblsqlite.DropTable("users"), `DROP TABLE "users";`},
		{"DropIndex", sqblsqlite.DropIndex("idx"), `DROP INDEX "idx";`},
		{"AlterTable", sqblsqlite.AlterTable("users").AddColumn("bio", "TEXT"), `ALTER TABLE "users" ADD COLUMN "bio" TEXT;`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sqltesting.AssertSQL(t, tt.b.ToSql(), tt.want)
		})
	}
}

// A realistic query shows every SQLite-specific rendering choice at once:
// dotted identifiers quoted per part, escaped quotes, ?N placeholders, and booleans as 1/0.
func TestDialectRendering(t *testing.T) {
	t.Parallel()
	got := sqblsqlite.From(sqblsqlite.As("users", "u")).
		Select("u.id", `u.nick"name`, sqblsqlite.As(sqblsqlite.Fn("COUNT", "o.id"), "orders")).
		LeftJoin(sqblsqlite.As("orders", "o"), sqblsqlite.Eq("u.id", "o.user_id")).
		Where(sqblsqlite.And(
			sqblsqlite.Eq("u.active", true),
			sqblsqlite.Eq("u.banned", false),
			sqblsqlite.Like("u.name", sqblsqlite.P(1)),
			sqblsqlite.Gt("o.amount", sqblsqlite.P(2)),
		)).
		GroupBy("u.id").
		ToSql()
	sqltesting.AssertSQL(t, got, `
		SELECT "u"."id", "u"."nick""name", COUNT("o"."id") AS "orders"
		FROM "users" AS "u"
		LEFT JOIN "orders" AS "o" ON "u"."id" = "o"."user_id"
		WHERE "u"."active" = 1 AND "u"."banned" = 0 AND "u"."name" LIKE ?1 AND "o"."amount" > ?2
		GROUP BY "u"."id";
	`)
}
