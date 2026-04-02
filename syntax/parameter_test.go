package syntax

import (
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
			got := P().ToSqlWithDialect(tt.dialect)
			if got != tt.expected {
				t.Errorf("P().ToSqlWithDialect(%s) = %q; want %q", tt.name, got, tt.expected)
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
			got := P(tt.index).ToSqlWithDialect(tt.dialect)
			if got != tt.expected {
				t.Errorf("P(%d).ToSqlWithDialect(%s) = %q; want %q", tt.index, tt.name, got, tt.expected)
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
			got := P(tt.param).ToSqlWithDialect(d)
			if got != tt.expected {
				t.Errorf("P(%q).ToSqlWithDialect() = %q; want %q", tt.param, got, tt.expected)
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
		{
			name:     "positional in Eq",
			expr:     Eq("status", P()),
			expected: `"status" = ?`,
		},
		{
			name:     "indexed in Eq",
			expr:     Eq("id", P(1)),
			expected: `"id" = $1`,
		},
		{
			name:     "named in Eq",
			expr:     Eq("status", P(":status")),
			expected: `"status" = :status`,
		},
		{
			name:     "indexed not double-quoted",
			expr:     Eq("col", P(1)),
			expected: `"col" = $1`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// t.Parallel()
			got := tt.expr.ToSqlWithDialect(d)
			if got != tt.expected {
				t.Errorf("%s: ToSqlWithDialect() = %q; want %q", tt.name, got, tt.expected)
			}
		})
	}
}

func TestParameter_InExpr(t *testing.T) {
	t.Parallel()
	d := &dialect.PostgresDialect{}
	got := In("status", P(1), P(2)).ToSqlWithDialect(d)
	expected := `"status" IN ($1, $2)`
	if got != expected {
		t.Errorf("In with parameters: got %q; want %q", got, expected)
	}
}

func TestParameter_BetweenExpr(t *testing.T) {
	t.Parallel()
	d := &dialect.PostgresDialect{}
	got := Between("age", P(1), P(2)).ToSqlWithDialect(d)
	expected := `"age" BETWEEN $1 AND $2`
	if got != expected {
		t.Errorf("Between with parameters: got %q; want %q", got, expected)
	}
}

// P with an unrecognised argument type falls back to positional mode.
func TestParameter_P_DefaultFallback(t *testing.T) {
	t.Parallel()
	d := &dialect.PostgresDialect{}
	if got, want := P(3.14).ToSqlWithDialect(d), "?"; got != want {
		t.Errorf("P(float64).ToSqlWithDialect() = %q, want %q", got, want)
	}
}

// A Parameter constructed with an out-of-range mode hits the default branch of ToSqlWithDialect.
func TestParameter_ToSql_DefaultBranch(t *testing.T) {
	t.Parallel()
	d := &dialect.PostgresDialect{}
	p := Parameter{mode: paramMode(99)}
	if got, want := p.ToSqlWithDialect(d), "?"; got != want {
		t.Errorf("unknown mode ToSqlWithDialect() = %q, want %q", got, want)
	}
}
