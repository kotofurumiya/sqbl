package builder_test

import (
	"testing"

	"github.com/kotofurumiya/sqbl/builder"
	"github.com/kotofurumiya/sqbl/dialect"
	"github.com/kotofurumiya/sqbl/syntax"
)

func newCreateIndexBuilder() *builder.SqlCreateIndexBuilder {
	return (&builder.SqlCreateIndexBuilder{}).Dialect(&dialect.SimpleDialect{})
}

func TestSqlCreateIndexBuilder_ImplementsSqlBuilder(t *testing.T) {
	t.Parallel()
	var _ builder.SqlBuilder = newCreateIndexBuilder()
}

func TestSqlCreateIndexBuilder_Basic(t *testing.T) {
	t.Parallel()
	got := newCreateIndexBuilder().
		Name("idx_users_email").
		On("users").
		Columns("email").
		ToSql()
	want := `CREATE INDEX "idx_users_email" ON "users" ("email");`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlCreateIndexBuilder_Unique(t *testing.T) {
	t.Parallel()
	got := newCreateIndexBuilder().
		Name("idx_users_email").
		On("users").
		Columns("email").
		Unique().
		ToSql()
	want := `CREATE UNIQUE INDEX "idx_users_email" ON "users" ("email");`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlCreateIndexBuilder_IfNotExists(t *testing.T) {
	t.Parallel()
	got := newCreateIndexBuilder().
		Name("idx_users_email").
		On("users").
		Columns("email").
		IfNotExists().
		ToSql()
	want := `CREATE INDEX IF NOT EXISTS "idx_users_email" ON "users" ("email");`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlCreateIndexBuilder_UniqueIfNotExists(t *testing.T) {
	t.Parallel()
	got := newCreateIndexBuilder().
		Name("idx_users_email").
		On("users").
		Columns("email").
		Unique().
		IfNotExists().
		ToSql()
	want := `CREATE UNIQUE INDEX IF NOT EXISTS "idx_users_email" ON "users" ("email");`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlCreateIndexBuilder_MultipleColumns(t *testing.T) {
	t.Parallel()
	got := newCreateIndexBuilder().
		Name("idx_users_name").
		On("users").
		Columns("last_name", "first_name").
		ToSql()
	want := `CREATE INDEX "idx_users_name" ON "users" ("last_name", "first_name");`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlCreateIndexBuilder_Using(t *testing.T) {
	t.Parallel()
	got := newCreateIndexBuilder().
		Name("idx_users_email_hash").
		On("users").
		Columns("email").
		Using("hash").
		ToSql()
	want := `CREATE INDEX "idx_users_email_hash" ON "users" USING hash ("email");`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlCreateIndexBuilder_Where(t *testing.T) {
	t.Parallel()
	got := newCreateIndexBuilder().
		Name("idx_active_users").
		On("users").
		Columns("email").
		Where(syntax.Eq("active", true)).
		ToSql()
	want := `CREATE INDEX "idx_active_users" ON "users" ("email") WHERE "active" = TRUE;`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlCreateIndexBuilder_WhereIsNull(t *testing.T) {
	t.Parallel()
	got := newCreateIndexBuilder().
		Name("idx_non_deleted").
		On("users").
		Columns("email").
		Where(syntax.IsNull("deleted_at")).
		ToSql()
	want := `CREATE INDEX "idx_non_deleted" ON "users" ("email") WHERE "deleted_at" IS NULL;`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlCreateIndexBuilder_AllOptions(t *testing.T) {
	t.Parallel()
	got := newCreateIndexBuilder().
		Name("idx_active_users_email").
		On("users").
		Columns("email").
		Unique().
		IfNotExists().
		Using("btree").
		Where(syntax.Eq("active", true)).
		ToSql()
	want := `CREATE UNIQUE INDEX IF NOT EXISTS "idx_active_users_email" ON "users" USING btree ("email") WHERE "active" = TRUE;`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlCreateIndexBuilder_ToSqlWithDialectNoSemicolon(t *testing.T) {
	t.Parallel()
	d := &dialect.SimpleDialect{}
	got := newCreateIndexBuilder().
		Name("idx_users_email").
		On("users").
		Columns("email").
		ToSqlWithDialect(d)
	want := `CREATE INDEX "idx_users_email" ON "users" ("email")`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}
