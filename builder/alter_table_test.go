package builder_test

import "testing"

func TestSqlAlterTableBuilder_Operations(t *testing.T) {
	t.Parallel()
	users := newAlterTable().Table("users")

	runSQLCases(t, []sqlCase{
		{"add column", users.AddColumn("bio", "TEXT NOT NULL"), `ALTER TABLE "users" ADD COLUMN "bio" TEXT NOT NULL;`},
		{"add column without type", users.AddColumn("bio", ""), `ALTER TABLE "users" ADD COLUMN "bio";`},
		{"drop column", users.DropColumn("bio"), `ALTER TABLE "users" DROP COLUMN "bio";`},
		{"rename column", users.RenameColumn("fullname", "name"), `ALTER TABLE "users" RENAME COLUMN "fullname" TO "name";`},
		{"rename table", users.RenameTable("members"), `ALTER TABLE "users" RENAME TO "members";`},
		{"qualified table", newAlterTable().Table("auth.users").DropColumn("bio"), `ALTER TABLE "auth"."users" DROP COLUMN "bio";`},
		// A statement holds a single operation, so a later call replaces the earlier one.
		{"later operation wins", users.AddColumn("bio", "TEXT").DropColumn("age"), `ALTER TABLE "users" DROP COLUMN "age";`},
		{"no operation", users, `ALTER TABLE "users";`},
	})
}
