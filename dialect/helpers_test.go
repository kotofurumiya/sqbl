package dialect

import (
	"bytes"
	"testing"
)

func TestWriteQuotedPart(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		quoteChar byte
		expected  string
	}{
		// Double-quote variant (Postgres / SQLite)
		{name: "simple identifier", input: "users", quoteChar: '"', expected: "\"users\""},
		{name: "empty string", input: "", quoteChar: '"', expected: "\"\""},
		{name: "quote in middle", input: "us\"ers", quoteChar: '"', expected: "\"us\"\"ers\""},
		{name: "leading quote", input: "\"users", quoteChar: '"', expected: "\"\"\"users\""},
		{name: "trailing quote", input: "users\"", quoteChar: '"', expected: "\"users\"\"\""},
		{name: "multiple quotes", input: "\"a\"b\"", quoteChar: '"', expected: "\"\"\"a\"\"b\"\"\""},

		// Backtick variant (MySQL)
		{name: "simple backtick", input: "users", quoteChar: '`', expected: "`users`"},
		{name: "backtick injection", input: "users` --", quoteChar: '`', expected: "`users`` --`"},
		{name: "multiple backticks", input: "`weird`name`", quoteChar: '`', expected: "```weird``name```"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			WriteQuotedPart(&buf, tt.input, tt.quoteChar)
			if got := buf.String(); got != tt.expected {
				t.Errorf("WriteQuotedPart(%q, %q) = %q; want %q", tt.input, tt.quoteChar, got, tt.expected)
			}
		})
	}
}

func TestQuoteIdentifierParts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		quoteChar byte
		expected  string
	}{
		// Double-quote variant (Postgres / SQLite)
		{name: "single part", input: "users", quoteChar: '"', expected: "\"users\""},
		{name: "two parts", input: "public.users", quoteChar: '"', expected: "\"public\".\"users\""},
		{name: "three parts", input: "db.sch.table", quoteChar: '"', expected: "\"db\".\"sch\".\"table\""},
		{name: "quote in part", input: "use\"rs", quoteChar: '"', expected: "\"use\"\"rs\""},
		{name: "quote in dotted part", input: "pub\"lic.users", quoteChar: '"', expected: "\"pub\"\"lic\".\"users\""},
		{name: "empty parts (consecutive dots)", input: "a..b", quoteChar: '"', expected: "\"a\".\"\".\"b\""},
		{name: "leading dot", input: ".table", quoteChar: '"', expected: "\"\".\"table\""},
		{name: "trailing dot", input: "schema.", quoteChar: '"', expected: "\"schema\".\"\""},

		// Backtick variant (MySQL)
		{name: "single part backtick", input: "users", quoteChar: '`', expected: "`users`"},
		{name: "two parts backtick", input: "mydb.users", quoteChar: '`', expected: "`mydb`.`users`"},
		{name: "backtick in part", input: "use`rs", quoteChar: '`', expected: "`use``rs`"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			QuoteIdentifierParts(&buf, tt.input, tt.quoteChar)
			if got := buf.String(); got != tt.expected {
				t.Errorf("QuoteIdentifierParts(%q, %q) = %q; want %q", tt.input, tt.quoteChar, got, tt.expected)
			}
		})
	}
}
