package builder_test

import (
	"testing"

	"github.com/kotofurumiya/sqbl/builder"
	"github.com/kotofurumiya/sqbl/dialect"
)

func newAlterBuilder() builder.SqlAlterTableBuilder {
	return (builder.SqlAlterTableBuilder{}).Dialect(&dialect.SimpleDialect{}).Table("users")
}

func TestSqlAlterTableBuilder_AddColumn(t *testing.T) {
	t.Parallel()
	got := newAlterBuilder().AddColumn("bio", "TEXT").ToSql()
	want := `ALTER TABLE "users" ADD COLUMN "bio" TEXT;`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlAlterTableBuilder_DropColumn(t *testing.T) {
	t.Parallel()
	got := newAlterBuilder().DropColumn("legacy_col").ToSql()
	want := `ALTER TABLE "users" DROP COLUMN "legacy_col";`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlAlterTableBuilder_RenameColumn(t *testing.T) {
	t.Parallel()
	got := newAlterBuilder().RenameColumn("fullname", "name").ToSql()
	want := `ALTER TABLE "users" RENAME COLUMN "fullname" TO "name";`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlAlterTableBuilder_RenameTable(t *testing.T) {
	t.Parallel()
	got := newAlterBuilder().RenameTable("members").ToSql()
	want := `ALTER TABLE "users" RENAME TO "members";`
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSqlAlterTableBuilder_MySQL(t *testing.T) {
	t.Parallel()
	got := (&builder.SqlAlterTableBuilder{}).
		Dialect(&dialect.MysqlDialect{}).
		Table("users").
		AddColumn("bio", "TEXT").
		ToSql()
	want := "ALTER TABLE `users` ADD COLUMN `bio` TEXT;"
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}
