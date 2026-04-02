package shared

import (
	"context"
	"testing"

	"github.com/kotofurumiya/sqbl/syntax"
)

// RunCRUD verifies INSERT, SELECT, UPDATE, and DELETE end-to-end using a
// product catalogue scenario. The table uses INTEGER for the active flag
// (1=active, 0=inactive) instead of BOOLEAN to remain portable across
// dialects that do not have a native boolean type.
//
// Explicit IDs are provided in the INSERT to avoid reliance on SERIAL or
// AUTO_INCREMENT, which have different syntax across databases.
func RunCRUD(t testing.TB, conn Conn, s Suite) {
	t.Helper()
	ctx := context.Background()

	// Use a TEMP TABLE so the test is isolated to this session and requires no
	// teardown even if the test fails mid-way.
	mustExec(t, conn, ctx, "create table", `
		CREATE TEMPORARY TABLE products (
			id       INTEGER PRIMARY KEY,
			name     TEXT    NOT NULL,
			category TEXT    NOT NULL,
			price    INTEGER NOT NULL,
			active   INTEGER NOT NULL DEFAULT 1
		)
	`)

	// Insert four products in a single statement.
	// Each row provides an explicit id to avoid reliance on auto-increment.
	insertSQL := s.InsertInto("products").
		Columns("id", "name", "category", "price", "active").
		Values(syntax.P(1), syntax.P(2), syntax.P(3), syntax.P(4), syntax.P(5)).
		Values(syntax.P(6), syntax.P(7), syntax.P(8), syntax.P(9), syntax.P(10)).
		Values(syntax.P(11), syntax.P(12), syntax.P(13), syntax.P(14), syntax.P(15)).
		Values(syntax.P(16), syntax.P(17), syntax.P(18), syntax.P(19), syntax.P(20)).
		ToSql()
	mustExec(t, conn, ctx, "insert", insertSQL,
		1, "Laptop Pro", "electronics", 150000, 1,
		2, "Wireless Mouse", "electronics", 3500, 1,
		3, "Desk Chair", "furniture", 45000, 1,
		4, "Old Keyboard", "electronics", 2000, 0,
	)

	// SELECT active electronics ordered by price descending.
	// "Old Keyboard" excluded (active=0); "Desk Chair" excluded (category=furniture).
	selectSQL := s.From("products").
		Select("name", "price").
		Where(syntax.And(
			syntax.Eq("category", syntax.P(1)),
			syntax.Eq("active", syntax.P(2)),
		)).
		OrderBy(syntax.Desc("price")).
		ToSql()

	type product struct {
		name  string
		price int
	}
	rows := mustQuery(t, conn, ctx, "select", selectSQL, "electronics", 1)
	got := collectRows(t, rows, func(r Rows) (product, error) {
		var p product
		return p, r.Scan(&p.name, &p.price)
	})
	assertSlice(t, got, []product{{"Laptop Pro", 150000}, {"Wireless Mouse", 3500}})

	// UPDATE the price of a single product and confirm the change persists.
	updateSQL := s.Update("products").
		Set("price", syntax.P(1)).
		Where(syntax.Eq("name", syntax.P(2))).
		ToSql()
	mustExec(t, conn, ctx, "update", updateSQL, 4500, "Wireless Mouse")

	var updatedPrice int
	checkSQL := s.From("products").Select("price").Where(syntax.Eq("name", syntax.P(1))).ToSql()
	if err := conn.QueryRow(ctx, checkSQL, "Wireless Mouse").Scan(&updatedPrice); err != nil {
		t.Fatalf("check update: %v", err)
	}
	if updatedPrice != 4500 {
		t.Errorf("price after update = %d; want 4500", updatedPrice)
	}

	// DELETE all inactive products and verify the final row count.
	deleteSQL := s.DeleteFrom("products").Where(syntax.Eq("active", syntax.P(1))).ToSql()
	mustExec(t, conn, ctx, "delete", deleteSQL, 0)

	var count int
	countSQL := s.From("products").Select("COUNT(*)").ToSql()
	if err := conn.QueryRow(ctx, countSQL).Scan(&count); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 3 {
		t.Errorf("count after delete = %d; want 3", count)
	}
}
