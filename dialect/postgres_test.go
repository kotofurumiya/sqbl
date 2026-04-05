package dialect

import (
	"bytes"
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
			var buf bytes.Buffer
			d.QuoteIdentifier(&buf, tt.input)
			if got := buf.String(); got != tt.expected {
				t.Errorf("QuoteIdentifier(%q) = %s; want %s", tt.input, got, tt.expected)
			}
		})
	}
}

func TestPostgresDialect_Quote(t *testing.T) {
	t.Parallel()
	d := &PostgresDialect{}

	// Quote should not split by dots, it treats everything as a single part.
	var buf bytes.Buffer
	d.Quote(&buf, "schema.table")
	if got, want := buf.String(), "\"schema.table\""; got != want {
		t.Errorf("Quote(%q) = %s; want %s", "schema.table", got, want)
	}
}

func TestPostgresDialect_PlaceholderPositional(t *testing.T) {
	t.Parallel()
	d := &PostgresDialect{}
	var buf bytes.Buffer
	d.PlaceholderPositional(&buf)
	if got := buf.String(); got != "?" {
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
			var buf bytes.Buffer
			d.PlaceholderIndexed(&buf, tt.index)
			if got := buf.String(); got != tt.expected {
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
			var buf bytes.Buffer
			d.Bool(&buf, tt.input)
			if got := buf.String(); got != tt.expected {
				t.Errorf("Bool(%v) = %s; want %s", tt.input, got, tt.expected)
			}
		})
	}
}
