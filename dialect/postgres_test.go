package dialect

import (
	"testing"
)

func TestPostgresDialect_QuoteIdentifier(t *testing.T) {
	t.Parallel()
	d := &PostgresDialect{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Standard cases
		{
			name:     "single table name",
			input:    "users",
			expected: "\"users\"",
		},
		{
			name:     "schema and table",
			input:    "public.users",
			expected: "\"public\".\"users\"",
		},
		{
			name:     "database, schema, and table",
			input:    "db.sch.table",
			expected: "\"db\".\"sch\".\"table\"",
		},

		// Malicious / Edge cases (SQL Injection attempts)
		{
			name:     "semicolon injection",
			input:    "users; DROP TABLE accounts",
			expected: "\"users; DROP TABLE accounts\"",
		},
		{
			name:     "quote injection",
			input:    "users\" --",
			expected: "\"users\"\" --\"",
		},
		{
			name:     "multiple quotes",
			input:    `"weird"name"`,
			expected: `"""weird""name"""`,
		},
		{
			name:     "comment injection",
			input:    "users --",
			expected: "\"users --\"",
		},

		// Structural edge cases
		{
			name:     "empty parts (consecutive dots)",
			input:    "a..b",
			expected: "\"a\".\"\".\"b\"",
		},
		{
			name:     "leading dot",
			input:    ".table",
			expected: "\"\".\"table\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// t.Parallel()
			got := d.QuoteIdentifier(tt.input)
			if got != tt.expected {
				t.Errorf("QuoteIdentifier(%q) = %s; want %s", tt.input, got, tt.expected)
			}
		})
	}
}

func TestPostgresDialect_Quote(t *testing.T) {
	t.Parallel()
	d := &PostgresDialect{}

	// Quote should not split by dots, it treats everything as a single part.
	input := "schema.table"
	expected := "\"schema.table\""
	got := d.Quote(input)

	if got != expected {
		t.Errorf("Quote(%q) = %s; want %s", input, got, expected)
	}
}

func TestPostgresDialect_PlaceholderPositional(t *testing.T) {
	t.Parallel()
	d := &PostgresDialect{}
	got := d.PlaceholderPositional()
	if got != "?" {
		t.Errorf("PlaceholderPositional() = %s; want ?", got)
	}
}

func TestPostgresDialect_PlaceholderIndexed(t *testing.T) {
	t.Parallel()
	d := &PostgresDialect{}

	tests := []struct {
		name     string
		index    int
		expected string
	}{
		{"first placeholder", 1, "$1"},
		{"second placeholder", 2, "$2"},
		{"large index", 999, "$999"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// t.Parallel()
			got := d.PlaceholderIndexed(tt.index)
			if got != tt.expected {
				t.Errorf("PlaceholderIndexed(%d) = %s; want %s", tt.index, got, tt.expected)
			}
		})
	}
}

func TestPostgresDialect_Bool(t *testing.T) {
	t.Parallel()
	d := &PostgresDialect{}

	tests := []struct {
		name     string
		input    bool
		expected string
	}{
		{"true value", true, "TRUE"},
		{"false value", false, "FALSE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// t.Parallel()
			got := d.Bool(tt.input)
			if got != tt.expected {
				t.Errorf("Bool(%v) = %s; want %s", tt.input, got, tt.expected)
			}
		})
	}
}
