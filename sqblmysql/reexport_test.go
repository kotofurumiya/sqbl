package sqblmysql_test

import (
	"bytes"
	"testing"

	"github.com/kotofurumiya/sqbl/dialect"
	"github.com/kotofurumiya/sqbl/internal/sqltesting"
	"github.com/kotofurumiya/sqbl/sqblmysql"
	"github.com/kotofurumiya/sqbl/syntax"
)

// Each re-export is a one-line wrapper, so one case per helper is enough to
// catch a wrapper that delegates to the wrong syntax function.
func TestReexports(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		f    syntax.SQLFragment
		want string
	}{
		{"As", sqblmysql.As("users", "u"), "`users` AS `u`"},
		{"AsColumns", sqblmysql.AsColumns(sqblmysql.Fn("unnest", sqblmysql.P(1)), "ids", "id"), "unnest(?) AS `ids` (`id`)"},
		{"Eq", sqblmysql.Eq("a", 1), "`a` = 1"},
		{"Ne", sqblmysql.Ne("a", 1), "`a` <> 1"},
		{"Lt", sqblmysql.Lt("a", 1), "`a` < 1"},
		{"Lte", sqblmysql.Lte("a", 1), "`a` <= 1"},
		{"Gt", sqblmysql.Gt("a", 1), "`a` > 1"},
		{"Gte", sqblmysql.Gte("a", 1), "`a` >= 1"},
		{"And", sqblmysql.And(sqblmysql.Eq("a", 1), sqblmysql.Eq("b", 2)), "`a` = 1 AND `b` = 2"},
		{"Or", sqblmysql.Or(sqblmysql.Eq("a", 1), sqblmysql.Eq("b", 2)), "`a` = 1 OR `b` = 2"},
		{"Asc", sqblmysql.Asc("a"), "`a` ASC"},
		{"Desc", sqblmysql.Desc("a"), "`a` DESC"},
		{"Desc expression", sqblmysql.Desc(sqblmysql.Fn("LOWER", "a")), "LOWER(`a`) DESC"},
		{"Not", sqblmysql.Not(sqblmysql.Eq("a", 1)), "NOT (`a` = 1)"},
		{"In", sqblmysql.In("a", 1, 2), "`a` IN (1, 2)"},
		{"NotIn", sqblmysql.NotIn("a", 1, 2), "`a` NOT IN (1, 2)"},
		{"IsNull", sqblmysql.IsNull("a"), "`a` IS NULL"},
		{"IsNotNull", sqblmysql.IsNotNull("a"), "`a` IS NOT NULL"},
		{"Between", sqblmysql.Between("a", 1, 2), "`a` BETWEEN 1 AND 2"},
		{"Like", sqblmysql.Like("a", sqblmysql.P(1)), "`a` LIKE ?"},
		{"P", sqblmysql.P(1), "?"},
		{"Fn", sqblmysql.Fn("SUM", "a"), "SUM(`a`)"},
		{"Cast", sqblmysql.Cast("a", "CHAR"), "CAST(`a` AS CHAR)"},
		{"Over", sqblmysql.Over(sqblmysql.Fn("ROW_NUMBER")).OrderBy("a"), "ROW_NUMBER() OVER (ORDER BY `a`)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			tt.f.AppendSQL(&buf, &dialect.MySQL{})
			sqltesting.AssertSQL(t, buf.String(), tt.want)
		})
	}
}
