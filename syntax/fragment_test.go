package syntax

import (
	"bytes"
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

func TestStringExpr(t *testing.T) {
	t.Parallel()
	d := &dialect.SimpleDialect{}

	// Function calls are expressions and must pass through unquoted.
	var buf bytes.Buffer
	NewStringExpr("COUNT(*)").AppendSQL(&buf, d)
	if got := buf.String(); got != "COUNT(*)" {
		t.Errorf("StringExpr(COUNT(*)) = %q; want %q", got, "COUNT(*)")
	}

	// Plain names — including those with spaces — are identifiers and must be quoted.
	buf.Reset()
	NewStringExpr("my column").AppendSQL(&buf, d)
	if got := buf.String(); got != `"my column"` {
		t.Errorf("StringExpr(my column) = %q; want %q", got, `"my column"`)
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
			input:    NewStringExpr("raw"),
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
			var buf bytes.Buffer
			ToFragment(tt.input).AppendSQL(&buf, d)
			if got := buf.String(); got != tt.expected {
				t.Errorf("ToFragment(%v).AppendSQL() = %q; want %q", tt.input, got, tt.expected)
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
			var buf bytes.Buffer
			As(tt.source, tt.alias).AppendSQL(&buf, d)
			if got := buf.String(); got != tt.expected {
				t.Errorf("As(%q, %q).AppendSQL() = %q; want %q", tt.source, tt.alias, got, tt.expected)
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
			var buf bytes.Buffer
			tt.order.AppendSQL(&buf, d)
			if got := buf.String(); got != tt.expected {
				t.Errorf("Order.AppendSQL() = %q; want %q", got, tt.expected)
			}
		})
	}
}
