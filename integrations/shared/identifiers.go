package shared

import (
	"context"
	"testing"

	"github.com/kotofurumiya/sqbl/syntax"
)

// RunIdentifiers_ReservedKeywords verifies that the builder correctly
// double-quotes SQL reserved words when they are used as table or column names.
func RunIdentifiers_ReservedKeywords(t testing.TB, conn Conn, s Suite) {
	t.Helper()
	ctx := context.Background()

	// Table name "order" and column names "select", "from", "where", "group"
	// are all SQL reserved keywords. The raw CREATE uses quotes; the builder
	// must also emit quotes when building INSERT and SELECT statements.
	mustExec(t, conn, ctx, "create table", `
		CREATE TEMPORARY TABLE "order" (
			id       INTEGER PRIMARY KEY,
			"select" TEXT NOT NULL,
			"from"   TEXT NOT NULL,
			"where"  TEXT NOT NULL,
			"group"  INTEGER NOT NULL
		)
	`)

	mustExec(t, conn, ctx, "insert",
		s.InsertInto("order").
			Columns("id", "select", "from", "where", "group").
			Values(syntax.P(1), syntax.P(2), syntax.P(3), syntax.P(4), syntax.P(5)).
			ToSql(),
		1, "foo", "bar", "baz", 42,
	)

	// WHERE on a reserved-keyword column to confirm quoting in the condition too.
	selectSQL := s.From("order").
		Select("select", "from", "where", "group").
		Where(syntax.Eq("group", syntax.P(1))).
		ToSql()

	var sel, from, where string
	var group int
	if err := conn.QueryRow(ctx, selectSQL, 42).Scan(&sel, &from, &where, &group); err != nil {
		t.Fatalf("select from reserved-keyword table: %v", err)
	}
	if sel != "foo" || from != "bar" || where != "baz" || group != 42 {
		t.Errorf("got (%q, %q, %q, %d); want (foo, bar, baz, 42)", sel, from, where, group)
	}
}

// RunIdentifiers_MixedCase verifies that the builder preserves the exact casing
// of identifier names by quoting them. Without quotes most databases fold
// identifiers to lowercase.
func RunIdentifiers_MixedCase(t testing.TB, conn Conn, s Suite) {
	t.Helper()
	ctx := context.Background()

	mustExec(t, conn, ctx, "create table", `
		CREATE TEMPORARY TABLE "People" (
			id          INTEGER PRIMARY KEY,
			"FirstName" TEXT NOT NULL,
			"LastName"  TEXT NOT NULL
		)
	`)

	mustExec(t, conn, ctx, "insert",
		s.InsertInto("People").
			Columns("id", "FirstName", "LastName").
			Values(syntax.P(1), syntax.P(2), syntax.P(3)).
			ToSql(),
		1, "Alice", "Smith",
	)

	selectSQL := s.From("People").
		Select("FirstName", "LastName").
		Where(syntax.Eq("FirstName", syntax.P(1))).
		ToSql()

	var first, last string
	if err := conn.QueryRow(ctx, selectSQL, "Alice").Scan(&first, &last); err != nil {
		t.Fatalf("select: %v", err)
	}
	if first != "Alice" || last != "Smith" {
		t.Errorf("got (%q, %q); want (Alice, Smith)", first, last)
	}
}

// RunIdentifiers_EmbeddedDoubleQuote verifies that a column name containing a
// double-quote character is correctly escaped as "" inside the surrounding
// quotes.
func RunIdentifiers_EmbeddedDoubleQuote(t testing.TB, conn Conn, s Suite) {
	t.Helper()
	ctx := context.Background()

	// The column is literally named: my"col
	mustExec(t, conn, ctx, "create table", `
		CREATE TEMPORARY TABLE t (
			id        INTEGER PRIMARY KEY,
			"my""col" TEXT NOT NULL
		)
	`)

	mustExec(t, conn, ctx, "insert",
		s.InsertInto("t").
			Columns("id", `my"col`).
			Values(syntax.P(1), syntax.P(2)).
			ToSql(),
		1, "hello",
	)

	selectSQL := s.From("t").
		Select(`my"col`).
		Where(syntax.Eq(`my"col`, syntax.P(1))).
		ToSql()

	var val string
	if err := conn.QueryRow(ctx, selectSQL, "hello").Scan(&val); err != nil {
		t.Fatalf("select with double-quoted column name: %v", err)
	}
	if val != "hello" {
		t.Errorf("got %q; want %q", val, "hello")
	}
}
