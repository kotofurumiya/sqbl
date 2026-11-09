package sqblsqlite_test

import (
	"bytes"
	"testing"

	"github.com/kotofurumiya/sqbl/dialect"
	"github.com/kotofurumiya/sqbl/internal/sqltesting"
	"github.com/kotofurumiya/sqbl/sqblsqlite"
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
		{"As", sqblsqlite.As("users", "u"), `"users" AS "u"`},
		{"Eq", sqblsqlite.Eq("a", 1), `"a" = 1`},
		{"Ne", sqblsqlite.Ne("a", 1), `"a" <> 1`},
		{"Lt", sqblsqlite.Lt("a", 1), `"a" < 1`},
		{"Lte", sqblsqlite.Lte("a", 1), `"a" <= 1`},
		{"Gt", sqblsqlite.Gt("a", 1), `"a" > 1`},
		{"Gte", sqblsqlite.Gte("a", 1), `"a" >= 1`},
		{"And", sqblsqlite.And(sqblsqlite.Eq("a", 1), sqblsqlite.Eq("b", 2)), `"a" = 1 AND "b" = 2`},
		{"Or", sqblsqlite.Or(sqblsqlite.Eq("a", 1), sqblsqlite.Eq("b", 2)), `"a" = 1 OR "b" = 2`},
		{"Asc", sqblsqlite.Asc("a"), `"a" ASC`},
		{"Desc", sqblsqlite.Desc("a"), `"a" DESC`},
		{"Not", sqblsqlite.Not(sqblsqlite.Eq("a", 1)), `NOT ("a" = 1)`},
		{"In", sqblsqlite.In("a", 1, 2), `"a" IN (1, 2)`},
		{"NotIn", sqblsqlite.NotIn("a", 1, 2), `"a" NOT IN (1, 2)`},
		{"IsNull", sqblsqlite.IsNull("a"), `"a" IS NULL`},
		{"IsNotNull", sqblsqlite.IsNotNull("a"), `"a" IS NOT NULL`},
		{"Between", sqblsqlite.Between("a", 1, 2), `"a" BETWEEN 1 AND 2`},
		{"Like", sqblsqlite.Like("a", sqblsqlite.P(1)), `"a" LIKE ?1`},
		{"P", sqblsqlite.P(1), `?1`},
		{"Fn", sqblsqlite.Fn("SUM", "a"), `SUM("a")`},
		{"Cast", sqblsqlite.Cast("a", "TEXT"), `CAST("a" AS TEXT)`},
		{"Over", sqblsqlite.Over(sqblsqlite.Fn("ROW_NUMBER")).OrderBy("a"), `ROW_NUMBER() OVER (ORDER BY "a")`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			tt.f.AppendSQL(&buf, &dialect.SqliteDialect{})
			sqltesting.AssertSQL(t, buf.String(), tt.want)
		})
	}
}
