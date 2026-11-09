package dialect_test

import (
	"bytes"
	"testing"

	"github.com/kotofurumiya/sqbl/dialect"
)

// perDialect holds the expected output of one rendering for each dialect.
type perDialect struct {
	postgres, mysql, sqlite string
}

func TestDialects(t *testing.T) {
	t.Parallel()

	quote := func(s string) func(*bytes.Buffer, dialect.Dialect) {
		return func(buf *bytes.Buffer, d dialect.Dialect) { d.Quote(buf, s) }
	}
	quoteIdent := func(s string) func(*bytes.Buffer, dialect.Dialect) {
		return func(buf *bytes.Buffer, d dialect.Dialect) { d.QuoteIdentifier(buf, s) }
	}
	indexed := func(i int) func(*bytes.Buffer, dialect.Dialect) {
		return func(buf *bytes.Buffer, d dialect.Dialect) { d.PlaceholderIndexed(buf, i) }
	}
	boolean := func(b bool) func(*bytes.Buffer, dialect.Dialect) {
		return func(buf *bytes.Buffer, d dialect.Dialect) { d.Bool(buf, b) }
	}

	tests := []struct {
		name   string
		render func(*bytes.Buffer, dialect.Dialect)
		want   perDialect
	}{
		{
			name:   "Quote plain",
			render: quote("users"),
			want:   perDialect{`"users"`, "`users`", `"users"`},
		},
		{
			// Quote handles a single part, so a dot stays inside the quotes.
			name:   "Quote keeps dots",
			render: quote("schema.table"),
			want:   perDialect{`"schema.table"`, "`schema.table`", `"schema.table"`},
		},
		{
			name:   "Quote escapes the dialect quote char",
			render: quote("a\"b`c"),
			want:   perDialect{`"a""b` + "`" + `c"`, "`a\"b``c`", `"a""b` + "`" + `c"`},
		},
		{
			name:   "QuoteIdentifier single part",
			render: quoteIdent("users"),
			want:   perDialect{`"users"`, "`users`", `"users"`},
		},
		{
			name:   "QuoteIdentifier splits on dots",
			render: quoteIdent("db.sch.table"),
			want:   perDialect{`"db"."sch"."table"`, "`db`.`sch`.`table`", `"db"."sch"."table"`},
		},
		{
			name:   "QuoteIdentifier keeps empty parts",
			render: quoteIdent(".a..b."),
			want:   perDialect{`""."a".""."b".""`, "``.`a`.``.`b`.``", `""."a".""."b".""`},
		},
		{
			// Escaping must apply to every part, including at part boundaries.
			name:   "QuoteIdentifier escapes in every part",
			render: quoteIdent("\"s\".`t`"),
			want: perDialect{
				`"""s"""."` + "`t`" + `"`,
				"`\"s\"`.```t```",
				`"""s"""."` + "`t`" + `"`,
			},
		},
		{
			// Injection payloads stay inside one quoted identifier.
			name:   "QuoteIdentifier contains injection",
			render: quoteIdent(`x"; DROP TABLE t; --`),
			want:   perDialect{`"x""; DROP TABLE t; --"`, "`x\"; DROP TABLE t; --`", `"x""; DROP TABLE t; --"`},
		},
		{
			name:   "PlaceholderPositional",
			render: func(buf *bytes.Buffer, d dialect.Dialect) { d.PlaceholderPositional(buf) },
			want:   perDialect{"?", "?", "?"},
		},
		{
			name:   "PlaceholderIndexed 1",
			render: indexed(1),
			want:   perDialect{"$1", "?", "?1"},
		},
		{
			name:   "PlaceholderIndexed multi-digit",
			render: indexed(123),
			want:   perDialect{"$123", "?", "?123"},
		},
		{
			name:   "Bool true",
			render: boolean(true),
			want:   perDialect{"TRUE", "TRUE", "1"},
		},
		{
			name:   "Bool false",
			render: boolean(false),
			want:   perDialect{"FALSE", "FALSE", "0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			for _, c := range []struct {
				name string
				d    dialect.Dialect
				want string
			}{
				{"postgres", &dialect.Postgres{}, tt.want.postgres},
				{"mysql", &dialect.MySQL{}, tt.want.mysql},
				{"sqlite", &dialect.SQLite{}, tt.want.sqlite},
			} {
				var buf bytes.Buffer
				tt.render(&buf, c.d)
				if got := buf.String(); got != c.want {
					t.Errorf("%s: got %s; want %s", c.name, got, c.want)
				}
			}
		})
	}
}
