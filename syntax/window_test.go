package syntax

import (
	"bytes"
	"testing"

	"github.com/kotofurumiya/sqbl/dialect"
)

func TestWindowExpr(t *testing.T) {
	t.Parallel()
	d := &dialect.SimpleDialect{}

	tests := []struct {
		name string
		w    WindowExpr
		want string
	}{
		{
			name: "empty window",
			w:    Over(Fn("ROW_NUMBER")),
			want: "ROW_NUMBER() OVER ()",
		},
		{
			name: "PARTITION BY",
			w:    Over(Fn("ROW_NUMBER")).PartitionBy("dept"),
			want: `ROW_NUMBER() OVER (PARTITION BY "dept")`,
		},
		{
			name: "ORDER BY string",
			w:    Over(Fn("RANK")).OrderBy("salary"),
			want: `RANK() OVER (ORDER BY "salary")`,
		},
		{
			name: "ORDER BY Desc",
			w:    Over(Fn("RANK")).OrderBy(Desc("salary")),
			want: `RANK() OVER (ORDER BY "salary" DESC)`,
		},
		{
			name: "PARTITION BY and ORDER BY",
			w:    Over(Fn("ROW_NUMBER")).PartitionBy("dept").OrderBy("salary"),
			want: `ROW_NUMBER() OVER (PARTITION BY "dept" ORDER BY "salary")`,
		},
		{
			name: "multiple PARTITION BY cols",
			w:    Over(Fn("RANK")).PartitionBy("dept", "region").OrderBy("salary"),
			want: `RANK() OVER (PARTITION BY "dept", "region" ORDER BY "salary")`,
		},
		{
			name: "ROWS frame",
			w:    Over(Fn("SUM", "amount")).OrderBy("id").Rows("BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW"),
			want: `SUM("amount") OVER (ORDER BY "id" ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW)`,
		},
		{
			name: "RANGE frame",
			w:    Over(Fn("SUM", "amount")).OrderBy("id").Range("BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW"),
			want: `SUM("amount") OVER (ORDER BY "id" RANGE BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW)`,
		},
		{
			name: "full window spec",
			w:    Over(Fn("SUM", "amount")).PartitionBy("dept").OrderBy("salary").Rows("BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW"),
			want: `SUM("amount") OVER (PARTITION BY "dept" ORDER BY "salary" ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW)`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			tt.w.AppendSQL(&buf, d)
			if got := buf.String(); got != tt.want {
				t.Errorf("WindowExpr.AppendSQL() = %q; want %q", got, tt.want)
			}
		})
	}
}

func TestWindowExpr_WithPostgresDialect(t *testing.T) {
	t.Parallel()
	d := &dialect.PostgresDialect{}

	tests := []struct {
		name string
		w    WindowExpr
		want string
	}{
		{
			name: "qualified column in PARTITION BY",
			w:    Over(Fn("ROW_NUMBER")).PartitionBy("u.dept").OrderBy("u.salary"),
			want: `ROW_NUMBER() OVER (PARTITION BY "u"."dept" ORDER BY "u"."salary")`,
		},
		{
			name: "SUM with qualified col and PARTITION BY",
			w:    Over(Fn("SUM", "p.amount")).PartitionBy("p.user_id").OrderBy("p.created_at"),
			want: `SUM("p"."amount") OVER (PARTITION BY "p"."user_id" ORDER BY "p"."created_at")`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			tt.w.AppendSQL(&buf, d)
			if got := buf.String(); got != tt.want {
				t.Errorf("WindowExpr.AppendSQL(pg) = %q; want %q", got, tt.want)
			}
		})
	}
}

func TestWindowExpr_AsFragment(t *testing.T) {
	t.Parallel()
	d := &dialect.SimpleDialect{}

	// WindowExpr implements SqlFragment and can be used with As()
	var buf bytes.Buffer
	As(Over(Fn("ROW_NUMBER")).PartitionBy("dept").OrderBy("salary"), "rn").AppendSQL(&buf, d)
	want := `ROW_NUMBER() OVER (PARTITION BY "dept" ORDER BY "salary") AS "rn"`
	if got := buf.String(); got != want {
		t.Errorf("As(Over(...), alias).AppendSQL() = %q; want %q", got, want)
	}
}
