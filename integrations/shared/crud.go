package shared

import (
	"database/sql"
	"testing"

	"github.com/kotofurumiya/sqbl/syntax"
)

// runCRUD drives a product row through INSERT, UPDATE, SELECT and DELETE.
// The active flag is a BOOLEAN written as a literal, so each dialect's boolean
// rendering must be accepted by its database in VALUES, SET and WHERE.
func runCRUD(t *testing.T, conn Conn, s Suite) {
	MustExec(t, conn, `
		CREATE TEMPORARY TABLE products (
			id     INTEGER     PRIMARY KEY,
			name   VARCHAR(64) NOT NULL,
			price  INTEGER     NOT NULL,
			active BOOLEAN     NOT NULL
		)
	`)

	type product struct {
		name  string
		price int
	}
	scanProduct := func(r Rows) (product, error) {
		var p product
		return p, r.Scan(&p.name, &p.price)
	}
	activeSQL := s.From("products").
		Select("name", "price").
		Where(syntax.Eq("active", true)).
		OrderBy(syntax.Asc("id")).
		ToSql()

	MustExec(t, conn,
		s.InsertInto("products").
			Columns("id", "name", "price", "active").
			Values(syntax.P(1), syntax.P(2), syntax.P(3), true).
			Values(syntax.P(4), syntax.P(5), syntax.P(6), true).
			Values(syntax.P(7), syntax.P(8), syntax.P(9), false).
			ToSql(),
		1, "Laptop", 150000,
		2, "Mouse", 3500,
		3, "Keyboard", 2000,
	)
	AssertSlice(t, QueryRows(t, conn, scanProduct, activeSQL),
		[]product{{"Laptop", 150000}, {"Mouse", 3500}})

	MustExec(t, conn,
		s.Update("products").
			Set("price", syntax.P(1)).
			Set("active", false).
			Where(syntax.Eq("name", syntax.P(2))).
			ToSql(),
		3000, "Mouse",
	)
	AssertSlice(t, QueryRows(t, conn, scanProduct, activeSQL),
		[]product{{"Laptop", 150000}})
	AssertSlice(t,
		QueryColumn[int](t, conn, s.From("products").Select("price").Where(syntax.Eq("name", syntax.P(1))).ToSql(), "Mouse"),
		[]int{3000})

	MustExec(t, conn, s.DeleteFrom("products").Where(syntax.Eq("active", false)).ToSql())
	AssertSlice(t, QueryColumn[int](t, conn, s.From("products").Select("id").ToSql()), []int{1})
}

// runParams proves that each placeholder style P supports is bound correctly
// by drivers that accept it.
func runParams(t *testing.T, conn Conn, s Suite) {
	MustExec(t, conn, `
		CREATE TEMPORARY TABLE events (
			id         INTEGER     PRIMARY KEY,
			name       VARCHAR(64) NOT NULL,
			created_at VARCHAR(32) NOT NULL,
			updated_at VARCHAR(32) NOT NULL
		)
	`)

	type event struct {
		id                   int
		name                 string
		createdAt, updatedAt string
	}
	assertEvent := func(t *testing.T, id int, want event) {
		t.Helper()
		var got event
		MustScan(t, conn,
			s.From("events").Select("id", "name", "created_at", "updated_at").Where(syntax.Eq("id", id)).ToSql(),
			nil, &got.id, &got.name, &got.createdAt, &got.updatedAt)
		if got != want {
			t.Errorf("got %+v; want %+v", got, want)
		}
	}
	insert := s.InsertInto("events").Columns("id", "name", "created_at", "updated_at")

	t.Run("positional", func(t *testing.T) {
		if !s.PositionalParams {
			t.Skip("driver does not accept positional placeholders")
		}
		MustExec(t, conn, insert.Values(syntax.P(), syntax.P(), syntax.P(), syntax.P()).ToSql(), 1, "alice", "t1", "t2")
		assertEvent(t, 1, event{1, "alice", "t1", "t2"})
	})

	t.Run("indexed reused out of order", func(t *testing.T) {
		if !s.IndexedParams {
			t.Skip("placeholders are not numbered")
		}
		MustExec(t, conn, insert.Values(syntax.P(3), syntax.P(2), syntax.P(1), syntax.P(1)).ToSql(), "t3", "bob", 2)
		assertEvent(t, 2, event{2, "bob", "t3", "t3"})
	})

	t.Run("named", func(t *testing.T) {
		if !s.NamedParams {
			t.Skip("driver does not bind named placeholders")
		}
		MustExec(t, conn, insert.Values(syntax.P(":id"), syntax.P(":name"), syntax.P(":at"), syntax.P(":at")).ToSql(),
			sql.Named("id", 3), sql.Named("name", "carol"), sql.Named("at", "t4"))
		assertEvent(t, 3, event{3, "carol", "t4", "t4"})
	})
}
