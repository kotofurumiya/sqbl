package sqblsqlite

import (
	"testing"

	"github.com/kotofurumiya/sqbl/dialect"
	"github.com/kotofurumiya/sqbl/syntax"
)

// These tests are intentionally minimal.
// Every function in reexport.go is a one-line wrapper around the corresponding
// syntax package function. The underlying behaviour is thoroughly tested in
// syntax/operator_test.go and syntax/parameter_test.go.
// The sole purpose here is to confirm that each reexport delegates to the
// correct syntax function, and to provide package-level coverage.
func TestReexports(t *testing.T) {
	t.Parallel()

	d := &dialect.SqliteDialect{}

	t.Run("As", func(t *testing.T) {
		t.Parallel()
		if got, want := As("users", "u").ToSqlWithDialect(d), syntax.As("users", "u").ToSqlWithDialect(d); got != want {
			t.Errorf("As: got %q, want %q", got, want)
		}
	})

	t.Run("Eq", func(t *testing.T) {
		t.Parallel()
		if got, want := Eq("a", "b").ToSqlWithDialect(d), syntax.Eq("a", "b").ToSqlWithDialect(d); got != want {
			t.Errorf("Eq: got %q, want %q", got, want)
		}
	})

	t.Run("Ne", func(t *testing.T) {
		t.Parallel()
		if got, want := Ne("a", "b").ToSqlWithDialect(d), syntax.Ne("a", "b").ToSqlWithDialect(d); got != want {
			t.Errorf("Ne: got %q, want %q", got, want)
		}
	})

	t.Run("Lt", func(t *testing.T) {
		t.Parallel()
		if got, want := Lt("age", 18).ToSqlWithDialect(d), syntax.Lt("age", 18).ToSqlWithDialect(d); got != want {
			t.Errorf("Lt: got %q, want %q", got, want)
		}
	})

	t.Run("Lte", func(t *testing.T) {
		t.Parallel()
		if got, want := Lte("age", 18).ToSqlWithDialect(d), syntax.Lte("age", 18).ToSqlWithDialect(d); got != want {
			t.Errorf("Lte: got %q, want %q", got, want)
		}
	})

	t.Run("Gt", func(t *testing.T) {
		t.Parallel()
		if got, want := Gt("score", 100).ToSqlWithDialect(d), syntax.Gt("score", 100).ToSqlWithDialect(d); got != want {
			t.Errorf("Gt: got %q, want %q", got, want)
		}
	})

	t.Run("Gte", func(t *testing.T) {
		t.Parallel()
		if got, want := Gte("score", 100).ToSqlWithDialect(d), syntax.Gte("score", 100).ToSqlWithDialect(d); got != want {
			t.Errorf("Gte: got %q, want %q", got, want)
		}
	})

	t.Run("And", func(t *testing.T) {
		t.Parallel()
		e1, e2 := syntax.Eq("a", "b"), syntax.Eq("c", "d")
		if got, want := And(e1, e2).ToSqlWithDialect(d), syntax.And(e1, e2).ToSqlWithDialect(d); got != want {
			t.Errorf("And: got %q, want %q", got, want)
		}
	})

	t.Run("Or", func(t *testing.T) {
		t.Parallel()
		e1, e2 := syntax.Eq("a", "b"), syntax.Eq("c", "d")
		if got, want := Or(e1, e2).ToSqlWithDialect(d), syntax.Or(e1, e2).ToSqlWithDialect(d); got != want {
			t.Errorf("Or: got %q, want %q", got, want)
		}
	})

	t.Run("Asc", func(t *testing.T) {
		t.Parallel()
		if got, want := Asc("name").ToSqlWithDialect(d), syntax.Asc("name").ToSqlWithDialect(d); got != want {
			t.Errorf("Asc: got %q, want %q", got, want)
		}
	})

	t.Run("Desc", func(t *testing.T) {
		t.Parallel()
		if got, want := Desc("created_at").ToSqlWithDialect(d), syntax.Desc("created_at").ToSqlWithDialect(d); got != want {
			t.Errorf("Desc: got %q, want %q", got, want)
		}
	})

	t.Run("Not", func(t *testing.T) {
		t.Parallel()
		e := syntax.Eq("active", false)
		if got, want := Not(e).ToSqlWithDialect(d), syntax.Not(e).ToSqlWithDialect(d); got != want {
			t.Errorf("Not: got %q, want %q", got, want)
		}
	})

	t.Run("In", func(t *testing.T) {
		t.Parallel()
		if got, want := In("status", "'active'", "'pending'").ToSqlWithDialect(d), syntax.In("status", "'active'", "'pending'").ToSqlWithDialect(d); got != want {
			t.Errorf("In: got %q, want %q", got, want)
		}
	})

	t.Run("NotIn", func(t *testing.T) {
		t.Parallel()
		if got, want := NotIn("status", "'deleted'").ToSqlWithDialect(d), syntax.NotIn("status", "'deleted'").ToSqlWithDialect(d); got != want {
			t.Errorf("NotIn: got %q, want %q", got, want)
		}
	})

	t.Run("IsNull", func(t *testing.T) {
		t.Parallel()
		if got, want := IsNull("deleted_at").ToSqlWithDialect(d), syntax.IsNull("deleted_at").ToSqlWithDialect(d); got != want {
			t.Errorf("IsNull: got %q, want %q", got, want)
		}
	})

	t.Run("IsNotNull", func(t *testing.T) {
		t.Parallel()
		if got, want := IsNotNull("email").ToSqlWithDialect(d), syntax.IsNotNull("email").ToSqlWithDialect(d); got != want {
			t.Errorf("IsNotNull: got %q, want %q", got, want)
		}
	})

	t.Run("Between", func(t *testing.T) {
		t.Parallel()
		if got, want := Between("age", 18, 65).ToSqlWithDialect(d), syntax.Between("age", 18, 65).ToSqlWithDialect(d); got != want {
			t.Errorf("Between: got %q, want %q", got, want)
		}
	})

	t.Run("Like", func(t *testing.T) {
		t.Parallel()
		if got, want := Like("name", "'%foo%'").ToSqlWithDialect(d), syntax.Like("name", "'%foo%'").ToSqlWithDialect(d); got != want {
			t.Errorf("Like: got %q, want %q", got, want)
		}
	})

	t.Run("P", func(t *testing.T) {
		t.Parallel()
		if got, want := P().ToSqlWithDialect(d), syntax.P().ToSqlWithDialect(d); got != want {
			t.Errorf("P(): got %q, want %q", got, want)
		}
		if got, want := P(1).ToSqlWithDialect(d), syntax.P(1).ToSqlWithDialect(d); got != want {
			t.Errorf("P(1): got %q, want %q", got, want)
		}
		if got, want := P(":status").ToSqlWithDialect(d), syntax.P(":status").ToSqlWithDialect(d); got != want {
			t.Errorf("P(named): got %q, want %q", got, want)
		}
	})
}
