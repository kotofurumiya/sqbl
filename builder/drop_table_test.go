package builder_test

import (
	"testing"

	"github.com/kotofurumiya/sqbl/builder"
	"github.com/kotofurumiya/sqbl/dialect"
)

func newDropTableBuilder(d dialect.SqlDialect) *builder.SqlDropTableBuilder {
	return (&builder.SqlDropTableBuilder{}).Dialect(d).Table("users")
}

func TestSqlDropTableBuilder_Basic(t *testing.T) {
	t.Parallel()
	got := newDropTableBuilder(&dialect.SimpleDialect{}).ToSql()
	want := `DROP TABLE "users";`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlDropTableBuilder_IfExists(t *testing.T) {
	t.Parallel()
	got := newDropTableBuilder(&dialect.SimpleDialect{}).IfExists().ToSql()
	want := `DROP TABLE IF EXISTS "users";`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlDropTableBuilder_Cascade_PostgreSQL(t *testing.T) {
	t.Parallel()
	got := newDropTableBuilder(&dialect.PostgresDialect{}).Cascade().ToSql()
	want := `DROP TABLE "users" CASCADE;`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlDropTableBuilder_Cascade_IgnoredForMySQL(t *testing.T) {
	t.Parallel()
	got := newDropTableBuilder(&dialect.MysqlDialect{}).Cascade().ToSql()
	// MySQL does not support CASCADE; it should be omitted
	want := "DROP TABLE `users`;"
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlDropTableBuilder_IfExistsCascade(t *testing.T) {
	t.Parallel()
	got := newDropTableBuilder(&dialect.PostgresDialect{}).IfExists().Cascade().ToSql()
	want := `DROP TABLE IF EXISTS "users" CASCADE;`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}
