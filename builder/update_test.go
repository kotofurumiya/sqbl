package builder_test

import (
	"testing"

	"github.com/kotofurumiya/sqbl/internal/sqltesting"
	"github.com/kotofurumiya/sqbl/syntax"
)

func TestUpdateBuilder_Clauses(t *testing.T) {
	t.Parallel()
	users := newUpdate().Table("users")

	runSQLCases(t, []sqlCase{
		{
			name: "assignments accumulate in call order",
			b:    users.Set("name", syntax.P(1)).Set("email", syntax.P(2)),
			want: `UPDATE "users" SET "name" = $1, "email" = $2;`,
		},
		{
			// A plain string is an identifier, so it copies another column.
			name: "value kinds",
			b: users.
				Set("a", 42).Set("b", true).Set("c", nil).Set("d", "other_col").
				Set("e", syntax.Raw(`"e" + 1`)).Set("f", syntax.Fn("NOW")),
			want: `UPDATE "users" SET "a" = 42, "b" = TRUE, "c" = NULL, "d" = "other_col", "e" = "e" + 1, "f" = NOW();`,
		},
		{
			name: "where and returning",
			b:    users.Set("name", syntax.P(1)).Where(syntax.Eq("id", syntax.P(2))).Returning("id", "name"),
			want: `UPDATE "users" SET "name" = $1 WHERE "id" = $2 RETURNING "id", "name";`,
		},
		{
			name: "where and returning are replaced by later calls",
			b: users.Set("name", syntax.P(1)).
				Where(syntax.Eq("old", 1)).Where(syntax.Eq("id", syntax.P(2))).
				Returning("old").Returning("id"),
			want: `UPDATE "users" SET "name" = $1 WHERE "id" = $2 RETURNING "id";`,
		},
	})
}

// TestUpdateBuilder_BranchIsolation checks that deriving two statements from
// a shared base does not leak assignments between them. The base holds three
// assignments so that a backing array with spare capacity would be shared if
// Set appended in place.
func TestUpdateBuilder_BranchIsolation(t *testing.T) {
	t.Parallel()
	base := newUpdate().Table("t").Set("a", 1).Set("b", 2).Set("c", 3)
	left := base.Set("l", 4)
	right := base.Set("r", 5)

	sqltesting.AssertSQL(t, base.MustSQL(), `UPDATE "t" SET "a" = 1, "b" = 2, "c" = 3;`)
	sqltesting.AssertSQL(t, left.MustSQL(), `UPDATE "t" SET "a" = 1, "b" = 2, "c" = 3, "l" = 4;`)
	sqltesting.AssertSQL(t, right.MustSQL(), `UPDATE "t" SET "a" = 1, "b" = 2, "c" = 3, "r" = 5;`)
}
