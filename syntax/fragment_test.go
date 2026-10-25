package syntax

import (
	"bytes"
	"testing"

	"github.com/kotofurumiya/sqbl/dialect"
)

func TestIdent(t *testing.T) {
	t.Parallel()
	d := &dialect.PostgresDialect{}

	tests := []struct {
		name  string
		input Ident
		want  string
	}{
		{name: "column", input: "name", want: `"name"`},
		{name: "qualified column", input: "u.id", want: `"u"."id"`},
		{name: "identifier with space", input: "my column", want: `"my column"`},
		// Text that looks like an expression is still a single identifier.
		{name: "function-like text", input: "COUNT(*)", want: `"COUNT(*)"`},
		{name: "literal-like text", input: "'value'", want: `"'value'"`},
		{name: "embedded quote is escaped", input: `evil"col; DROP TABLE t; --`, want: `"evil""col; DROP TABLE t; --"`},
		{name: "wildcard", input: "*", want: "*"},
		{name: "qualified wildcard", input: "u.*", want: `"u".*`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			tt.input.AppendSQL(&buf, d)
			if got := buf.String(); got != tt.want {
				t.Errorf("Ident(%q) = %q; want %q", string(tt.input), got, tt.want)
			}
		})
	}
}

func TestRaw(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	Raw(`"count" + 1`).AppendSQL(&buf, &dialect.PostgresDialect{})
	if got, want := buf.String(), `"count" + 1`; got != want {
		t.Errorf("Raw = %q; want %q", got, want)
	}
}

func TestAppendValue(t *testing.T) {
	t.Parallel()
	d := &dialect.SqliteDialect{}

	tests := []struct {
		name  string
		input any
		want  string
	}{
		{name: "string is identifier", input: "users.id", want: `"users"."id"`},
		{name: "Ident", input: Ident("name"), want: `"name"`},
		{name: "Raw", input: Raw("NOW()"), want: "NOW()"},
		{name: "Parameter", input: P(), want: "?"},
		{name: "Fn", input: Fn("LOWER", "email"), want: `LOWER("email")`},
		{name: "int", input: 42, want: "42"},
		{name: "float", input: 1.5, want: "1.5"},
		{name: "bool uses dialect", input: true, want: "1"},
		{name: "nil is NULL", input: nil, want: "NULL"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			AppendValue(&buf, d, tt.input)
			if got := buf.String(); got != tt.want {
				t.Errorf("AppendValue(%v) = %q; want %q", tt.input, got, tt.want)
			}

			// ToFragment must render the same SQL as AppendValue.
			buf.Reset()
			ToFragment(tt.input).AppendSQL(&buf, d)
			if got := buf.String(); got != tt.want {
				t.Errorf("ToFragment(%v) = %q; want %q", tt.input, got, tt.want)
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
			source:   Fn("SUM", "amount"),
			alias:    "total",
			expected: `SUM("amount") AS "total"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// t.Parallel()
			var buf bytes.Buffer
			As(tt.source, tt.alias).AppendSQL(&buf, d)
			if got := buf.String(); got != tt.expected {
				t.Errorf("As(%v, %q).AppendSQL() = %q; want %q", tt.source, tt.alias, got, tt.expected)
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
