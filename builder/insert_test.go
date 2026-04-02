package builder_test

import (
	"testing"

	"github.com/kotofurumiya/sqbl/builder"
	"github.com/kotofurumiya/sqbl/dialect"
	"github.com/kotofurumiya/sqbl/syntax"
)

func newInsertBuilder() *builder.SqlInsertBuilder {
	return (&builder.SqlInsertBuilder{}).Dialect(&dialect.SimpleDialect{})
}

func TestSqlInsertBuilder_SingleRow(t *testing.T) {
	t.Parallel()
	got := newInsertBuilder().
		Into("users").
		Columns("name", "email").
		Values(syntax.P(1), syntax.P(2)).
		ToSql()
	want := `INSERT INTO "users" ("name", "email") VALUES (?1, ?2);`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlInsertBuilder_MultipleRows(t *testing.T) {
	t.Parallel()
	got := newInsertBuilder().
		Into("users").
		Columns("name", "email").
		Values(syntax.P(1), syntax.P(2)).
		Values(syntax.P(3), syntax.P(4)).
		ToSql()
	want := `INSERT INTO "users" ("name", "email") VALUES (?1, ?2), (?3, ?4);`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlInsertBuilder_NoColumns(t *testing.T) {
	t.Parallel()
	got := newInsertBuilder().
		Into("users").
		Values(syntax.P(1), syntax.P(2)).
		ToSql()
	want := `INSERT INTO "users" VALUES (?1, ?2);`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlInsertBuilder_BoolValue(t *testing.T) {
	t.Parallel()
	got := newInsertBuilder().
		Into("users").
		Columns("name", "active").
		Values(syntax.P(1), true).
		ToSql()
	want := `INSERT INTO "users" ("name", "active") VALUES (?1, TRUE);`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlInsertBuilder_PositionalParam(t *testing.T) {
	t.Parallel()
	got := newInsertBuilder().
		Into("logs").
		Columns("message").
		Values(syntax.P()).
		ToSql()
	want := `INSERT INTO "logs" ("message") VALUES (?);`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlInsertBuilder_StringValue(t *testing.T) {
	t.Parallel()
	got := newInsertBuilder().
		Into("users").
		Columns("name").
		Values("'Alice'").
		ToSql()
	want := `INSERT INTO "users" ("name") VALUES ('Alice');`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlInsertBuilder_IntValue(t *testing.T) {
	t.Parallel()
	got := newInsertBuilder().
		Into("counters").
		Columns("n").
		Values(42).
		ToSql()
	want := `INSERT INTO "counters" ("n") VALUES (42);`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlInsertBuilder_ToSqlWithDialect(t *testing.T) {
	t.Parallel()
	d := &dialect.SimpleDialect{}
	got := newInsertBuilder().
		Into("users").
		Columns("name").
		Values(syntax.P(1)).
		ToSqlWithDialect(d)
	want := `INSERT INTO "users" ("name") VALUES (?1)`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlInsertBuilder_ImplementsSqlBuilder(t *testing.T) {
	t.Parallel()
	var _ builder.SqlBuilder = newInsertBuilder()
}

func TestSqlInsertBuilder_Returning(t *testing.T) {
	t.Parallel()
	got := newInsertBuilder().
		Into("users").
		Columns("name", "email").
		Values(syntax.P(1), syntax.P(2)).
		Returning("id", "created_at").
		ToSql()
	want := `INSERT INTO "users" ("name", "email") VALUES (?1, ?2) RETURNING "id", "created_at";`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlInsertBuilder_OnConflictDoNothing(t *testing.T) {
	t.Parallel()
	got := newInsertBuilder().
		Into("tags").
		Columns("name").
		Values(syntax.P(1)).
		OnConflictDoNothing().
		ToSql()
	want := `INSERT INTO "tags" ("name") VALUES (?1) ON CONFLICT DO NOTHING;`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlInsertBuilder_OnConflictDoNothing_WithTarget(t *testing.T) {
	t.Parallel()
	got := newInsertBuilder().
		Into("tags").
		Columns("name").
		Values(syntax.P(1)).
		OnConflict("name").
		OnConflictDoNothing().
		ToSql()
	want := `INSERT INTO "tags" ("name") VALUES (?1) ON CONFLICT ("name") DO NOTHING;`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlInsertBuilder_OnConflictDoUpdate(t *testing.T) {
	t.Parallel()
	got := newInsertBuilder().
		Into("user_settings").
		Columns("user_id", "key", "value").
		Values(syntax.P(1), syntax.P(2), syntax.P(3)).
		OnConflict("user_id", "key").
		DoUpdate("value = EXCLUDED.value").
		ToSql()
	want := `INSERT INTO "user_settings" ("user_id", "key", "value") VALUES (?1, ?2, ?3) ON CONFLICT ("user_id", "key") DO UPDATE SET value = EXCLUDED.value;`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlInsertBuilder_OnConflictDoUpdate_MultipleSet(t *testing.T) {
	t.Parallel()
	got := newInsertBuilder().
		Into("profiles").
		Columns("user_id", "name", "bio").
		Values(syntax.P(1), syntax.P(2), syntax.P(3)).
		OnConflict("user_id").
		DoUpdate("name = EXCLUDED.name", "bio = EXCLUDED.bio").
		ToSql()
	want := `INSERT INTO "profiles" ("user_id", "name", "bio") VALUES (?1, ?2, ?3) ON CONFLICT ("user_id") DO UPDATE SET name = EXCLUDED.name, bio = EXCLUDED.bio;`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlInsertBuilder_OnConflictDoNothing_WithReturning(t *testing.T) {
	t.Parallel()
	got := newInsertBuilder().
		Into("tags").
		Columns("name").
		Values(syntax.P(1)).
		OnConflictDoNothing().
		Returning("id").
		ToSql()
	want := `INSERT INTO "tags" ("name") VALUES (?1) ON CONFLICT DO NOTHING RETURNING "id";`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlInsertBuilder_OnDuplicateKeyUpdate(t *testing.T) {
	t.Parallel()
	got := (&builder.SqlInsertBuilder{}).
		Dialect(&dialect.MysqlDialect{}).
		Into("tags").
		Columns("name", "count").
		Values(syntax.P(1), syntax.P(2)).
		OnDuplicateKeyUpdate("count = VALUES(count)").
		ToSql()
	want := "INSERT INTO `tags` (`name`, `count`) VALUES (?, ?) ON DUPLICATE KEY UPDATE count = VALUES(count);"
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlInsertBuilder_MySQLReturning(t *testing.T) {
	t.Parallel()
	// RETURNING is output as-is regardless of dialect; the DB handles any errors.
	got := (&builder.SqlInsertBuilder{}).
		Dialect(&dialect.MysqlDialect{}).
		Into("users").
		Columns("name").
		Values(syntax.P(1)).
		Returning("id").
		ToSql()
	want := "INSERT INTO `users` (`name`) VALUES (?) RETURNING `id`;"
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlInsertBuilder_OrReplace(t *testing.T) {
	t.Parallel()
	got := (&builder.SqlInsertBuilder{}).
		Dialect(&dialect.SqliteDialect{}).
		Into("users").
		Columns("id", "name").
		Values(syntax.P(1), syntax.P(2)).
		OrReplace().
		ToSql()
	want := `INSERT OR REPLACE INTO "users" ("id", "name") VALUES (?, ?);`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlInsertBuilder_OrIgnore(t *testing.T) {
	t.Parallel()
	got := (&builder.SqlInsertBuilder{}).
		Dialect(&dialect.SqliteDialect{}).
		Into("tags").
		Columns("name").
		Values(syntax.P(1)).
		OrIgnore().
		ToSql()
	want := `INSERT OR IGNORE INTO "tags" ("name") VALUES (?);`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlInsertBuilder_SQLiteReturning(t *testing.T) {
	t.Parallel()
	got := (&builder.SqlInsertBuilder{}).
		Dialect(&dialect.SqliteDialect{}).
		Into("users").
		Columns("name").
		Values(syntax.P(1)).
		Returning("id").
		ToSql()
	want := `INSERT INTO "users" ("name") VALUES (?) RETURNING "id";`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlInsertBuilder_OrReplace_NonSQLiteDialect(t *testing.T) {
	t.Parallel()
	// OrReplace/OrIgnore are silently ignored for non-SQLite dialects.
	got := newInsertBuilder().
		Into("users").
		Columns("id", "name").
		Values(syntax.P(1), syntax.P(2)).
		OrReplace().
		ToSql()
	want := `INSERT INTO "users" ("id", "name") VALUES (?1, ?2);`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}
