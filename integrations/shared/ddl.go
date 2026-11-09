package shared

import (
	"testing"

	"github.com/kotofurumiya/sqbl/syntax"
)

// runCreateTable builds two related tables with every kind of constraint the
// builder offers and proves the database enforces each one.
func runCreateTable(t *testing.T, conn Conn, s Suite) {
	UseTables(t, conn, "it_ct_orders", "it_ct_users")

	createUsers := s.CreateTable("it_ct_users").
		IfNotExists().
		Column("id", "BIGINT", "NOT NULL").
		Column("email", "VARCHAR(255)", "NOT NULL", "UNIQUE").
		Column("status", "VARCHAR(16)", "NOT NULL", "DEFAULT 'active'").
		Column("age", "INTEGER").
		PrimaryKey("id").
		Check(syntax.Gte("age", 0)).
		MustSQL()
	MustExec(t, conn, createUsers)
	// IF NOT EXISTS makes a repeated CREATE a no-op instead of an error.
	MustExec(t, conn, createUsers)
	MustExec(t, conn,
		s.CreateTable("it_ct_orders").
			Column("id", "BIGINT", "NOT NULL").
			Column("user_id", "BIGINT", "NOT NULL").
			Column("code", "VARCHAR(32)", "NOT NULL").
			PrimaryKey("id").
			Unique("user_id", "code").
			ForeignKey([]string{"user_id"}, "it_ct_users", []string{"id"}, "ON DELETE CASCADE").
			MustSQL(),
	)

	insertUser := s.InsertInto("it_ct_users").Columns("id", "email", "age").Values(syntax.P(1), syntax.P(2), syntax.P(3)).MustSQL()
	insertOrder := s.InsertInto("it_ct_orders").Columns("id", "user_id", "code").Values(syntax.P(1), syntax.P(2), syntax.P(3)).MustSQL()

	MustExec(t, conn, insertUser, 1, "alice@example.com", 30)
	AssertSlice(t,
		QueryColumn[string](t, conn, s.From("it_ct_users").Select("status").MustSQL()),
		[]string{"active"})
	MustFail(t, conn, "column UNIQUE", insertUser, 2, "alice@example.com", 30)
	MustFail(t, conn, "CHECK age >= 0", insertUser, 3, "bob@example.com", -1)

	MustExec(t, conn, insertOrder, 1, 1, "A")
	MustFail(t, conn, "table UNIQUE (user_id, code)", insertOrder, 2, 1, "A")
	MustFail(t, conn, "FOREIGN KEY", insertOrder, 3, 999, "B")

	MustExec(t, conn, s.DeleteFrom("it_ct_users").Where(syntax.Eq("id", syntax.P(1))).MustSQL(), 1)
	AssertCount(t, conn, "it_ct_orders", 0)
}

// runCreateIndex proves that a UNIQUE index is enforced and, where supported,
// that a partial index applies only to rows matching its WHERE clause.
func runCreateIndex(t *testing.T, conn Conn, s Suite) {
	UseTables(t, conn, "it_ci_accounts")
	MustExec(t, conn, `
		CREATE TABLE it_ci_accounts (
			id     INTEGER      PRIMARY KEY,
			handle VARCHAR(32)  NOT NULL,
			email  VARCHAR(255) NOT NULL,
			active INTEGER      NOT NULL
		)
	`)
	insert := s.InsertInto("it_ci_accounts").
		Columns("id", "handle", "email", "active").
		Values(syntax.P(1), syntax.P(2), syntax.P(3), syntax.P(4)).
		MustSQL()

	t.Run("Unique", func(t *testing.T) {
		MustExec(t, conn,
			s.CreateIndex("it_ci_accounts_handle").On("it_ci_accounts").Columns("handle").Unique().IfNotExists().MustSQL())
		MustExec(t, conn, insert, 1, "alice", "alice@example.com", 0)
		MustFail(t, conn, "unique index on handle", insert, 2, "alice", "other@example.com", 0)
	})

	t.Run("Partial", func(t *testing.T) {
		if !s.PartialIndex {
			t.Skip("partial indexes are not supported")
		}
		MustExec(t, conn,
			s.CreateIndex("it_ci_accounts_active_email").
				On("it_ci_accounts").
				Columns("email").
				Unique().
				Where(syntax.Eq("active", 1)).
				MustSQL())
		// Inactive rows fall outside the index, so they may share an email.
		MustExec(t, conn, insert, 10, "bob1", "bob@example.com", 0)
		MustExec(t, conn, insert, 11, "bob2", "bob@example.com", 0)
		MustExec(t, conn, insert, 12, "bob3", "bob@example.com", 1)
		MustFail(t, conn, "partial unique index on active email", insert, 13, "bob4", "bob@example.com", 1)
	})
}

// runSchemaEvolution walks a table through a migration: every ALTER TABLE
// form, an index created and dropped, and finally DROP TABLE.
func runSchemaEvolution(t *testing.T, conn Conn, s Suite) {
	const (
		table   = "it_se_users"
		renamed = "it_se_people"
		index   = "it_se_people_name"
	)
	UseTables(t, conn, table, renamed)

	MustExec(t, conn,
		s.CreateTable(table).
			Column("id", "INTEGER", "NOT NULL").
			Column("name", "VARCHAR(32)", "NOT NULL").
			PrimaryKey("id").
			MustSQL())
	MustExec(t, conn, s.InsertInto(table).Columns("id", "name").Values(syntax.P(1), syntax.P(2)).MustSQL(), 1, "Alice")

	MustExec(t, conn, s.AlterTable(table).AddColumn("bio", "VARCHAR(64)").MustSQL())
	MustExec(t, conn, s.Update(table).Set("bio", syntax.P(1)).MustSQL(), "hello")
	MustExec(t, conn, s.AlterTable(table).RenameColumn("bio", "about").MustSQL())
	AssertSlice(t, QueryColumn[string](t, conn, s.From(table).Select("about").MustSQL()), []string{"hello"})

	MustExec(t, conn, s.AlterTable(table).DropColumn("about").MustSQL())
	// An UPDATE target, unlike a SELECT expression, cannot fall back to a
	// string literal in SQLite, so it reliably fails on a missing column.
	MustFail(t, conn, "dropped column", s.Update(table).Set("about", syntax.P(1)).MustSQL(), "x")

	MustExec(t, conn, s.AlterTable(table).RenameTable(renamed).MustSQL())
	MustFail(t, conn, "old table name", s.From(table).Select("id").MustSQL())
	AssertSlice(t, QueryColumn[string](t, conn, s.From(renamed).Select("name").MustSQL()), []string{"Alice"})

	MustExec(t, conn, s.CreateIndex(index).On(renamed).Columns("name").MustSQL())
	dropIndex := s.DropIndex(index).IfExists()
	if s.DropIndexOnTable {
		dropIndex = dropIndex.On(renamed)
	}
	MustExec(t, conn, dropIndex.MustSQL())
	// IF EXISTS makes dropping the already-dropped index a no-op.
	MustExec(t, conn, dropIndex.MustSQL())

	MustExec(t, conn, s.DropTable(renamed).IfExists().MustSQL())
	MustFail(t, conn, "dropped table", s.From(renamed).Select("id").MustSQL())
}
