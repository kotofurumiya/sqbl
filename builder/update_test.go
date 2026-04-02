package builder_test

import (
	"testing"

	"github.com/kotofurumiya/sqbl/builder"
	"github.com/kotofurumiya/sqbl/dialect"
	"github.com/kotofurumiya/sqbl/syntax"
)

func newUpdateBuilder() *builder.SqlUpdateBuilder {
	return (&builder.SqlUpdateBuilder{}).Dialect(&dialect.SimpleDialect{})
}

func TestSqlUpdateBuilder_SingleSet(t *testing.T) {
	t.Parallel()
	got := newUpdateBuilder().
		Table("users").
		Set("name", syntax.P(1)).
		ToSql()
	want := `UPDATE "users" SET "name" = ?1;`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlUpdateBuilder_MultipleSets(t *testing.T) {
	t.Parallel()
	got := newUpdateBuilder().
		Table("users").
		Set("name", syntax.P(1)).
		Set("email", syntax.P(2)).
		ToSql()
	want := `UPDATE "users" SET "name" = ?1, "email" = ?2;`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlUpdateBuilder_WithWhere(t *testing.T) {
	t.Parallel()
	got := newUpdateBuilder().
		Table("users").
		Set("name", syntax.P(1)).
		Where(syntax.Eq("id", syntax.P(2))).
		ToSql()
	want := `UPDATE "users" SET "name" = ?1 WHERE "id" = ?2;`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlUpdateBuilder_BoolValue(t *testing.T) {
	t.Parallel()
	got := newUpdateBuilder().
		Table("users").
		Set("active", false).
		Where(syntax.Eq("id", syntax.P(1))).
		ToSql()
	want := `UPDATE "users" SET "active" = FALSE WHERE "id" = ?1;`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlUpdateBuilder_WhereAnd(t *testing.T) {
	t.Parallel()
	got := newUpdateBuilder().
		Table("users").
		Set("status", syntax.P(1)).
		Where(syntax.And(
			syntax.Eq("role", "'member'"),
			syntax.IsNull("deleted_at"),
		)).
		ToSql()
	want := `UPDATE "users" SET "status" = ?1 WHERE "role" = 'member' AND "deleted_at" IS NULL;`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlUpdateBuilder_ToSqlWithDialect(t *testing.T) {
	t.Parallel()
	d := &dialect.SimpleDialect{}
	got := newUpdateBuilder().
		Table("users").
		Set("name", syntax.P(1)).
		Where(syntax.Eq("id", syntax.P(2))).
		ToSqlWithDialect(d)
	want := `UPDATE "users" SET "name" = ?1 WHERE "id" = ?2`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlUpdateBuilder_ImplementsSqlBuilder(t *testing.T) {
	t.Parallel()
	var _ builder.SqlBuilder = newUpdateBuilder()
}

func TestSqlUpdateBuilder_Returning(t *testing.T) {
	t.Parallel()
	got := newUpdateBuilder().
		Table("users").
		Set("name", syntax.P(1)).
		Where(syntax.Eq("id", syntax.P(2))).
		Returning("id", "name").
		ToSql()
	want := `UPDATE "users" SET "name" = ?1 WHERE "id" = ?2 RETURNING "id", "name";`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlUpdateBuilder_ReturningWithoutWhere(t *testing.T) {
	t.Parallel()
	got := newUpdateBuilder().
		Table("counters").
		Set("n", syntax.P(1)).
		Returning("n").
		ToSql()
	want := `UPDATE "counters" SET "n" = ?1 RETURNING "n";`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}
