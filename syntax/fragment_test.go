package syntax

import (
	"bytes"
	"errors"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/kotofurumiya/sqbl/dialect"
	"github.com/kotofurumiya/sqbl/internal/sqlerr"
	"github.com/kotofurumiya/sqbl/internal/sqltesting"
)

type (
	userID int64
	flag   bool
	status string
)

// renderE renders f with PostgreSQL, the dialect used where dialect differences
// are not under test, and returns a rendering failure as an error the way builders do.
func renderE(f SQLFragment) (sql string, err error) {
	defer sqlerr.Catch(&err)
	var buf bytes.Buffer
	f.AppendSQL(&buf, &dialect.Postgres{})
	return buf.String(), nil
}

// render is renderE for fragments that must render; a failure shows up in the SQL mismatch.
func render(f SQLFragment) string {
	sql, err := renderE(f)
	if err != nil {
		return "error: " + err.Error()
	}
	return sql
}

func TestDialectDependentFragments(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                    string
		fragment                SQLFragment
		postgres, mysql, sqlite string
	}{
		{"Ident", Ident("name"), `"name"`, "`name`", `"name"`},
		{"Ident qualified", Ident("u.id"), `"u"."id"`, "`u`.`id`", `"u"."id"`},
		{"Ident wildcard", Ident("*"), "*", "*", "*"},
		{"Ident qualified wildcard", Ident("u.*"), `"u".*`, "`u`.*", `"u".*`},
		{"bool true", ToFragment(true), "TRUE", "TRUE", "1"},
		{"bool false", ToFragment(false), "FALSE", "FALSE", "0"},
		{"P positional", P(), "?", "?", "?"},
		// MySQL has no numbered placeholders, so the index is dropped there.
		{"P indexed", P(2), "$2", "?", "?2"},
		{"P indexed with another integer type", P(uint8(3)), "$3", "?", "?3"},
		{"P named is written as-is", P(":status"), ":status", ":status", ":status"},
		{"P named with @", P("@user_id"), "@user_id", "@user_id", "@user_id"},
		{"P named with $", P("$name"), "$name", "$name", "$name"},
		// Empty conditions render their identity value, which keeps WHERE valid.
		{"empty And", And(), "TRUE", "TRUE", "1"},
		{"empty Or", Or(), "FALSE", "FALSE", "0"},
		{"empty In", In("id"), "FALSE", "FALSE", "0"},
		{"empty NotIn", NotIn("id"), "TRUE", "TRUE", "1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			for _, c := range []struct {
				name string
				d    dialect.Dialect
				want string
			}{
				{"postgres", &dialect.Postgres{}, tt.postgres},
				{"mysql", &dialect.MySQL{}, tt.mysql},
				{"sqlite", &dialect.SQLite{}, tt.sqlite},
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
		{"uint64 max", uint64(18446744073709551615), "18446744073709551615"},
		{"named integer type", userID(9), "9"},
		{"float", 1.5, "1.5"},
		// float32 is formatted at its own precision, not widened to float64 digits.
		{"float32", float32(0.1), "0.1"},
		{"large float uses an exponent", 1e21, "1e+21"},
		{"named bool type", flag(true), "TRUE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sqltesting.AssertSQL(t, render(ToFragment(tt.input)), tt.want)

			var buf bytes.Buffer
			AppendValue(&buf, &dialect.Postgres{}, tt.input)
			sqltesting.AssertSQL(t, buf.String(), tt.want)
		})
	}
}

func TestAliasAndOrder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		fragment SQLFragment
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
		{"Asc expression", Asc(Fn("LOWER", "name")), `LOWER("name") ASC`},
		{"Desc position", Desc(1), "1 DESC"},
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
	exprs := []SQLFragment{Raw("x"), Raw("y")}

	tests := []struct {
		name     string
		fragment SQLFragment
		want     string
	}{
		{"AsColumns", AsColumns("t", "s", names...), `"t" AS "s" ("a", "b")`},
		{"Fn", Fn("f", args...), `f("a", "b")`},
		{"And", And(exprs...), `x AND y`},
		{"Or", Or(exprs...), `x OR y`},
		{"In", In("c", args...), `"c" IN ("a", "b")`},
		{"NotIn", NotIn("c", args...), `"c" NOT IN ("a", "b")`},
		{"window", Over(Fn("f")).PartitionBy(args...).OrderBy(args...), `f() OVER (PARTITION BY "a", "b" ORDER BY "a", "b")`},
	}
	names[0], args[0], exprs[0] = "z", "z", Raw("z")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sqltesting.AssertSQL(t, render(tt.fragment), tt.want)
		})
	}
}

// TestRenderErrors checks that input with no safe SQL form fails the render
// with a message that names the problem, instead of producing broken SQL.
func TestRenderErrors(t *testing.T) {
	t.Parallel()
	n := 1

	tests := []struct {
		name     string
		fragment SQLFragment
		wantErr  string
	}{
		// Values whose kind has no SQL literal must be bound with P.
		{"named string type", Eq("status", status("x' OR '1'='1")), "convert it to string for an identifier"},
		{"time.Time", Eq("at", time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)), "cannot render time.Time"},
		{"[]byte", Eq("blob", []byte("ab")), "cannot render []uint8"},
		{"pointer", Eq("n", &n), "cannot render *int"},
		{"error value", Eq("msg", errors.New("boom")), "cannot render *errors.errorString"},
		{"slice in In is not spread", In("id", []int{1, 2}), "cannot render []int"},
		{"NaN", Gt("score", math.NaN()), "NaN and infinity"},
		{"infinity", Lt("score", math.Inf(1)), "NaN and infinity"},
		{"ToFragment defers the failure to rendering", ToFragment(status("x")), "cannot render syntax.status"},

		{"P index zero", P(0), "between 1 and"},
		{"P negative index", P(-1), "between 1 and"},
		{"P name without prefix", P("status"), "invalid parameter name"},
		{"P name with SQL", P(":a; DROP TABLE t"), "invalid parameter name"},
		{"P float", P(1.5), "integer index or a placeholder name"},
		{"P too many arguments", P(1, 2), "at most one argument"},

		{"Eq with nil", Eq("deleted_at", nil), "use IsNull or IsNotNull"},
		{"Ne with nil on the left", Ne(nil, "a"), "use IsNull or IsNotNull"},
		{"Like with nil", Like("name", nil), "use IsNull or IsNotNull"},
		{"nil inside And", And(Eq("a", 1), nil), "AND operand 2 is nil"},
		{"nil inside Or", Or(nil), "OR operand 1 is nil"},
		{"Not nil", Not(nil), "Not has a nil condition"},
		{"As without source", As(nil, "x"), `alias "x" has no source`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sql, err := renderE(tt.fragment)
			if err == nil {
				t.Fatalf("want an error containing %q, rendered %s", tt.wantErr, sql)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error %q does not contain %q", err, tt.wantErr)
			}
		})
	}
}
