package shared

import (
	"context"
	"testing"

	"github.com/kotofurumiya/sqbl/syntax"
)

// RunFilteringAndPagination simulates browsing a product catalogue.
// It exercises IN, NOT IN, IS NULL / IS NOT NULL, DISTINCT, and LIMIT+OFFSET
// as commonly combined in list-page queries.
func RunFilteringAndPagination(t *testing.T, conn Conn, s Suite) {
	t.Helper()
	ctx := context.Background()

	mustExec(t, conn, ctx, "create products", `
		CREATE TEMPORARY TABLE products (
			id          INTEGER NOT NULL,
			name        TEXT    NOT NULL,
			category    TEXT    NOT NULL,
			price       INTEGER NOT NULL,
			description TEXT
		)
	`)

	// Seed: 6 products across 4 categories; 2 have a description, 4 do not.
	insertSQL := s.InsertInto("products").
		Columns("id", "name", "category", "price", "description").
		Values(syntax.P(1), syntax.P(2), syntax.P(3), syntax.P(4), syntax.P(5)).
		ToSql()

	type product struct {
		id       int
		name     string
		category string
		price    int
		desc     any // nil becomes SQL NULL
	}
	seed := []product{
		{1, "Go Programming", "books", 3000, "A programming book"},
		{2, "Python Basics", "books", 2500, nil},
		{3, "Mechanical Keyboard", "electronics", 15000, nil},
		{4, "USB Hub", "electronics", 5000, "A 4-port hub"},
		{5, "Coffee Mug", "kitchen", 1200, nil},
		{6, "Notebook", "stationery", 800, nil},
	}
	for _, p := range seed {
		mustExec(t, conn, ctx, "insert", insertSQL, p.id, p.name, p.category, p.price, p.desc)
	}

	t.Run("IN", func(t *testing.T) {
		// category IN ('books', 'electronics') → ids 1, 2, 3, 4
		sql := s.From("products").
			Select("id").
			Where(syntax.In("category", syntax.P(1), syntax.P(2))).
			OrderBy(syntax.Asc("id")).
			ToSql()
		rows := mustQuery(t, conn, ctx, "IN", sql, "books", "electronics")
		got := collectRows(t, rows, func(r Rows) (int, error) {
			var id int
			return id, r.Scan(&id)
		})
		assertSlice(t, got, []int{1, 2, 3, 4})
	})

	t.Run("NOT IN", func(t *testing.T) {
		// category NOT IN ('kitchen', 'stationery') → ids 1, 2, 3, 4
		sql := s.From("products").
			Select("id").
			Where(syntax.NotIn("category", syntax.P(1), syntax.P(2))).
			OrderBy(syntax.Asc("id")).
			ToSql()
		rows := mustQuery(t, conn, ctx, "NOT IN", sql, "kitchen", "stationery")
		got := collectRows(t, rows, func(r Rows) (int, error) {
			var id int
			return id, r.Scan(&id)
		})
		assertSlice(t, got, []int{1, 2, 3, 4})
	})

	t.Run("IS NULL", func(t *testing.T) {
		// description IS NULL → ids 2, 3, 5, 6
		sql := s.From("products").
			Select("id").
			Where(syntax.IsNull("description")).
			OrderBy(syntax.Asc("id")).
			ToSql()
		rows := mustQuery(t, conn, ctx, "IS NULL", sql)
		got := collectRows(t, rows, func(r Rows) (int, error) {
			var id int
			return id, r.Scan(&id)
		})
		assertSlice(t, got, []int{2, 3, 5, 6})
	})

	t.Run("IS NOT NULL", func(t *testing.T) {
		// description IS NOT NULL → ids 1, 4
		sql := s.From("products").
			Select("id").
			Where(syntax.IsNotNull("description")).
			OrderBy(syntax.Asc("id")).
			ToSql()
		rows := mustQuery(t, conn, ctx, "IS NOT NULL", sql)
		got := collectRows(t, rows, func(r Rows) (int, error) {
			var id int
			return id, r.Scan(&id)
		})
		assertSlice(t, got, []int{1, 4})
	})

	t.Run("DISTINCT", func(t *testing.T) {
		// SELECT DISTINCT category → {books, electronics, kitchen, stationery}
		sql := s.From("products").
			Select("category").
			Distinct().
			OrderBy(syntax.Asc("category")).
			ToSql()
		rows := mustQuery(t, conn, ctx, "DISTINCT", sql)
		got := collectRows(t, rows, func(r Rows) (string, error) {
			var cat string
			return cat, r.Scan(&cat)
		})
		assertSlice(t, got, []string{"books", "electronics", "kitchen", "stationery"})
	})

	t.Run("LIMIT and OFFSET", func(t *testing.T) {
		// Page 1 (LIMIT 3): ids 1, 2, 3
		page1SQL := s.From("products").
			Select("id").
			OrderBy(syntax.Asc("id")).
			Limit(3).
			ToSql()
		rows1 := mustQuery(t, conn, ctx, "page 1", page1SQL)
		got1 := collectRows(t, rows1, func(r Rows) (int, error) {
			var id int
			return id, r.Scan(&id)
		})
		assertSlice(t, got1, []int{1, 2, 3})

		// Page 2 (LIMIT 3 OFFSET 3): ids 4, 5, 6
		page2SQL := s.From("products").
			Select("id").
			OrderBy(syntax.Asc("id")).
			Limit(3).
			Offset(3).
			ToSql()
		rows2 := mustQuery(t, conn, ctx, "page 2", page2SQL)
		got2 := collectRows(t, rows2, func(r Rows) (int, error) {
			var id int
			return id, r.Scan(&id)
		})
		assertSlice(t, got2, []int{4, 5, 6})
	})
}
