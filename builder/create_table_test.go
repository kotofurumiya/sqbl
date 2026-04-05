package builder_test

import (
	"testing"

	"github.com/kotofurumiya/sqbl/builder"
	"github.com/kotofurumiya/sqbl/dialect"
	"github.com/kotofurumiya/sqbl/syntax"
)

func newCreateTableBuilder() builder.SqlCreateTableBuilder {
	return (builder.SqlCreateTableBuilder{}).Dialect(&dialect.SimpleDialect{})
}

func TestSqlCreateTableBuilder_ImplementsSqlBuilder(t *testing.T) {
	t.Parallel()
	var _ builder.SqlBuilder = newCreateTableBuilder()
}

func TestSqlCreateTableBuilder_Basic(t *testing.T) {
	t.Parallel()
	got := newCreateTableBuilder().Table("users").ToSql()
	want := `CREATE TABLE "users" ();`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlCreateTableBuilder_SingleColumn(t *testing.T) {
	t.Parallel()
	got := newCreateTableBuilder().Table("users").
		Column("id", "BIGINT").
		ToSql()
	want := `CREATE TABLE "users" ("id" BIGINT);`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlCreateTableBuilder_MultipleColumns(t *testing.T) {
	t.Parallel()
	got := newCreateTableBuilder().Table("users").
		Column("id", "BIGINT", "NOT NULL").
		Column("name", "TEXT", "NOT NULL").
		Column("email", "TEXT").
		ToSql()
	want := `CREATE TABLE "users" ("id" BIGINT NOT NULL, "name" TEXT NOT NULL, "email" TEXT);`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlCreateTableBuilder_ColumnMultipleConstraints(t *testing.T) {
	t.Parallel()
	got := newCreateTableBuilder().Table("t").
		Column("score", "INTEGER", "NOT NULL", "DEFAULT 0").
		ToSql()
	want := `CREATE TABLE "t" ("score" INTEGER NOT NULL DEFAULT 0);`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlCreateTableBuilder_IfNotExists(t *testing.T) {
	t.Parallel()
	got := newCreateTableBuilder().Table("users").IfNotExists().
		Column("id", "BIGINT", "NOT NULL").
		ToSql()
	want := `CREATE TABLE IF NOT EXISTS "users" ("id" BIGINT NOT NULL);`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlCreateTableBuilder_PrimaryKey(t *testing.T) {
	t.Parallel()
	got := newCreateTableBuilder().Table("users").
		Column("id", "BIGINT", "NOT NULL").
		Column("name", "TEXT", "NOT NULL").
		PrimaryKey("id").
		ToSql()
	want := `CREATE TABLE "users" ("id" BIGINT NOT NULL, "name" TEXT NOT NULL, PRIMARY KEY ("id"));`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlCreateTableBuilder_CompositePrimaryKey(t *testing.T) {
	t.Parallel()
	got := newCreateTableBuilder().Table("order_items").
		Column("order_id", "BIGINT", "NOT NULL").
		Column("item_id", "BIGINT", "NOT NULL").
		PrimaryKey("order_id", "item_id").
		ToSql()
	want := `CREATE TABLE "order_items" ("order_id" BIGINT NOT NULL, "item_id" BIGINT NOT NULL, PRIMARY KEY ("order_id", "item_id"));`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlCreateTableBuilder_Unique(t *testing.T) {
	t.Parallel()
	got := newCreateTableBuilder().Table("users").
		Column("id", "BIGINT", "NOT NULL").
		Column("email", "TEXT", "NOT NULL").
		PrimaryKey("id").
		Unique("email").
		ToSql()
	want := `CREATE TABLE "users" ("id" BIGINT NOT NULL, "email" TEXT NOT NULL, PRIMARY KEY ("id"), UNIQUE ("email"));`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlCreateTableBuilder_ForeignKey(t *testing.T) {
	t.Parallel()
	got := newCreateTableBuilder().Table("orders").
		Column("id", "BIGINT", "NOT NULL").
		Column("user_id", "BIGINT", "NOT NULL").
		PrimaryKey("id").
		ForeignKey([]string{"user_id"}, "users", []string{"id"}).
		ToSql()
	want := `CREATE TABLE "orders" ("id" BIGINT NOT NULL, "user_id" BIGINT NOT NULL, PRIMARY KEY ("id"), FOREIGN KEY ("user_id") REFERENCES "users" ("id"));`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlCreateTableBuilder_ForeignKeyWithActions(t *testing.T) {
	t.Parallel()
	got := newCreateTableBuilder().Table("orders").
		Column("user_id", "BIGINT", "NOT NULL").
		ForeignKey([]string{"user_id"}, "users", []string{"id"}, "ON DELETE CASCADE", "ON UPDATE RESTRICT").
		ToSql()
	want := `CREATE TABLE "orders" ("user_id" BIGINT NOT NULL, FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE RESTRICT);`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlCreateTableBuilder_CompositeForeignKey(t *testing.T) {
	t.Parallel()
	got := newCreateTableBuilder().Table("order_items").
		Column("order_id", "BIGINT", "NOT NULL").
		Column("product_id", "BIGINT", "NOT NULL").
		ForeignKey([]string{"order_id", "product_id"}, "orders", []string{"id", "product_id"}).
		ToSql()
	want := `CREATE TABLE "order_items" ("order_id" BIGINT NOT NULL, "product_id" BIGINT NOT NULL, FOREIGN KEY ("order_id", "product_id") REFERENCES "orders" ("id", "product_id"));`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlCreateTableBuilder_Check(t *testing.T) {
	t.Parallel()
	got := newCreateTableBuilder().Table("products").
		Column("id", "BIGINT", "NOT NULL").
		Column("price", "INTEGER", "NOT NULL").
		Check(syntax.Gt("price", 0)).
		ToSql()
	want := `CREATE TABLE "products" ("id" BIGINT NOT NULL, "price" INTEGER NOT NULL, CHECK ("price" > 0));`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlCreateTableBuilder_MultipleConstraints(t *testing.T) {
	t.Parallel()
	got := newCreateTableBuilder().Table("orders").
		Column("id", "BIGINT", "NOT NULL").
		Column("user_id", "BIGINT", "NOT NULL").
		Column("email", "TEXT", "NOT NULL").
		Column("total", "INTEGER", "NOT NULL").
		PrimaryKey("id").
		Unique("email").
		ForeignKey([]string{"user_id"}, "users", []string{"id"}, "ON DELETE CASCADE").
		Check(syntax.Gte("total", 0)).
		ToSql()
	want := `CREATE TABLE "orders" ("id" BIGINT NOT NULL, "user_id" BIGINT NOT NULL, "email" TEXT NOT NULL, "total" INTEGER NOT NULL, PRIMARY KEY ("id"), UNIQUE ("email"), FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE, CHECK ("total" >= 0));`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlCreateTableBuilder_ToSqlWithDialectNoSemicolon(t *testing.T) {
	t.Parallel()
	d := &dialect.SimpleDialect{}
	got := newCreateTableBuilder().Table("users").
		Column("id", "BIGINT", "NOT NULL").
		ToSqlWithDialect(d)
	want := `CREATE TABLE "users" ("id" BIGINT NOT NULL)`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlCreateTableBuilder_ToSqlWithDialectUsesGivenDialect(t *testing.T) {
	t.Parallel()
	d := &dialect.PostgresDialect{}
	got := newCreateTableBuilder().Table("users").
		Column("id", "BIGINT", "NOT NULL").
		ToSqlWithDialect(d)
	want := `CREATE TABLE "users" ("id" BIGINT NOT NULL)`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}
