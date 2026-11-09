package syntax

import (
	"bytes"
	"testing"

	"github.com/kotofurumiya/sqbl/dialect"
	"github.com/kotofurumiya/sqbl/internal/sqltesting"
)

// render renders f with PostgreSQL, the dialect used where dialect differences are not under test.
func render(f SqlFragment) string {
	var buf bytes.Buffer
	f.AppendSQL(&buf, &dialect.PostgresDialect{})
	return buf.String()
}

func TestDialectDependentFragments(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                            string
		fragment                        SqlFragment
		postgres, mysql, sqlite, simple string
	}{
		{"Ident", Ident("name"), `"name"`, "`name`", `"name"`, `"name"`},
		{"Ident qualified", Ident("u.id"), `"u"."id"`, "`u`.`id`", `"u"."id"`, `"u.id"`},
		{"Ident wildcard", Ident("*"), "*", "*", "*", "*"},
		{"Ident qualified wildcard", Ident("u.*"), `"u".*`, "`u`.*", `"u".*`, `"u".*`},
		{"bool true", ToFragment(true), "TRUE", "TRUE", "1", "TRUE"},
		{"bool false", ToFragment(false), "FALSE", "FALSE", "0", "FALSE"},
		{"P positional", P(), "?", "?", "?", "?"},
		{"P indexed", P(2), "$2", "?", "?2", "?2"},
		{"P named is written as-is", P(":status"), ":status", ":status", ":status", ":status"},
		{"P with unsupported argument falls back to positional", P(3.14), "?", "?", "?", "?"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			for _, c := range []struct {
				name string
				d    dialect.SqlDialect
				want string
			}{
				{"postgres", &dialect.PostgresDialect{}, tt.postgres},
				{"mysql", &dialect.MysqlDialect{}, tt.mysql},
				{"sqlite", &dialect.SqliteDialect{}, tt.sqlite},
				{"simple", &dialect.SimpleDialect{}, tt.simple},
			} {
				var buf bytes.Buffer
				tt.fragment.AppendSQL(&buf, c.d)
				if got := buf.String(); got != c.want {
					t.Errorf("%s: got %s; want %s", c.name, got, c.want)
				}
			}
		})
	}
}

// TestToFragment checks how each Go type is interpreted where an expression is expected.
// AppendValue must agree with ToFragment, so both are checked against the same expectation.
func TestToFragment(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input any
		want  string
	}{
		{"string is an identifier", "users.id", `"users"."id"`},
		// Text that looks like SQL is still one identifier, which keeps strings injection-safe.
		{"string never becomes SQL", `n + 1; DROP TABLE "t"`, `"n + 1; DROP TABLE ""t"""`},
		{"string with a quote is escaped", `a"b`, `"a""b"`},
		{"Ident", Ident("name"), `"name"`},
		{"Raw is written as-is", Raw(`"count" + 1`), `"count" + 1`},
		{"Parameter", P(1), "$1"},
		{"other fragment", Fn("LOWER", "email"), `LOWER("email")`},
		{"bool", true, "TRUE"},
		{"nil is NULL", nil, "NULL"},
		{"int", 42, "42"},
		{"negative int64", int64(-7), "-7"},
		{"float", 1.5, "1.5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sqltesting.AssertSQL(t, render(ToFragment(tt.input)), tt.want)

			var buf bytes.Buffer
			AppendValue(&buf, &dialect.PostgresDialect{}, tt.input)
			sqltesting.AssertSQL(t, buf.String(), tt.want)
		})
	}
}

func TestAliasAndOrder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		fragment SqlFragment
		want     string
	}{
		{"As table", As("users", "u"), `"users" AS "u"`},
		{"As qualified table", As("public.users", "u"), `"public"."users" AS "u"`},
		{"As expression", As(Fn("SUM", "amount"), "total"), `SUM("amount") AS "total"`},
		// The alias is a single name, so a dot is not a qualifier there.
		{"As alias is not split on dots", As("users", "u.x"), `"users" AS "u.x"`},
		{"As alias is escaped", As("users", `u"`), `"users" AS "u"""`},
		{"AsColumns one column", AsColumns(Fn("unnest", P(1)), "ids", "id"), `unnest($1) AS "ids" ("id")`},
		{"AsColumns many columns", AsColumns(Fn("unnest", P(1), P(2)), "pairs", "id", "name"), `unnest($1, $2) AS "pairs" ("id", "name")`},
		{"AsColumns without columns", AsColumns("users", "u"), `"users" AS "u"`},
		{"Asc", Asc("name"), `"name" ASC`},
		{"Desc qualified", Desc("u.created_at"), `"u"."created_at" DESC`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sqltesting.AssertSQL(t, render(tt.fragment), tt.want)
		})
	}
}

// TestCopiesCallerSlices checks that fragments keep their own copy of slices
// passed to variadic constructors, so reusing a slice cannot rewrite them.
func TestCopiesCallerSlices(t *testing.T) {
	t.Parallel()
	names := []string{"a", "b"}
	args := []any{"a", "b"}
	exprs := []SqlFragment{Raw("x"), Raw("y")}

	tests := []struct {
		name     string
		fragment SqlFragment
		want     string
	}{
		{"AsColumns", AsColumns("t", "s", names...), `"t" AS "s" ("a", "b")`},
		{"Fn", Fn("f", args...), `f("a", "b")`},
		{"And", And(exprs...), `x AND y`},
		{"Or", Or(exprs...), `x OR y`},
		{"In", In("c", args...), `"c" IN ("a", "b")`},
		{"NotIn", NotIn("c", args...), `"c" NOT IN ("a", "b")`},
		{"window", Over(Fn("f")).PartitionBy(names...).OrderBy(args...), `f() OVER (PARTITION BY "a", "b" ORDER BY "a", "b")`},
	}
	names[0], args[0], exprs[0] = "z", "z", Raw("z")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sqltesting.AssertSQL(t, render(tt.fragment), tt.want)
		})
	}
}
