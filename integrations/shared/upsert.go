package shared

import (
	"errors"
	"testing"

	"github.com/kotofurumiya/sqbl/syntax"
)

// runReturning reads affected rows back through RETURNING on INSERT, UPDATE
// and DELETE.
func runReturning(t *testing.T, conn Conn, s Suite) {
	if !s.Returning {
		t.Skip("RETURNING is not supported on every statement")
	}
	MustExec(t, conn, `CREATE TEMPORARY TABLE articles (id INTEGER PRIMARY KEY, title VARCHAR(64) NOT NULL)`)

	var id int
	var title string
	MustScan(t, conn,
		s.InsertInto("articles").Columns("id", "title").Values(syntax.P(1), syntax.P(2)).Returning("id", "title").MustSQL(),
		[]any{1, "Hello"}, &id, &title)
	if id != 1 || title != "Hello" {
		t.Errorf("INSERT RETURNING = (%d, %q); want (1, %q)", id, title, "Hello")
	}

	MustScan(t, conn,
		s.Update("articles").Set("title", syntax.P(1)).Where(syntax.Eq("id", syntax.P(2))).Returning("title").MustSQL(),
		[]any{"Updated", 1}, &title)
	if title != "Updated" {
		t.Errorf("UPDATE RETURNING = %q; want %q", title, "Updated")
	}

	MustScan(t, conn,
		s.DeleteFrom("articles").Where(syntax.Eq("id", syntax.P(1))).Returning("id").MustSQL(),
		[]any{1}, &id)
	if id != 1 {
		t.Errorf("DELETE RETURNING = %d; want 1", id)
	}
	AssertCount(t, conn, "articles", 0)
}

// runOnConflict covers each ON CONFLICT form: ignoring the row, overwriting it
// with raw or structured assignments, overwriting only when a condition holds,
// and targeting a partial unique index.
func runOnConflict(t *testing.T, conn Conn, s Suite) {
	if !s.OnConflict {
		t.Skip("ON CONFLICT is not supported")
	}

	type user struct {
		id      int
		name    string
		version int
	}
	scanUser := func(r Rows) (user, error) {
		var u user
		return u, r.Scan(&u.id, &u.name, &u.version)
	}
	// newTable creates a users table holding only {1, Alice, 1} and returns a
	// query listing its rows.
	newTable := func(t *testing.T, table string) string {
		t.Helper()
		UseTables(t, conn, table)
		MustExec(t, conn, s.CreateTable(table).
			Column("id", "INTEGER", "PRIMARY KEY").
			Column("name", "VARCHAR(32)", "NOT NULL").
			Column("version", "INTEGER", "NOT NULL").
			MustSQL())
		MustExec(t, conn, s.InsertInto(table).Columns("id", "name", "version").Values(1, syntax.Raw("'Alice'"), 1).MustSQL())
		return s.From(table).Select("id", "name", "version").OrderBy(syntax.Asc("id")).MustSQL()
	}

	t.Run("DoNothing", func(t *testing.T) {
		const table = "it_oc_nothing"
		selectAll := newTable(t, table)
		sql := s.InsertInto(table).
			Columns("id", "name", "version").
			Values(syntax.P(1), syntax.P(2), syntax.P(3)).
			OnConflictDoNothing().
			Returning("id").
			MustSQL()
		var id int
		if err := conn.QueryRow(t.Context(), sql, 1, "Bob", 2).Scan(&id); !errors.Is(err, ErrNoRows) {
			t.Errorf("ignored insert returned %v; want ErrNoRows", err)
		}
		AssertSlice(t, QueryRows(t, conn, scanUser, selectAll), []user{{1, "Alice", 1}})
	})

	t.Run("DoUpdate", func(t *testing.T) {
		const table = "it_oc_update"
		selectAll := newTable(t, table)
		raw := s.InsertInto(table).
			Columns("id", "name", "version").
			Values(syntax.P(1), syntax.P(2), syntax.P(3)).
			OnConflict("id").
			DoUpdate("name = excluded.name").
			MustSQL()
		MustExec(t, conn, raw, 1, "Bob", 9)
		AssertSlice(t, QueryRows(t, conn, scanUser, selectAll), []user{{1, "Bob", 1}})

		structured := s.InsertInto(table).
			Columns("id", "name", "version").
			Values(syntax.P(1), syntax.P(2), syntax.P(3)).
			OnConflict("id").
			DoUpdateSet("name", "excluded.name").
			DoUpdateSet("version", syntax.Raw(table+".version + 1")).
			MustSQL()
		MustExec(t, conn, structured, 1, "Carol", 9)
		AssertSlice(t, QueryRows(t, conn, scanUser, selectAll), []user{{1, "Carol", 2}})
	})

	t.Run("DoUpdateWhere", func(t *testing.T) {
		const table = "it_oc_versioned"
		selectAll := newTable(t, table)
		// Accept a write only when it carries a newer version.
		sql := s.InsertInto(table).
			Columns("id", "name", "version").
			Values(syntax.P(1), syntax.P(2), syntax.P(3)).
			OnConflict("id").
			DoUpdateSet("name", "excluded.name").
			DoUpdateSet("version", "excluded.version").
			DoUpdateWhere(syntax.Lt(table+".version", "excluded.version")).
			MustSQL()
		MustExec(t, conn, sql, 1, "Stale", 0)
		AssertSlice(t, QueryRows(t, conn, scanUser, selectAll), []user{{1, "Alice", 1}})
		MustExec(t, conn, sql, 1, "Fresh", 2)
		AssertSlice(t, QueryRows(t, conn, scanUser, selectAll), []user{{1, "Fresh", 2}})
	})

	t.Run("OnConflictWhere", func(t *testing.T) {
		if !s.PartialIndex {
			t.Skip("partial indexes are not supported")
		}
		const table = "it_oc_partial"
		selectAll := newTable(t, table)
		// name is unique only among rows with version 1.
		MustExec(t, conn, s.CreateIndex(table+"_name").On(table).Columns("name").Unique().Where(syntax.Eq("version", 1)).MustSQL())
		sql := s.InsertInto(table).
			Columns("id", "name", "version").
			Values(syntax.P(1), syntax.P(2), syntax.P(3)).
			OnConflict("name").
			OnConflictWhere(syntax.Eq("version", 1)).
			DoUpdateSet("id", "excluded.id").
			MustSQL()
		// Conflicts with Alice inside the index, so her row takes the new id.
		MustExec(t, conn, sql, 2, "Alice", 1)
		// Version 2 rows fall outside the index, so this is a plain insert.
		MustExec(t, conn, sql, 3, "Alice", 2)
		AssertSlice(t, QueryRows(t, conn, scanUser, selectAll), []user{{2, "Alice", 1}, {3, "Alice", 2}})
	})
}
