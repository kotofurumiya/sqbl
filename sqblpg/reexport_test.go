package sqblpg_test

import (
	"bytes"
	"testing"

	"github.com/kotofurumiya/sqbl/dialect"
	"github.com/kotofurumiya/sqbl/internal/sqltesting"
	"github.com/kotofurumiya/sqbl/sqblpg"
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
		{"As", sqblpg.As("users", "u"), `"users" AS "u"`},
		{"AsColumns", sqblpg.AsColumns(sqblpg.Fn("unnest", sqblpg.P(1)), "ids", "id"), `unnest($1) AS "ids" ("id")`},
		{"Eq", sqblpg.Eq("a", 1), `"a" = 1`},
		{"Ne", sqblpg.Ne("a", 1), `"a" <> 1`},
		{"Lt", sqblpg.Lt("a", 1), `"a" < 1`},
		{"Lte", sqblpg.Lte("a", 1), `"a" <= 1`},
		{"Gt", sqblpg.Gt("a", 1), `"a" > 1`},
		{"Gte", sqblpg.Gte("a", 1), `"a" >= 1`},
		{"And", sqblpg.And(sqblpg.Eq("a", 1), sqblpg.Eq("b", 2)), `"a" = 1 AND "b" = 2`},
		{"Or", sqblpg.Or(sqblpg.Eq("a", 1), sqblpg.Eq("b", 2)), `"a" = 1 OR "b" = 2`},
		{"Asc", sqblpg.Asc("a"), `"a" ASC`},
		{"Desc", sqblpg.Desc("a"), `"a" DESC`},
		{"Not", sqblpg.Not(sqblpg.Eq("a", 1)), `NOT ("a" = 1)`},
		{"In", sqblpg.In("a", 1, 2), `"a" IN (1, 2)`},
		{"NotIn", sqblpg.NotIn("a", 1, 2), `"a" NOT IN (1, 2)`},
		{"IsNull", sqblpg.IsNull("a"), `"a" IS NULL`},
		{"IsNotNull", sqblpg.IsNotNull("a"), `"a" IS NOT NULL`},
		{"Between", sqblpg.Between("a", 1, 2), `"a" BETWEEN 1 AND 2`},
		{"Like", sqblpg.Like("a", sqblpg.P(1)), `"a" LIKE $1`},
		{"ILike", sqblpg.ILike("a", sqblpg.P(1)), `"a" ILIKE $1`},
		{"P", sqblpg.P(1), `$1`},
		{"Fn", sqblpg.Fn("SUM", "a"), `SUM("a")`},
		{"Cast", sqblpg.Cast("a", "TEXT"), `CAST("a" AS TEXT)`},
		{"Over", sqblpg.Over(sqblpg.Fn("ROW_NUMBER")).OrderBy("a"), `ROW_NUMBER() OVER (ORDER BY "a")`},
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
