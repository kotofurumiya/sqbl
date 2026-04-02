package syntax

import (
	"testing"

	"github.com/kotofurumiya/sqbl/dialect"
)

func TestIsSimpleIdentifier(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{name: "column", input: "name", expected: true},
		{name: "qualified column", input: "u.id", expected: true},
		{name: "multi-part", input: "p.purchased_at", expected: true},
		{name: "function call", input: "SUM(p.amount)", expected: false},
		{name: "no-arg function", input: "NOW()", expected: false},
		{name: "string literal", input: "'value'", expected: false},
		{name: "empty string", input: "", expected: true},
		// Spaces are valid in quoted identifiers ("my column"), so a space
		// alone must not disqualify a string from being treated as an identifier.
		{name: "identifier with space", input: "my column", expected: true},
		{name: "injection-like identifier", input: `evil"col; DROP TABLE t; --`, expected: true},
		{name: "wildcard", input: "*", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// t.Parallel()
			got := isSimpleIdentifier(tt.input)
			if got != tt.expected {
				t.Errorf("isSimpleIdentifier(%q) = %v; want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestStringSource(t *testing.T) {
	t.Parallel()
	d := &dialect.SimpleDialect{}

	// Function calls are expressions and must pass through unquoted.
	got := NewStringSource("COUNT(*)").ToSqlWithDialect(d)
	if got != "COUNT(*)" {
		t.Errorf("StringSource(COUNT(*)) = %q; want %q", got, "COUNT(*)")
	}

	// Plain names — including those with spaces — are identifiers and must be quoted.
	got = NewStringSource("my column").ToSqlWithDialect(d)
	if got != `"my column"` {
		t.Errorf("StringSource(my column) = %q; want %q", got, `"my column"`)
	}
}

func TestToFragment(t *testing.T) {
	t.Parallel()
	d := &dialect.SimpleDialect{}

	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{
			name:     "string",
			input:    "hello",
			expected: `"hello"`,
		},
		{
			name:     "SqlFragment",
			input:    NewStringSource("raw"),
			expected: `"raw"`,
		},
		{
			name:     "int",
			input:    42,
			expected: `"42"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// t.Parallel()
			got := ToFragment(tt.input).ToSqlWithDialect(d)
			if got != tt.expected {
				t.Errorf("ToFragment(%v).ToSqlWithDialect() = %q; want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestAliased(t *testing.T) {
	t.Parallel()
	d := &dialect.SimpleDialect{}

	tests := []struct {
		name     string
		source   any
		alias    string
		expected string
	}{
		{
			name:     "table alias",
			source:   "users",
			alias:    "u",
			expected: `"users" AS "u"`,
		},
		{
			name:     "expression alias",
			source:   "SUM(p.amount)",
			alias:    "total",
			expected: `SUM(p.amount) AS "total"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// t.Parallel()
			got := As(tt.source, tt.alias).ToSqlWithDialect(d)
			if got != tt.expected {
				t.Errorf("As(%q, %q).ToSqlWithDialect() = %q; want %q", tt.source, tt.alias, got, tt.expected)
			}
		})
	}
}

func TestOrder(t *testing.T) {
	t.Parallel()
	d := &dialect.SimpleDialect{}

	tests := []struct {
		name     string
		order    Order
		expected string
	}{
		{
			name:     "Asc",
			order:    Asc("name"),
			expected: `"name" ASC`,
		},
		{
			name:     "Desc",
			order:    Desc("created_at"),
			expected: `"created_at" DESC`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// t.Parallel()
			got := tt.order.ToSqlWithDialect(d)
			if got != tt.expected {
				t.Errorf("Order.ToSqlWithDialect() = %q; want %q", got, tt.expected)
			}
		})
	}
}
