package builder_test

import (
	"testing"

	"github.com/kotofurumiya/sqbl/builder"
	"github.com/kotofurumiya/sqbl/dialect"
)

func TestSqlCreateDatabaseBuilder(t *testing.T) {
	t.Parallel()
	base := (builder.SqlCreateDatabaseBuilder{}).Name("app_test")

	tests := []struct {
		name string
		got  string
		want string
	}{
		{name: "basic", got: base.Dialect(&dialect.PostgresDialect{}).ToSql(), want: `CREATE DATABASE "app_test";`},
		{name: "if not exists", got: base.Dialect(&dialect.MysqlDialect{}).IfNotExists().ToSql(), want: "CREATE DATABASE IF NOT EXISTS `app_test`;"},
		{name: "dot is not a separator", got: base.Name("app.test").ToSqlWithDialect(&dialect.PostgresDialect{}), want: `CREATE DATABASE "app.test"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q; want %q", tt.got, tt.want)
			}
		})
	}
}

func TestSqlDropDatabaseBuilder(t *testing.T) {
	t.Parallel()
	base := (builder.SqlDropDatabaseBuilder{}).Dialect(&dialect.PostgresDialect{}).Name("app_test")

	tests := []struct {
		name string
		got  string
		want string
	}{
		{name: "basic", got: base.ToSql(), want: `DROP DATABASE "app_test";`},
		{name: "if exists", got: base.IfExists().ToSql(), want: `DROP DATABASE IF EXISTS "app_test";`},
		{name: "force", got: base.IfExists().Force().ToSql(), want: `DROP DATABASE IF EXISTS "app_test" WITH (FORCE);`},
		{name: "mysql", got: base.ToSqlWithDialect(&dialect.MysqlDialect{}), want: "DROP DATABASE `app_test`"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q; want %q", tt.got, tt.want)
			}
		})
	}
}
