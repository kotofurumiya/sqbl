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

func TestSqlDropIndexBuilder_IfExistsAndOn_AnyDialect(t *testing.T) {
	t.Parallel()
	// sqbl does not validate dialect support: IF EXISTS and ON are rendered whenever requested.
	tests := []struct {
		name string
		d    dialect.SqlDialect
		want string
	}{
		{name: "MySQL", d: &dialect.MysqlDialect{}, want: "DROP INDEX IF EXISTS `idx_users_email` ON `users`;"},
		{name: "PostgreSQL", d: &dialect.PostgresDialect{}, want: `DROP INDEX IF EXISTS "idx_users_email" ON "users";`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := (&builder.SqlDropIndexBuilder{}).
				Dialect(tt.d).
				Name("idx_users_email").
				IfExists().
				On("users").
				ToSql()
			if got != tt.want {
				t.Errorf("got %q; want %q", got, tt.want)
			}
		})
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
