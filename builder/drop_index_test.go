package builder_test

import (
	"testing"

	"github.com/kotofurumiya/sqbl/builder"
	"github.com/kotofurumiya/sqbl/dialect"
)

func TestSqlDropIndexBuilder_Basic_PostgreSQL(t *testing.T) {
	t.Parallel()
	got := (&builder.SqlDropIndexBuilder{}).
		Dialect(&dialect.PostgresDialect{}).
		Name("idx_users_email").
		ToSql()
	want := `DROP INDEX "idx_users_email";`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlDropIndexBuilder_IfExists_PostgreSQL(t *testing.T) {
	t.Parallel()
	got := (&builder.SqlDropIndexBuilder{}).
		Dialect(&dialect.PostgresDialect{}).
		Name("idx_users_email").
		IfExists().
		ToSql()
	want := `DROP INDEX IF EXISTS "idx_users_email";`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlDropIndexBuilder_MySQL(t *testing.T) {
	t.Parallel()
	got := (&builder.SqlDropIndexBuilder{}).
		Dialect(&dialect.MysqlDialect{}).
		Name("idx_users_email").
		On("users").
		ToSql()
	want := "DROP INDEX `idx_users_email` ON `users`;"
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlDropIndexBuilder_MySQL_IfExistsIgnored(t *testing.T) {
	t.Parallel()
	// MySQL does not support IF EXISTS for DROP INDEX; it should be omitted.
	got := (&builder.SqlDropIndexBuilder{}).
		Dialect(&dialect.MysqlDialect{}).
		Name("idx_users_email").
		IfExists().
		On("users").
		ToSql()
	want := "DROP INDEX `idx_users_email` ON `users`;"
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlDropIndexBuilder_SQLite(t *testing.T) {
	t.Parallel()
	got := (&builder.SqlDropIndexBuilder{}).
		Dialect(&dialect.SqliteDialect{}).
		Name("idx_users_email").
		IfExists().
		ToSql()
	want := `DROP INDEX IF EXISTS "idx_users_email";`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}
