package shared

import (
	"slices"
	"testing"

	"github.com/kotofurumiya/sqbl/builder"
	"github.com/kotofurumiya/sqbl/syntax"
)

// runFiltering checks every condition helper against a product catalogue,
// together with DISTINCT and LIMIT/OFFSET paging.
func runFiltering(t *testing.T, conn Conn, s Suite) {
	MustExec(t, conn, `
		CREATE TEMPORARY TABLE products (
			id          INTEGER     PRIMARY KEY,
			name        VARCHAR(64) NOT NULL,
			category    VARCHAR(32) NOT NULL,
			price       INTEGER     NOT NULL,
			description VARCHAR(64)
		)
	`)
	MustExec(t, conn,
		s.InsertInto("products").
			Columns("id", "name", "category", "price", "description").
			Values(syntax.P(1), syntax.P(2), syntax.P(3), syntax.P(4), syntax.P(5)).
			Values(syntax.P(6), syntax.P(7), syntax.P(8), syntax.P(9), syntax.P(10)).
			Values(syntax.P(11), syntax.P(12), syntax.P(13), syntax.P(14), syntax.P(15)).
			Values(syntax.P(16), syntax.P(17), syntax.P(18), syntax.P(19), syntax.P(20)).
			Values(syntax.P(21), syntax.P(22), syntax.P(23), syntax.P(24), syntax.P(25)).
			Values(syntax.P(26), syntax.P(27), syntax.P(28), syntax.P(29), syntax.P(30)).
			MustSQL(),
		1, "Go Programming", "books", 3000, "A programming book",
		2, "Python Basics", "books", 2500, nil,
		3, "Mechanical Keyboard", "electronics", 15000, nil,
		4, "USB Hub", "electronics", 5000, "A 4-port hub",
		5, "Coffee Mug", "kitchen", 1200, nil,
		6, "Notebook", "stationery", 800, nil,
	)

	tests := []struct {
		name  string
		where syntax.SQLFragment
		args  []any
		want  []int
	}{
		{"Eq", syntax.Eq("category", syntax.P(1)), []any{"kitchen"}, []int{5}},
		{"Ne", syntax.Ne("category", syntax.P(1)), []any{"books"}, []int{3, 4, 5, 6}},
		{"Lt", syntax.Lt("price", syntax.P(1)), []any{1200}, []int{6}},
		{"Lte", syntax.Lte("price", syntax.P(1)), []any{1200}, []int{5, 6}},
		{"Gt", syntax.Gt("price", syntax.P(1)), []any{5000}, []int{3}},
		{"Gte", syntax.Gte("price", syntax.P(1)), []any{5000}, []int{3, 4}},
		{"Between", syntax.Between("price", syntax.P(1), syntax.P(2)), []any{1000, 3000}, []int{1, 2, 5}},
		{"In", syntax.In("category", syntax.P(1), syntax.P(2)), []any{"books", "kitchen"}, []int{1, 2, 5}},
		{"NotIn", syntax.NotIn("category", syntax.P(1), syntax.P(2)), []any{"books", "electronics"}, []int{5, 6}},
		{"IsNull", syntax.IsNull("description"), nil, []int{2, 3, 5, 6}},
		{"IsNotNull", syntax.IsNotNull("description"), nil, []int{1, 4}},
		{"Like", syntax.Like("name", syntax.P(1)), []any{"%Hub"}, []int{4}},
		{"Not", syntax.Not(syntax.Gt("price", syntax.P(1))), []any{3000}, []int{1, 2, 5, 6}},
		// Each OR branch matches a different book; the AND excludes the cheap mug.
		{
			"AndOr",
			syntax.And(syntax.Eq("category", syntax.P(1)), syntax.Or(syntax.Lt("price", syntax.P(2)), syntax.IsNotNull("description"))),
			[]any{"books", 2800},
			[]int{1, 2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql := s.From("products").Select("id").Where(tt.where).OrderBy(syntax.Asc("id")).MustSQL()
			AssertSlice(t, QueryColumn[int](t, conn, sql, tt.args...), tt.want)
		})
	}

	t.Run("Distinct", func(t *testing.T) {
		sql := s.From("products").Select("category").Distinct().OrderBy(syntax.Asc("category")).MustSQL()
		AssertSlice(t, QueryColumn[string](t, conn, sql),
			[]string{"books", "electronics", "kitchen", "stationery"})
	})

	t.Run("LimitOffset", func(t *testing.T) {
		page := s.From("products").Select("id").OrderBy(syntax.Desc("price")).Limit(2)
		AssertSlice(t, QueryColumn[int](t, conn, page.MustSQL()), []int{3, 4})
		AssertSlice(t, QueryColumn[int](t, conn, page.Offset(2).MustSQL()), []int{1, 2})
	})
}

// runJoins checks every join kind against customers and orders where Carol has
// no orders and order 4 has no customer, so outer joins differ from inner ones.
func runJoins(t *testing.T, conn Conn, s Suite) {
	seedCustomersAndOrders(t, conn, s)

	type pair struct {
		name    string
		orderID int
	}
	scanPair := func(r Rows) (pair, error) {
		var p pair
		return p, r.Scan(&p.name, &p.orderID)
	}
	// Outer joins leave one side NULL; COALESCE keeps the scan types simple.
	pairColumns := []any{
		syntax.As(syntax.Fn("COALESCE", "c.name", syntax.Raw("'?'")), "name"),
		syntax.As(syntax.Fn("COALESCE", "o.id", 0), "order_id"),
	}
	from := s.From(syntax.As("customers", "c")).Select(pairColumns...)
	on := syntax.Eq("c.id", "o.customer_id")

	t.Run("Inner", func(t *testing.T) {
		sql := from.InnerJoin(syntax.As("orders", "o"), on).OrderBy("order_id").MustSQL()
		AssertSlice(t, QueryRows(t, conn, scanPair, sql),
			[]pair{{"Alice", 1}, {"Alice", 2}, {"Bob", 3}})
	})

	t.Run("Left", func(t *testing.T) {
		sql := from.LeftJoin(syntax.As("orders", "o"), on).OrderBy("order_id").MustSQL()
		AssertSlice(t, QueryRows(t, conn, scanPair, sql),
			[]pair{{"Carol", 0}, {"Alice", 1}, {"Alice", 2}, {"Bob", 3}})
	})

	t.Run("Right", func(t *testing.T) {
		sql := from.RightJoin(syntax.As("orders", "o"), on).OrderBy("order_id").MustSQL()
		AssertSlice(t, QueryRows(t, conn, scanPair, sql),
			[]pair{{"Alice", 1}, {"Alice", 2}, {"Bob", 3}, {"?", 4}})
	})

	t.Run("Full", func(t *testing.T) {
		if !s.FullJoin {
			t.Skip("FULL OUTER JOIN is not supported")
		}
		sql := from.FullJoin(syntax.As("orders", "o"), on).OrderBy("order_id").MustSQL()
		AssertSlice(t, QueryRows(t, conn, scanPair, sql),
			[]pair{{"Carol", 0}, {"Alice", 1}, {"Alice", 2}, {"Bob", 3}, {"?", 4}})
	})

	t.Run("Cross", func(t *testing.T) {
		sql := s.From(syntax.As("customers", "a")).
			Select(syntax.Fn("COUNT", "*")).
			CrossJoin(syntax.As("customers", "b")).
			MustSQL()
		AssertSlice(t, QueryColumn[int](t, conn, sql), []int{9})
	})
}

// runSubqueries embeds a SELECT builder in each position a subquery may take.
func runSubqueries(t *testing.T, conn Conn, s Suite) {
	seedCustomersAndOrders(t, conn, s)

	type nameTotal struct {
		name  string
		total int
	}
	scanNameTotal := func(r Rows) (nameTotal, error) {
		var v nameTotal
		return v, r.Scan(&v.name, &v.total)
	}
	totals := s.From("orders").
		Select("customer_id", syntax.As(syntax.Fn("SUM", "amount"), "total")).
		GroupBy("customer_id")

	t.Run("From", func(t *testing.T) {
		sql := s.From(syntax.As(totals, "t")).
			Select("t.customer_id").
			Where(syntax.Gt("t.total", syntax.P(1))).
			OrderBy(syntax.Asc("t.customer_id")).
			MustSQL()
		AssertSlice(t, QueryColumn[int](t, conn, sql, 600), []int{1, 99})
	})

	t.Run("Join", func(t *testing.T) {
		sql := s.From(syntax.As("customers", "c")).
			Select("c.name", syntax.As(syntax.Fn("COALESCE", "t.total", 0), "total")).
			LeftJoin(syntax.As(totals, "t"), syntax.Eq("c.id", "t.customer_id")).
			OrderBy(syntax.Asc("c.id")).
			MustSQL()
		AssertSlice(t, QueryRows(t, conn, scanNameTotal, sql),
			[]nameTotal{{"Alice", 3000}, {"Bob", 500}, {"Carol", 0}})
	})

	t.Run("CorrelatedColumn", func(t *testing.T) {
		orderCount := s.From("orders").
			Select(syntax.Fn("COUNT", "*")).
			Where(syntax.Eq("orders.customer_id", "c.id"))
		sql := s.From(syntax.As("customers", "c")).
			Select("c.name", syntax.As(orderCount, "n")).
			OrderBy(syntax.Asc("c.id")).
			MustSQL()
		AssertSlice(t, QueryRows(t, conn, scanNameTotal, sql),
			[]nameTotal{{"Alice", 2}, {"Bob", 1}, {"Carol", 0}})
	})

	t.Run("WhereIn", func(t *testing.T) {
		buyers := s.From("orders").Select("customer_id")
		sql := s.From("customers").Select("name").Where(syntax.In("id", buyers)).OrderBy(syntax.Asc("id")).MustSQL()
		AssertSlice(t, QueryColumn[string](t, conn, sql), []string{"Alice", "Bob"})
	})
}

// seedCustomersAndOrders creates customers 1-3 (Alice, Bob, Carol) and orders
// 1-4, where Alice has two orders, Bob one, Carol none, and order 4 belongs to
// a customer that does not exist.
func seedCustomersAndOrders(t *testing.T, conn Conn, s Suite) {
	t.Helper()
	MustExec(t, conn, `CREATE TEMPORARY TABLE customers (id INTEGER PRIMARY KEY, name VARCHAR(32) NOT NULL)`)
	MustExec(t, conn, `
		CREATE TEMPORARY TABLE orders (
			id          INTEGER PRIMARY KEY,
			customer_id INTEGER NOT NULL,
			amount      INTEGER NOT NULL
		)
	`)
	MustExec(t, conn,
		s.InsertInto("customers").
			Columns("id", "name").
			Values(syntax.P(1), syntax.P(2)).
			Values(syntax.P(3), syntax.P(4)).
			Values(syntax.P(5), syntax.P(6)).
			MustSQL(),
		1, "Alice", 2, "Bob", 3, "Carol",
	)
	MustExec(t, conn,
		s.InsertInto("orders").
			Columns("id", "customer_id", "amount").
			Values(syntax.P(1), syntax.P(2), syntax.P(3)).
			Values(syntax.P(4), syntax.P(5), syntax.P(6)).
			Values(syntax.P(7), syntax.P(8), syntax.P(9)).
			Values(syntax.P(10), syntax.P(11), syntax.P(12)).
			MustSQL(),
		1, 1, 1000,
		2, 1, 2000,
		3, 2, 500,
		4, 99, 700,
	)
}

// runCTE aggregates a CTE-filtered sales table with GROUP BY and HAVING.
// Parameters 1 and 2 live inside the CTE and 3 in the outer HAVING, so the
// numbering must line up across the nested builder.
func runCTE(t *testing.T, conn Conn, s Suite) {
	MustExec(t, conn, `
		CREATE TEMPORARY TABLE sales (
			id       INTEGER     PRIMARY KEY,
			rep_name VARCHAR(32) NOT NULL,
			amount   INTEGER     NOT NULL,
			sold_at  DATE        NOT NULL
		)
	`)
	MustExec(t, conn,
		s.InsertInto("sales").
			Columns("id", "rep_name", "amount", "sold_at").
			Values(syntax.P(1), syntax.P(2), syntax.P(3), syntax.P(4)).
			Values(syntax.P(5), syntax.P(6), syntax.P(7), syntax.P(8)).
			Values(syntax.P(9), syntax.P(10), syntax.P(11), syntax.P(12)).
			Values(syntax.P(13), syntax.P(14), syntax.P(15), syntax.P(16)).
			Values(syntax.P(17), syntax.P(18), syntax.P(19), syntax.P(20)).
			Values(syntax.P(21), syntax.P(22), syntax.P(23), syntax.P(24)).
			MustSQL(),
		1, "Alice", 5000, "2025-01-10",
		2, "Alice", 3000, "2025-01-20",
		3, "Bob", 8000, "2025-01-15",
		4, "Carol", 4000, "2025-01-05", // in January but below the HAVING threshold
		5, "Carol", 9000, "2024-12-20", // outside January: removed by the CTE
		6, "Bob", 2000, "2025-02-01", // outside January: removed by the CTE
	)

	jan := s.From("sales").
		Select("rep_name", "amount").
		Where(syntax.Between("sold_at", syntax.P(1), syntax.P(2)))
	sql := s.From("jan_sales").
		With("jan_sales", jan).
		Select("rep_name", syntax.As(syntax.Fn("SUM", "amount"), "total")).
		GroupBy("rep_name").
		Having(syntax.Gte(syntax.Fn("SUM", "amount"), syntax.P(3))).
		OrderBy(syntax.Asc("rep_name")).
		MustSQL()

	type repTotal struct {
		name  string
		total int
	}
	got := QueryRows(t, conn, func(r Rows) (repTotal, error) {
		var v repTotal
		return v, r.Scan(&v.name, &v.total)
	}, sql, "2025-01-01", "2025-01-31", 5000)
	AssertSlice(t, got, []repTotal{{"Alice", 8000}, {"Bob", 8000}})
}

// runRecursiveCTE walks an org chart to collect every direct and indirect
// report of one manager.
func runRecursiveCTE(t *testing.T, conn Conn, s Suite) {
	MustExec(t, conn, `
		CREATE TEMPORARY TABLE employees (
			id         INTEGER PRIMARY KEY,
			manager_id INTEGER
		)
	`)
	// 1 manages 2 and 3; 2 manages 4; 4 manages 5.
	MustExec(t, conn,
		s.InsertInto("employees").
			Columns("id", "manager_id").
			Values(syntax.P(1), syntax.P(2)).
			Values(syntax.P(3), syntax.P(4)).
			Values(syntax.P(5), syntax.P(6)).
			Values(syntax.P(7), syntax.P(8)).
			Values(syntax.P(9), syntax.P(10)).
			MustSQL(),
		1, nil, 2, 1, 3, 1, 4, 2, 5, 4,
	)

	anchor := s.From("employees").Select("id").Where(syntax.Eq("manager_id", syntax.P(1)))
	step := s.From(syntax.As("employees", "e")).
		Select("e.id").
		InnerJoin("reports", syntax.Eq("e.manager_id", "reports.id"))
	sql := s.From("reports").
		WithRecursive("reports", anchor.UnionAll(step)).
		Select("id").
		OrderBy(syntax.Asc("id")).
		MustSQL()

	AssertSlice(t, QueryColumn[int](t, conn, sql, 2), []int{4, 5})
}

// runSetOperations combines two membership lists with each set operator.
// User 2 is listed twice in premium so that the ALL variants differ from the
// deduplicating ones.
func runSetOperations(t *testing.T, conn Conn, s Suite) {
	MustExec(t, conn, `CREATE TEMPORARY TABLE premium (user_id INTEGER NOT NULL)`)
	MustExec(t, conn, `CREATE TEMPORARY TABLE newsletter (user_id INTEGER NOT NULL)`)
	insertIDs := func(table string, ids ...any) {
		b := s.InsertInto(table).Columns("user_id")
		for i := range ids {
			b = b.Values(syntax.P(i + 1))
		}
		MustExec(t, conn, b.MustSQL(), ids...)
	}
	insertIDs("premium", 1, 2, 2, 3)
	insertIDs("newsletter", 2, 3, 4, 5)

	premium := s.From("premium").Select("user_id")
	newsletter := s.From("newsletter").Select("user_id")
	tests := []struct {
		name    string
		query   builder.SelectBuilder
		needAll bool
		want    []int
	}{
		{"Union", premium.Union(newsletter), false, []int{1, 2, 3, 4, 5}},
		{"UnionAll", premium.UnionAll(newsletter), false, []int{1, 2, 2, 2, 3, 3, 4, 5}},
		{"Intersect", premium.Intersect(newsletter), false, []int{2, 3}},
		{"IntersectAll", premium.IntersectAll(newsletter), true, []int{2, 3}},
		{"Except", premium.Except(newsletter), false, []int{1}},
		{"ExceptAll", premium.ExceptAll(newsletter), true, []int{1, 2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.needAll && !s.SetOpAll {
				t.Skip("INTERSECT ALL / EXCEPT ALL are not supported")
			}
			// A set operation has no overall ORDER BY in sqbl, so sort here.
			got := QueryColumn[int](t, conn, tt.query.MustSQL())
			slices.Sort(got)
			AssertSlice(t, got, tt.want)
		})
	}
}

// runWindowFunctions ranks and totals sales per rep, and keeps a running total
// with an explicit frame, all through OVER clauses.
func runWindowFunctions(t *testing.T, conn Conn, s Suite) {
	MustExec(t, conn, `
		CREATE TEMPORARY TABLE sales (
			id       INTEGER     PRIMARY KEY,
			rep_name VARCHAR(32) NOT NULL,
			amount   INTEGER     NOT NULL
		)
	`)
	MustExec(t, conn,
		s.InsertInto("sales").
			Columns("id", "rep_name", "amount").
			Values(syntax.P(1), syntax.P(2), syntax.P(3)).
			Values(syntax.P(4), syntax.P(5), syntax.P(6)).
			Values(syntax.P(7), syntax.P(8), syntax.P(9)).
			MustSQL(),
		1, "Alice", 3000,
		2, "Bob", 7000,
		3, "Alice", 5000,
	)

	sql := s.From("sales").
		Select(
			"id",
			syntax.As(syntax.Over(syntax.Fn("ROW_NUMBER")).PartitionBy("rep_name").OrderBy(syntax.Desc("amount")), "rank_in_rep"),
			syntax.As(syntax.Over(syntax.Fn("SUM", "amount")).PartitionBy("rep_name"), "rep_total"),
			syntax.As(syntax.Over(syntax.Fn("SUM", "amount")).OrderBy("id").Rows("BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW"), "running"),
		).
		OrderBy(syntax.Asc("id")).
		MustSQL()

	type row struct{ id, rank, repTotal, running int }
	got := QueryRows(t, conn, func(r Rows) (row, error) {
		var v row
		return v, r.Scan(&v.id, &v.rank, &v.repTotal, &v.running)
	}, sql)
	AssertSlice(t, got, []row{
		{1, 2, 8000, 3000},
		{2, 1, 7000, 10000},
		{3, 1, 8000, 15000},
	})
}
