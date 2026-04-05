package builder_test

import (
	"testing"

	"github.com/kotofurumiya/sqbl/builder"
	"github.com/kotofurumiya/sqbl/dialect"
	"github.com/kotofurumiya/sqbl/syntax"
)

func newDeleteBuilder() builder.SqlDeleteBuilder {
	return (builder.SqlDeleteBuilder{}).Dialect(&dialect.SimpleDialect{})
}

func TestSqlDeleteBuilder_WithWhere(t *testing.T) {
	t.Parallel()
	got := newDeleteBuilder().
		From("users").
		Where(syntax.Eq("id", syntax.P(1))).
		ToSql()
	want := `DELETE FROM "users" WHERE "id" = ?1;`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlDeleteBuilder_NoWhere(t *testing.T) {
	t.Parallel()
	got := newDeleteBuilder().
		From("sessions").
		ToSql()
	want := `DELETE FROM "sessions";`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlDeleteBuilder_WhereAnd(t *testing.T) {
	t.Parallel()
	got := newDeleteBuilder().
		From("logs").
		Where(syntax.And(
			syntax.Eq("level", "'debug'"),
			syntax.Lt("created_at", syntax.P(1)),
		)).
		ToSql()
	want := `DELETE FROM "logs" WHERE "level" = 'debug' AND "created_at" < ?1;`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlDeleteBuilder_ToSqlWithDialect(t *testing.T) {
	t.Parallel()
	d := &dialect.SimpleDialect{}
	got := newDeleteBuilder().
		From("users").
		Where(syntax.Eq("id", syntax.P(1))).
		ToSqlWithDialect(d)
	want := `DELETE FROM "users" WHERE "id" = ?1`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlDeleteBuilder_ImplementsSqlBuilder(t *testing.T) {
	t.Parallel()
	var _ builder.SqlBuilder = newDeleteBuilder()
}

func TestSqlDeleteBuilder_Returning(t *testing.T) {
	t.Parallel()
	got := newDeleteBuilder().
		From("sessions").
		Where(syntax.Eq("user_id", syntax.P(1))).
		Returning("id", "user_id").
		ToSql()
	want := `DELETE FROM "sessions" WHERE "user_id" = ?1 RETURNING "id", "user_id";`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlDeleteBuilder_ReturningWithoutWhere(t *testing.T) {
	t.Parallel()
	got := newDeleteBuilder().
		From("events").
		Returning("id").
		ToSql()
	want := `DELETE FROM "events" RETURNING "id";`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}
