package sqblpg_test

import (
	"testing"

	"github.com/kotofurumiya/sqbl/builder"
	"github.com/kotofurumiya/sqbl/internal/sqltesting"
	"github.com/kotofurumiya/sqbl/sqblpg"
)

func TestEntryPoints(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		b    builder.Statement
		want string
	}{
		{"From", sqblpg.From("users"), `SELECT * FROM "users";`},
		{"Select", sqblpg.Select("id"), `SELECT "id";`},
		{"InsertInto", sqblpg.InsertInto("users").Columns("name").Values(sqblpg.P(1)), `INSERT INTO "users" ("name") VALUES ($1);`},
		{"Update", sqblpg.Update("users").Set("name", sqblpg.P(1)), `UPDATE "users" SET "name" = $1;`},
		{"DeleteFrom", sqblpg.DeleteFrom("users"), `DELETE FROM "users";`},
		{"CreateTable", sqblpg.CreateTable("users").Column("id", "BIGINT"), `CREATE TABLE "users" ("id" BIGINT);`},
		{"CreateIndex", sqblpg.CreateIndex("idx").On("users").Columns("email"), `CREATE INDEX "idx" ON "users" ("email");`},
		{"DropTable", sqblpg.DropTable("users"), `DROP TABLE "users";`},
		{"DropIndex", sqblpg.DropIndex("idx"), `DROP INDEX "idx";`},
		{"AlterTable", sqblpg.AlterTable("users").AddColumn("bio", "TEXT"), `ALTER TABLE "users" ADD COLUMN "bio" TEXT;`},
		{"CreateDatabase", sqblpg.CreateDatabase("app"), `CREATE DATABASE "app";`},
		{"DropDatabase", sqblpg.DropDatabase("app"), `DROP DATABASE "app";`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sqltesting.AssertSQL(t, tt.b.ToSQL(), tt.want)
		})
	}
}

// A realistic query shows every PostgreSQL-specific rendering choice at once:
// dotted identifiers quoted per part, escaped quotes, $N placeholders, TRUE literals, and ILIKE.
func TestDialectRendering(t *testing.T) {
	t.Parallel()
	got := sqblpg.From(sqblpg.As("users", "u")).
		Select("u.id", `u.nick"name`, sqblpg.As(sqblpg.Fn("COUNT", "o.id"), "orders")).
		LeftJoin(sqblpg.As("orders", "o"), sqblpg.Eq("u.id", "o.user_id")).
		Where(sqblpg.And(
			sqblpg.Eq("u.active", true),
			sqblpg.ILike("u.name", sqblpg.P(1)),
			sqblpg.Gt("o.amount", sqblpg.P(2)),
		)).
		GroupBy("u.id").
		ToSQL()
	sqltesting.AssertSQL(t, got, `
		SELECT "u"."id", "u"."nick""name", COUNT("o"."id") AS "orders"
		FROM "users" AS "u"
		LEFT JOIN "orders" AS "o" ON "u"."id" = "o"."user_id"
		WHERE "u"."active" = TRUE AND "u"."name" ILIKE $1 AND "o"."amount" > $2
		GROUP BY "u"."id";
	`)
}
