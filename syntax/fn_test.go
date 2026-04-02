package syntax

import (
	"testing"

	"github.com/kotofurumiya/sqbl/dialect"
)

func TestFn(t *testing.T) {
	t.Parallel()
	d := &dialect.SimpleDialect{}

	tests := []struct {
		name string
		fn   *SqlFn
		want string
	}{
		{
			name: "no args",
			fn:   Fn("NOW"),
			want: "NOW()",
		},
		{
			name: "single identifier arg",
			fn:   Fn("SUM", "amount"),
			want: `SUM("amount")`,
		},
		{
			// SimpleDialect.QuoteIdentifier does not split on dots.
			// Use PostgresDialect for dialect-aware quoting in TestFn_WithPostgresDialect.
			name: "qualified identifier arg",
			fn:   Fn("SUM", "p.amount"),
			want: `SUM("p.amount")`,
		},
		{
			name: "wildcard arg",
			fn:   Fn("COUNT", "*"),
			want: "COUNT(*)",
		},
		{
			name: "numeric literal arg",
			fn:   Fn("COALESCE", "x", 0),
			want: `COALESCE("x", 0)`,
		},
		{
			name: "multiple args",
			fn:   Fn("COALESCE", "s.total", 0),
			want: `COALESCE("s.total", 0)`,
		},
		{
			name: "nested Fn",
			fn:   Fn("SUM", Fn("ABS", "val")),
			want: `SUM(ABS("val"))`,
		},
		{
			name: "parameter arg",
			fn:   Fn("COALESCE", "x", P(1)),
			want: `COALESCE("x", ?1)`, // SimpleDialect uses ?N for indexed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fn.ToSqlWithDialect(d)
			if got != tt.want {
				t.Errorf("Fn.ToSqlWithDialect() = %q; want %q", got, tt.want)
			}
		})
	}
}

func TestFn_WithPostgresDialect(t *testing.T) {
	t.Parallel()
	d := &dialect.PostgresDialect{}

	tests := []struct {
		name string
		fn   *SqlFn
		want string
	}{
		{
			name: "SUM with identifier",
			fn:   Fn("SUM", "amount"),
			want: `SUM("amount")`,
		},
		{
			name: "COUNT star",
			fn:   Fn("COUNT", "*"),
			want: "COUNT(*)",
		},
		{
			name: "COALESCE with literal",
			fn:   Fn("COALESCE", "s.total", 0),
			want: `COALESCE("s"."total", 0)`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fn.ToSqlWithDialect(d)
			if got != tt.want {
				t.Errorf("Fn.ToSqlWithDialect(pg) = %q; want %q", got, tt.want)
			}
		})
	}
}

func TestFn_AsFragment(t *testing.T) {
	t.Parallel()
	d := &dialect.SimpleDialect{}

	// Fn implements SqlFragment and can be used with As()
	got := As(Fn("SUM", "amount"), "total").ToSqlWithDialect(d)
	want := `SUM("amount") AS "total"`
	if got != want {
		t.Errorf("As(Fn(...), alias) = %q; want %q", got, want)
	}
}
