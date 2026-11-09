package sqbl_test

import (
	"bytes"
	"testing"

	"github.com/kotofurumiya/sqbl"
	"github.com/kotofurumiya/sqbl/dialect"
	"github.com/kotofurumiya/sqbl/internal/sqltesting"
	"github.com/kotofurumiya/sqbl/syntax"
)

// Each re-export is a one-line wrapper, so one case per helper is enough to
// catch a wrapper that delegates to the wrong syntax function.
func TestReexports(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		f    syntax.SqlFragment
		want string
	}{
		{"As", sqbl.As("users", "u"), `"users" AS "u"`},
		{"AsColumns", sqbl.AsColumns(sqbl.Fn("unnest", sqbl.P(1)), "ids", "id"), `unnest($1) AS "ids" ("id")`},
		{"Eq", sqbl.Eq("a", 1), `"a" = 1`},
		{"Ne", sqbl.Ne("a", 1), `"a" <> 1`},
		{"Lt", sqbl.Lt("a", 1), `"a" < 1`},
		{"Lte", sqbl.Lte("a", 1), `"a" <= 1`},
		{"Gt", sqbl.Gt("a", 1), `"a" > 1`},
		{"Gte", sqbl.Gte("a", 1), `"a" >= 1`},
		{"And", sqbl.And(sqbl.Eq("a", 1), sqbl.Eq("b", 2)), `"a" = 1 AND "b" = 2`},
		{"Or", sqbl.Or(sqbl.Eq("a", 1), sqbl.Eq("b", 2)), `"a" = 1 OR "b" = 2`},
		{"Asc", sqbl.Asc("a"), `"a" ASC`},
		{"Desc", sqbl.Desc("a"), `"a" DESC`},
		{"Not", sqbl.Not(sqbl.Eq("a", 1)), `NOT ("a" = 1)`},
		{"In", sqbl.In("a", 1, 2), `"a" IN (1, 2)`},
		{"NotIn", sqbl.NotIn("a", 1, 2), `"a" NOT IN (1, 2)`},
		{"IsNull", sqbl.IsNull("a"), `"a" IS NULL`},
		{"IsNotNull", sqbl.IsNotNull("a"), `"a" IS NOT NULL`},
		{"Between", sqbl.Between("a", 1, 2), `"a" BETWEEN 1 AND 2`},
		{"Like", sqbl.Like("a", sqbl.P(1)), `"a" LIKE $1`},
		{"ILike", sqbl.ILike("a", sqbl.P(1)), `"a" ILIKE $1`},
		{"P", sqbl.P(1), `$1`},
		{"Fn", sqbl.Fn("SUM", "a"), `SUM("a")`},
		{"Cast", sqbl.Cast("a", "TEXT"), `CAST("a" AS TEXT)`},
		{"Over", sqbl.Over(sqbl.Fn("ROW_NUMBER")).OrderBy("a"), `ROW_NUMBER() OVER (ORDER BY "a")`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			tt.f.AppendSQL(&buf, &dialect.PostgresDialect{})
			sqltesting.AssertSQL(t, buf.String(), tt.want)
		})
	}
}
