package sqblmysql_test

import (
	"testing"

	"github.com/kotofurumiya/sqbl/builder"
	"github.com/kotofurumiya/sqbl/internal/sqltesting"
	"github.com/kotofurumiya/sqbl/sqblmysql"
)

func TestEntryPoints(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		b    builder.Statement
		want string
	}{
		{"From", sqblmysql.From("users"), "SELECT * FROM `users`;"},
		{"Select", sqblmysql.Select("id"), "SELECT `id`;"},
		{"InsertInto", sqblmysql.InsertInto("users").Columns("name").Values(sqblmysql.P(1)), "INSERT INTO `users` (`name`) VALUES (?);"},
		{"Update", sqblmysql.Update("users").Set("name", sqblmysql.P(1)), "UPDATE `users` SET `name` = ?;"},
		{"DeleteFrom", sqblmysql.DeleteFrom("users"), "DELETE FROM `users`;"},
		{"CreateTable", sqblmysql.CreateTable("users").Column("id", "BIGINT"), "CREATE TABLE `users` (`id` BIGINT);"},
		{"CreateIndex", sqblmysql.CreateIndex("idx").On("users").Columns("email"), "CREATE INDEX `idx` ON `users` (`email`);"},
		{"DropTable", sqblmysql.DropTable("users"), "DROP TABLE `users`;"},
		{"DropIndex", sqblmysql.DropIndex("idx").On("users"), "DROP INDEX `idx` ON `users`;"},
		{"AlterTable", sqblmysql.AlterTable("users").AddColumn("bio", "TEXT"), "ALTER TABLE `users` ADD COLUMN `bio` TEXT;"},
		{"CreateDatabase", sqblmysql.CreateDatabase("app"), "CREATE DATABASE `app`;"},
		{"DropDatabase", sqblmysql.DropDatabase("app"), "DROP DATABASE `app`;"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sqltesting.AssertSQL(t, tt.b.MustSQL(), tt.want)
		})
	}
}

// A realistic query shows every MySQL-specific rendering choice at once:
// backtick quoting per dotted part, escaped backticks, ? for indexed parameters
// (the index is dropped), and TRUE literals.
func TestDialectRendering(t *testing.T) {
	t.Parallel()
	got := sqblmysql.From(sqblmysql.As("users", "u")).
		Select("u.id", "u.nick`name", sqblmysql.As(sqblmysql.Fn("COUNT", "o.id"), "orders")).
		LeftJoin(sqblmysql.As("orders", "o"), sqblmysql.Eq("u.id", "o.user_id")).
		Where(sqblmysql.And(
			sqblmysql.Eq("u.active", true),
			sqblmysql.Like("u.name", sqblmysql.P(1)),
			sqblmysql.Gt("o.amount", sqblmysql.P(2)),
		)).
		GroupBy("u.id").
		MustSQL()
	sqltesting.AssertSQL(t, got, "SELECT `u`.`id`, `u`.`nick``name`, COUNT(`o`.`id`) AS `orders` "+
		"FROM `users` AS `u` "+
		"LEFT JOIN `orders` AS `o` ON `u`.`id` = `o`.`user_id` "+
		"WHERE `u`.`active` = TRUE AND `u`.`name` LIKE ? AND `o`.`amount` > ? "+
		"GROUP BY `u`.`id`;")
}
