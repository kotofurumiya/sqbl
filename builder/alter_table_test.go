package builder_test

import (
	"testing"

	"github.com/kotofurumiya/sqbl/internal/sqltesting"
)

func TestAlterTableBuilder_Operations(t *testing.T) {
	t.Parallel()
	users := newAlterTable().Table("users")

	runSQLCases(t, []sqlCase{
		{"add column", users.AddColumn("bio", "TEXT NOT NULL"), `ALTER TABLE "users" ADD COLUMN "bio" TEXT NOT NULL;`},
		{"add column without type", users.AddColumn("bio", ""), `ALTER TABLE "users" ADD COLUMN "bio";`},
		{"drop column", users.DropColumn("bio"), `ALTER TABLE "users" DROP COLUMN "bio";`},
		{"rename column", users.RenameColumn("fullname", "name"), `ALTER TABLE "users" RENAME COLUMN "fullname" TO "name";`},
		{"rename table", users.RenameTable("members"), `ALTER TABLE "users" RENAME TO "members";`},
		{"qualified table", newAlterTable().Table("auth.users").DropColumn("bio"), `ALTER TABLE "auth"."users" DROP COLUMN "bio";`},
		{
			name: "operations accumulate in call order",
			b:    users.AddColumn("bio", "TEXT").DropColumn("age").RenameColumn("fullname", "name"),
			want: `ALTER TABLE "users" ADD COLUMN "bio" TEXT, DROP COLUMN "age", RENAME COLUMN "fullname" TO "name";`,
		},
		{"no operation", users, `ALTER TABLE "users";`},
	})
}

// TestAlterTableBuilder_BranchIsolation checks that deriving two statements
// from a shared base does not leak operations between them. The base holds
// three operations so that a backing array with spare capacity would be shared
// if an operation were appended in place.
func TestAlterTableBuilder_BranchIsolation(t *testing.T) {
	t.Parallel()
	base := newAlterTable().Table("t").DropColumn("a").DropColumn("b").DropColumn("c")
	left := base.DropColumn("l")
	right := base.DropColumn("r")

	sqltesting.AssertSQL(t, base.MustSQL(), `ALTER TABLE "t" DROP COLUMN "a", DROP COLUMN "b", DROP COLUMN "c";`)
	sqltesting.AssertSQL(t, left.MustSQL(), `ALTER TABLE "t" DROP COLUMN "a", DROP COLUMN "b", DROP COLUMN "c", DROP COLUMN "l";`)
	sqltesting.AssertSQL(t, right.MustSQL(), `ALTER TABLE "t" DROP COLUMN "a", DROP COLUMN "b", DROP COLUMN "c", DROP COLUMN "r";`)
}
