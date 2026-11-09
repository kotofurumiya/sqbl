package syntax

import (
	"testing"

	"github.com/kotofurumiya/sqbl/internal/sqltesting"
)

func TestFnAndCast(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		fragment SqlFragment
		want     string
	}{
		{"Fn without args", Fn("NOW"), "NOW()"},
		{"Fn identifier arg", Fn("SUM", "p.amount"), `SUM("p"."amount")`},
		{"Fn wildcard arg", Fn("COUNT", "*"), "COUNT(*)"},
		{"Fn mixed args", Fn("COALESCE", "x", P(1), 0, nil, true), `COALESCE("x", $1, 0, NULL, TRUE)`},
		{"Fn raw arg", Fn("DATE_TRUNC", Raw("'day'"), "ts"), `DATE_TRUNC('day', "ts")`},
		{"Fn nested", Fn("SUM", Fn("ABS", "val")), `SUM(ABS("val"))`},
		{"Cast parameter", Cast(P(1), "bigint[]"), "CAST($1 AS bigint[])"},
		{"Cast identifier", Cast("t.price", "TEXT"), `CAST("t"."price" AS TEXT)`},
		{"Cast function", Cast(Fn("NOW"), "DATE"), "CAST(NOW() AS DATE)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sqltesting.AssertSQL(t, render(tt.fragment), tt.want)
		})
	}
}
