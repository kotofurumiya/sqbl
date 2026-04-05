package syntax

import (
	"bytes"
	"testing"

	"github.com/kotofurumiya/sqbl/dialect"
)

func TestParameter_Positional(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		dialect  dialect.SqlDialect
		expected string
	}{
		{"postgres", &dialect.PostgresDialect{}, "?"},
		{"mysql", &dialect.MysqlDialect{}, "?"},
		{"sqlite", &dialect.SqliteDialect{}, "?"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// t.Parallel()
			var buf bytes.Buffer
			P().AppendSQL(&buf, tt.dialect)
			if got := buf.String(); got != tt.expected {
				t.Errorf("P().AppendSQL(%s) = %q; want %q", tt.name, got, tt.expected)
			}
		})
	}
}

func TestParameter_Indexed(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		dialect  dialect.SqlDialect
		index    int
		expected string
	}{
		{"postgres $1", &dialect.PostgresDialect{}, 1, "$1"},
		{"postgres $3", &dialect.PostgresDialect{}, 3, "$3"},
		{"mysql", &dialect.MysqlDialect{}, 1, "?"},
		{"sqlite", &dialect.SqliteDialect{}, 1, "?"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// t.Parallel()
			var buf bytes.Buffer
			P(tt.index).AppendSQL(&buf, tt.dialect)
			if got := buf.String(); got != tt.expected {
				t.Errorf("P(%d).AppendSQL(%s) = %q; want %q", tt.index, tt.name, got, tt.expected)
			}
		})
	}
}

func TestParameter_Named(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		param    string
		expected string
	}{
		{"colon prefix", ":status", ":status"},
		{"at prefix", "@name", "@name"},
		{"dollar prefix", "$param", "$param"},
	}

	d := &dialect.PostgresDialect{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// t.Parallel()
			var buf bytes.Buffer
			P(tt.param).AppendSQL(&buf, d)
			if got := buf.String(); got != tt.expected {
				t.Errorf("P(%q).AppendSQL() = %q; want %q", tt.param, got, tt.expected)
			}
		})
	}
}

func TestParameter_NotQuoted_InComparison(t *testing.T) {
	t.Parallel()
	d := &dialect.PostgresDialect{}

	tests := []struct {
		name     string
		expr     SqlFragment
		expected string
	}{
		{name: "positional in Eq", expr: Eq("status", P()), expected: `"status" = ?`},
		{name: "indexed in Eq", expr: Eq("id", P(1)), expected: `"id" = $1`},
		{name: "named in Eq", expr: Eq("status", P(":status")), expected: `"status" = :status`},
		{name: "indexed not double-quoted", expr: Eq("col", P(1)), expected: `"col" = $1`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// t.Parallel()
			var buf bytes.Buffer
			tt.expr.AppendSQL(&buf, d)
			if got := buf.String(); got != tt.expected {
				t.Errorf("%s: AppendSQL() = %q; want %q", tt.name, got, tt.expected)
			}
		})
	}
}

func TestParameter_InExpr(t *testing.T) {
	t.Parallel()
	d := &dialect.PostgresDialect{}
	var buf bytes.Buffer
	In("status", P(1), P(2)).AppendSQL(&buf, d)
	if got := buf.String(); got != `"status" IN ($1, $2)` {
		t.Errorf("In with parameters: got %q; want %q", got, `"status" IN ($1, $2)`)
	}
}

func TestParameter_BetweenExpr(t *testing.T) {
	t.Parallel()
	d := &dialect.PostgresDialect{}
	var buf bytes.Buffer
	Between("age", P(1), P(2)).AppendSQL(&buf, d)
	if got := buf.String(); got != `"age" BETWEEN $1 AND $2` {
		t.Errorf("Between with parameters: got %q; want %q", got, `"age" BETWEEN $1 AND $2`)
	}
}

// P with an unrecognised argument type falls back to positional mode.
func TestParameter_P_DefaultFallback(t *testing.T) {
	t.Parallel()
	d := &dialect.PostgresDialect{}
	var buf bytes.Buffer
	P(3.14).AppendSQL(&buf, d)
	if got := buf.String(); got != "?" {
		t.Errorf("P(float64).AppendSQL() = %q, want %q", got, "?")
	}
}

// A Parameter constructed with an out-of-range mode hits the default branch of AppendSQL.
func TestParameter_ToSql_DefaultBranch(t *testing.T) {
	t.Parallel()
	d := &dialect.PostgresDialect{}
	p := Parameter{mode: paramMode(99)}
	var buf bytes.Buffer
	p.AppendSQL(&buf, d)
	if got := buf.String(); got != "?" {
		t.Errorf("unknown mode AppendSQL() = %q, want %q", got, "?")
	}
}
