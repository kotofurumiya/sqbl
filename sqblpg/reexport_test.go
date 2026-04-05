package sqblpg

import (
	"bytes"
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

	d := &dialect.PostgresDialect{}

	t.Run("As", func(t *testing.T) {
		t.Parallel()
		var got, want bytes.Buffer
		As("users", "u").AppendSQL(&got, d)
		syntax.As("users", "u").AppendSQL(&want, d)
		if got.String() != want.String() {
			t.Errorf("As: got %q, want %q", got.String(), want.String())
		}
	})

	t.Run("Eq", func(t *testing.T) {
		t.Parallel()
		var got, want bytes.Buffer
		Eq("a", "b").AppendSQL(&got, d)
		syntax.Eq("a", "b").AppendSQL(&want, d)
		if got.String() != want.String() {
			t.Errorf("Eq: got %q, want %q", got.String(), want.String())
		}
	})

	t.Run("Ne", func(t *testing.T) {
		t.Parallel()
		var got, want bytes.Buffer
		Ne("a", "b").AppendSQL(&got, d)
		syntax.Ne("a", "b").AppendSQL(&want, d)
		if got.String() != want.String() {
			t.Errorf("Ne: got %q, want %q", got.String(), want.String())
		}
	})

	t.Run("Lt", func(t *testing.T) {
		t.Parallel()
		var got, want bytes.Buffer
		Lt("age", 18).AppendSQL(&got, d)
		syntax.Lt("age", 18).AppendSQL(&want, d)
		if got.String() != want.String() {
			t.Errorf("Lt: got %q, want %q", got.String(), want.String())
		}
	})

	t.Run("Lte", func(t *testing.T) {
		t.Parallel()
		var got, want bytes.Buffer
		Lte("age", 18).AppendSQL(&got, d)
		syntax.Lte("age", 18).AppendSQL(&want, d)
		if got.String() != want.String() {
			t.Errorf("Lte: got %q, want %q", got.String(), want.String())
		}
	})

	t.Run("Gt", func(t *testing.T) {
		t.Parallel()
		var got, want bytes.Buffer
		Gt("score", 100).AppendSQL(&got, d)
		syntax.Gt("score", 100).AppendSQL(&want, d)
		if got.String() != want.String() {
			t.Errorf("Gt: got %q, want %q", got.String(), want.String())
		}
	})

	t.Run("Gte", func(t *testing.T) {
		t.Parallel()
		var got, want bytes.Buffer
		Gte("score", 100).AppendSQL(&got, d)
		syntax.Gte("score", 100).AppendSQL(&want, d)
		if got.String() != want.String() {
			t.Errorf("Gte: got %q, want %q", got.String(), want.String())
		}
	})

	t.Run("And", func(t *testing.T) {
		t.Parallel()
		e1, e2 := syntax.Eq("a", "b"), syntax.Eq("c", "d")
		var got, want bytes.Buffer
		And(e1, e2).AppendSQL(&got, d)
		syntax.And(e1, e2).AppendSQL(&want, d)
		if got.String() != want.String() {
			t.Errorf("And: got %q, want %q", got.String(), want.String())
		}
	})

	t.Run("Or", func(t *testing.T) {
		t.Parallel()
		e1, e2 := syntax.Eq("a", "b"), syntax.Eq("c", "d")
		var got, want bytes.Buffer
		Or(e1, e2).AppendSQL(&got, d)
		syntax.Or(e1, e2).AppendSQL(&want, d)
		if got.String() != want.String() {
			t.Errorf("Or: got %q, want %q", got.String(), want.String())
		}
	})

	t.Run("Asc", func(t *testing.T) {
		t.Parallel()
		var got, want bytes.Buffer
		Asc("name").AppendSQL(&got, d)
		syntax.Asc("name").AppendSQL(&want, d)
		if got.String() != want.String() {
			t.Errorf("Asc: got %q, want %q", got.String(), want.String())
		}
	})

	t.Run("Desc", func(t *testing.T) {
		t.Parallel()
		var got, want bytes.Buffer
		Desc("created_at").AppendSQL(&got, d)
		syntax.Desc("created_at").AppendSQL(&want, d)
		if got.String() != want.String() {
			t.Errorf("Desc: got %q, want %q", got.String(), want.String())
		}
	})

	t.Run("Not", func(t *testing.T) {
		t.Parallel()
		e := syntax.Eq("active", false)
		var got, want bytes.Buffer
		Not(e).AppendSQL(&got, d)
		syntax.Not(e).AppendSQL(&want, d)
		if got.String() != want.String() {
			t.Errorf("Not: got %q, want %q", got.String(), want.String())
		}
	})

	t.Run("In", func(t *testing.T) {
		t.Parallel()
		var got, want bytes.Buffer
		In("status", "'active'", "'pending'").AppendSQL(&got, d)
		syntax.In("status", "'active'", "'pending'").AppendSQL(&want, d)
		if got.String() != want.String() {
			t.Errorf("In: got %q, want %q", got.String(), want.String())
		}
	})

	t.Run("NotIn", func(t *testing.T) {
		t.Parallel()
		var got, want bytes.Buffer
		NotIn("status", "'deleted'").AppendSQL(&got, d)
		syntax.NotIn("status", "'deleted'").AppendSQL(&want, d)
		if got.String() != want.String() {
			t.Errorf("NotIn: got %q, want %q", got.String(), want.String())
		}
	})

	t.Run("IsNull", func(t *testing.T) {
		t.Parallel()
		var got, want bytes.Buffer
		IsNull("deleted_at").AppendSQL(&got, d)
		syntax.IsNull("deleted_at").AppendSQL(&want, d)
		if got.String() != want.String() {
			t.Errorf("IsNull: got %q, want %q", got.String(), want.String())
		}
	})

	t.Run("IsNotNull", func(t *testing.T) {
		t.Parallel()
		var got, want bytes.Buffer
		IsNotNull("email").AppendSQL(&got, d)
		syntax.IsNotNull("email").AppendSQL(&want, d)
		if got.String() != want.String() {
			t.Errorf("IsNotNull: got %q, want %q", got.String(), want.String())
		}
	})

	t.Run("Between", func(t *testing.T) {
		t.Parallel()
		var got, want bytes.Buffer
		Between("age", 18, 65).AppendSQL(&got, d)
		syntax.Between("age", 18, 65).AppendSQL(&want, d)
		if got.String() != want.String() {
			t.Errorf("Between: got %q, want %q", got.String(), want.String())
		}
	})

	t.Run("Like", func(t *testing.T) {
		t.Parallel()
		var got, want bytes.Buffer
		Like("name", "'%foo%'").AppendSQL(&got, d)
		syntax.Like("name", "'%foo%'").AppendSQL(&want, d)
		if got.String() != want.String() {
			t.Errorf("Like: got %q, want %q", got.String(), want.String())
		}
	})

	t.Run("ILike", func(t *testing.T) {
		t.Parallel()
		var got, want bytes.Buffer
		ILike("name", "'%foo%'").AppendSQL(&got, d)
		syntax.ILike("name", "'%foo%'").AppendSQL(&want, d)
		if got.String() != want.String() {
			t.Errorf("ILike: got %q, want %q", got.String(), want.String())
		}
	})

	t.Run("P", func(t *testing.T) {
		t.Parallel()
		var got, want bytes.Buffer

		P().AppendSQL(&got, d)
		syntax.P().AppendSQL(&want, d)
		if got.String() != want.String() {
			t.Errorf("P(): got %q, want %q", got.String(), want.String())
		}

		got.Reset()
		want.Reset()
		P(1).AppendSQL(&got, d)
		syntax.P(1).AppendSQL(&want, d)
		if got.String() != want.String() {
			t.Errorf("P(1): got %q, want %q", got.String(), want.String())
		}

		got.Reset()
		want.Reset()
		P(":status").AppendSQL(&got, d)
		syntax.P(":status").AppendSQL(&want, d)
		if got.String() != want.String() {
			t.Errorf("P(named): got %q, want %q", got.String(), want.String())
		}
	})
}
