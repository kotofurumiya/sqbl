package syntax

import (
	"testing"

	"github.com/kotofurumiya/sqbl/internal/sqltesting"
)

func TestWindowExpr(t *testing.T) {
	t.Parallel()

	const frame = "BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW"
	tests := []struct {
		name string
		w    SQLFragment
		want string
	}{
		{"empty window", Over(Fn("ROW_NUMBER")), "ROW_NUMBER() OVER ()"},
		{"PARTITION BY only", Over(Fn("ROW_NUMBER")).PartitionBy("u.dept", "region"), `ROW_NUMBER() OVER (PARTITION BY "u"."dept", "region")`},
		{"ORDER BY only", Over(Fn("RANK")).OrderBy("salary", Desc("u.hired_at")), `RANK() OVER (ORDER BY "salary", "u"."hired_at" DESC)`},
		{"frame only", Over(Fn("SUM", "amount")).Rows(frame), `SUM("amount") OVER (ROWS ` + frame + `)`},
		{"PARTITION BY and frame", Over(Fn("SUM", "amount")).PartitionBy("dept").Range(frame), `SUM("amount") OVER (PARTITION BY "dept" RANGE ` + frame + `)`},
		{
			"full spec",
			Over(Fn("SUM", "amount")).PartitionBy("dept").OrderBy("id").Rows(frame),
			`SUM("amount") OVER (PARTITION BY "dept" ORDER BY "id" ROWS ` + frame + `)`,
		},
		// Setters replace rather than accumulate, so the last call wins.
		{"later setters replace earlier ones", Over(Fn("RANK")).OrderBy("a").OrderBy("b").Rows(frame).Range(frame), `RANK() OVER (ORDER BY "b" RANGE ` + frame + `)`},
		{"aliased", As(Over(Fn("ROW_NUMBER")).OrderBy("id"), "rn"), `ROW_NUMBER() OVER (ORDER BY "id") AS "rn"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sqltesting.AssertSQL(t, render(tt.w), tt.want)
		})
	}
}
