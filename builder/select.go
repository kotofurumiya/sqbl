package builder

import (
	"fmt"
	"strings"

	"github.com/kotofurumiya/sqbl/dialect"
	"github.com/kotofurumiya/sqbl/syntax"
)

// cteClause represents a single WITH ... AS (...) entry.
type cteClause struct {
	Name      string
	Builder   *SqlSelectBuilder
	Recursive bool
}

// setOpClause represents a single set operation entry (UNION, INTERSECT, or EXCEPT).
// Op is one of "UNION", "INTERSECT", or "EXCEPT".
type setOpClause struct {
	Op      string
	All     bool
	Builder *SqlSelectBuilder
}

// SqlSelectBuilder constructs a SQL SELECT statement using a fluent method chain.
// Use a dialect-specific constructor (e.g. sqblpg.From) to create an instance.
//
//	q := sqblpg.From("users").
//	    Select("id", "name").
//	    Where(sqblpg.Eq("active", true)).
//	    OrderBy("name ASC").
//	    Limit(20)
//	sql := q.ToSql()
type SqlSelectBuilder struct {
	dialect      dialect.SqlDialect
	ctes         []cteClause          // WITH ...
	distinct     bool                 // DISTINCT
	distinctOn   []string             // DISTINCT ON (col1, col2) — PostgreSQL only
	columns      []syntax.SqlFragment // SELECT ... (StringSource | *Aliased)
	from         syntax.SqlFragment   // FROM (StringSource | *SqlSelectBuilder | *Aliased)
	joins        []syntax.JoinClause  // JOIN ...
	where        syntax.SqlFragment   // WHERE ...
	groups       []syntax.SqlFragment // GROUP BY ...
	having       syntax.SqlFragment   // HAVING ...
	orders       []syntax.SqlFragment // ORDER BY ...
	limit        *int                 // LIMIT
	offset       *int                 // OFFSET
	lock         string               // FOR UPDATE, FOR NO KEY UPDATE, FOR SHARE, FOR KEY SHARE
	lockOption   string               // NOWAIT, SKIP LOCKED
	setOps       []setOpClause        // UNION, INTERSECT, EXCEPT (and ALL variants)
}

var _ syntax.SqlFragment = (*SqlSelectBuilder)(nil)

// ToSql renders the query using the dialect set on this builder.
//
//	sql := sqblpg.From("users").Select("id").ToSql()
func (b *SqlSelectBuilder) ToSql() string {
	return b.renderSQL(b.dialect) + ";"
}

// ToSqlWithDialect implements syntax.SqlFragment for subquery embedding.
func (b *SqlSelectBuilder) ToSqlWithDialect(d dialect.SqlDialect) string {
	return "(" + b.renderSQL(d) + ")"
}

// renderSQL renders the query as a SQL string without a trailing semicolon.
func (b *SqlSelectBuilder) renderSQL(d dialect.SqlDialect) string {
	var parts []string

	// WITH [RECURSIVE]
	if len(b.ctes) > 0 {
		withKw := "WITH"
		for _, c := range b.ctes {
			if c.Recursive {
				withKw = "WITH RECURSIVE"
				break
			}
		}
		ctes := mapJoin(b.ctes, ", ", func(c cteClause) string {
			return c.Name + " AS (" + c.Builder.renderSQL(d) + ")"
		})
		parts = append(parts, withKw+" "+ctes)
	}

	// SELECT [DISTINCT [ON (cols)]]
	selectKw := "SELECT"
	if len(b.distinctOn) > 0 {
		if _, isPg := d.(*dialect.PostgresDialect); isPg {
			cols := make([]string, 0, len(b.distinctOn))
			for _, col := range b.distinctOn {
				cols = append(cols, d.QuoteIdentifier(col))
			}
			selectKw = "SELECT DISTINCT ON (" + strings.Join(cols, ", ") + ")"
		}
	} else if b.distinct {
		selectKw = "SELECT DISTINCT"
	}
	if cols := mapJoin(b.columns, ", ", func(f syntax.SqlFragment) string { return f.ToSqlWithDialect(d) }); cols != "" {
		parts = append(parts, selectKw+" "+cols)
	} else {
		parts = append(parts, selectKw+" *")
	}

	// FROM
	if b.from != nil {
		parts = append(parts, "FROM "+b.from.ToSqlWithDialect(d))
	}

	// JOINs
	for _, j := range b.joins {
		lateral := ""
		if j.Lateral {
			lateral = " LATERAL"
		}
		switch {
		case j.Kind == "CROSS":
			parts = append(parts, "CROSS"+lateral+" JOIN "+j.Table.ToSqlWithDialect(d))
		case j.Lateral:
			parts = append(parts, j.Kind+lateral+" JOIN "+j.Table.ToSqlWithDialect(d)+" ON TRUE")
		default:
			parts = append(parts, j.Kind+" JOIN "+j.Table.ToSqlWithDialect(d)+" ON "+j.On.ToSqlWithDialect(d))
		}
	}

	// WHERE
	if b.where != nil {
		parts = append(parts, "WHERE "+b.where.ToSqlWithDialect(d))
	}

	// GROUP BY
	if groups := mapJoin(b.groups, ", ", func(f syntax.SqlFragment) string { return f.ToSqlWithDialect(d) }); groups != "" {
		parts = append(parts, "GROUP BY "+groups)
	}

	// HAVING
	if b.having != nil {
		parts = append(parts, "HAVING "+b.having.ToSqlWithDialect(d))
	}

	// ORDER BY
	if orders := mapJoin(b.orders, ", ", func(f syntax.SqlFragment) string { return f.ToSqlWithDialect(d) }); orders != "" {
		parts = append(parts, "ORDER BY "+orders)
	}

	// LIMIT / OFFSET
	if b.limit != nil {
		parts = append(parts, fmt.Sprintf("LIMIT %d", *b.limit))
	}
	if b.offset != nil {
		parts = append(parts, fmt.Sprintf("OFFSET %d", *b.offset))
	}

	// FOR UPDATE/SHARE/etc
	if b.lock != "" {
		lock := b.lock
		if b.lockOption != "" {
			lock += " " + b.lockOption
		}
		parts = append(parts, lock)
	}

	// UNION / INTERSECT / EXCEPT
	for _, u := range b.setOps {
		op := u.Op
		if u.All {
			op += " ALL"
		}
		parts = append(parts, op, u.Builder.renderSQL(d))
	}

	return strings.Join(parts, " ")
}

// Dialect sets the SQL dialect used when rendering the query.
// Dialect-specific constructors (e.g. sqblpg.From) call this automatically.
//
//	b.Dialect(pgDialect)
func (b *SqlSelectBuilder) Dialect(d dialect.SqlDialect) *SqlSelectBuilder {
	b2 := *b
	b2.dialect = d
	return &b2
}

// From sets the FROM clause.
// Accepts a table name string, an aliased expression, or a subquery.
//
//	sqblpg.From("users")
//	sqblpg.From(sqbl.As("users", "u"))
//	sqblpg.From(sqbl.As(subquery, "sub"))
func (b *SqlSelectBuilder) From(table any) *SqlSelectBuilder {
	b2 := *b
	b2.from = syntax.ToFragment(table)
	return &b2
}

// Select sets the columns to retrieve.
// Accepts column name strings, aliased expressions, or raw SQL fragments.
// Calling Select again replaces the previous column list.
//
//	b.Select("id", "name")
//	b.Select(sqbl.As("SUM(amount)", "total"))
func (b *SqlSelectBuilder) Select(columns ...any) *SqlSelectBuilder {
	b2 := *b
	cols := make([]syntax.SqlFragment, len(columns))
	for i, col := range columns {
		cols[i] = syntax.ToFragment(col)
	}
	b2.columns = cols
	return &b2
}

// Join adds an INNER JOIN clause. It is an alias for InnerJoin.
//
//	b.Join(sqbl.As("orders", "o"), "u.id = o.user_id")
//	// INNER JOIN orders AS o ON u.id = o.user_id
func (b *SqlSelectBuilder) Join(table any, on syntax.SqlFragment) *SqlSelectBuilder {
	return b.InnerJoin(table, on)
}

// Where sets the WHERE condition.
// Use dialect-specific helpers (e.g. sqblpg.Eq, sqblpg.Gt) to build conditions.
// Combine multiple conditions explicitly with And or Or.
//
//	b.Where(sqblpg.Eq("status", "active"))
//	b.Where(sqblpg.And(sqblpg.Eq("status", "active"), sqblpg.Gt("age", 18)))
func (b *SqlSelectBuilder) Where(expr syntax.SqlFragment) *SqlSelectBuilder {
	b2 := *b
	b2.where = expr
	return &b2
}

// Distinct adds the DISTINCT keyword to the SELECT clause.
//
//	sqblpg.From("users").Select("country").Distinct()
//	// SELECT DISTINCT country FROM users
func (b *SqlSelectBuilder) Distinct() *SqlSelectBuilder {
	b2 := *b
	b2.distinct = true
	return &b2
}

// DistinctOn sets the DISTINCT ON columns (PostgreSQL only).
// Rows are deduplicated based on the specified columns; the first row for each group is kept.
// ORDER BY should be used to control which row is considered "first".
//
//	sqblpg.From("orders").Select("user_id", "amount").DistinctOn("user_id").OrderBy(sqblpg.Desc("amount"))
//	// SELECT DISTINCT ON ("user_id") "user_id", "amount" FROM "orders" ORDER BY "amount" DESC
func (b *SqlSelectBuilder) DistinctOn(cols ...string) *SqlSelectBuilder {
	b2 := *b
	b2.distinctOn = cols
	return &b2
}

// GroupBy sets the GROUP BY columns or expressions.
// Accepts column name strings, aliased expressions, or raw SQL fragments.
//
//	b.GroupBy("department", "role")
//	b.GroupBy(sqbl.As("EXTRACT(year FROM created_at)", "year"))
func (b *SqlSelectBuilder) GroupBy(cols ...any) *SqlSelectBuilder {
	b2 := *b
	groups := make([]syntax.SqlFragment, len(cols))
	for i, col := range cols {
		groups[i] = syntax.ToFragment(col)
	}
	b2.groups = groups
	return &b2
}

// Having sets the HAVING condition, applied after grouping.
// Use dialect-specific helpers to build conditions.
// Combine multiple conditions explicitly with And or Or.
//
//	b.Having(sqblpg.Gt("SUM(amount)", 1000))
//	// HAVING SUM(amount) > 1000
func (b *SqlSelectBuilder) Having(expr syntax.SqlFragment) *SqlSelectBuilder {
	b2 := *b
	b2.having = expr
	return &b2
}

// OrderBy sets the ORDER BY columns.
// Accepts column name strings or Order expressions (syntax.Asc / syntax.Desc).
//
//	b.OrderBy(syntax.Asc("name"), syntax.Desc("created_at"))
//	// ORDER BY name ASC, created_at DESC
func (b *SqlSelectBuilder) OrderBy(cols ...any) *SqlSelectBuilder {
	b2 := *b
	orders := make([]syntax.SqlFragment, len(cols))
	for i, col := range cols {
		orders[i] = syntax.ToFragment(col)
	}
	b2.orders = orders
	return &b2
}

// Limit sets the maximum number of rows to return.
//
//	b.Limit(20)
//	// LIMIT 20
func (b *SqlSelectBuilder) Limit(n int) *SqlSelectBuilder {
	b2 := *b
	b2.limit = &n
	return &b2
}

// Offset sets the number of rows to skip before returning results.
//
//	b.Limit(20).Offset(40)
//	// LIMIT 20 OFFSET 40
func (b *SqlSelectBuilder) Offset(n int) *SqlSelectBuilder {
	b2 := *b
	b2.offset = &n
	return &b2
}

// ForUpdate appends a FOR UPDATE locking clause to the query.
// Locks selected rows for the duration of the current transaction.
//
//	b.ForUpdate()
//	// SELECT ... FOR UPDATE
func (b *SqlSelectBuilder) ForUpdate() *SqlSelectBuilder {
	b2 := *b
	b2.lock = "FOR UPDATE"
	return &b2
}

// ForShare appends a FOR SHARE locking clause to the query.
// Allows concurrent reads but prevents modification of selected rows.
//
//	b.ForShare()
//	// SELECT ... FOR SHARE
func (b *SqlSelectBuilder) ForShare() *SqlSelectBuilder {
	b2 := *b
	b2.lock = "FOR SHARE"
	return &b2
}

// ForNoKeyUpdate appends a FOR NO KEY UPDATE locking clause.
// Weaker than FOR UPDATE; does not block INSERT of rows referencing this table.
//
//	b.ForNoKeyUpdate()
//	// SELECT ... FOR NO KEY UPDATE
func (b *SqlSelectBuilder) ForNoKeyUpdate() *SqlSelectBuilder {
	b2 := *b
	b2.lock = "FOR NO KEY UPDATE"
	return &b2
}

// ForKeyShare appends a FOR KEY SHARE locking clause.
// Weakest lock; only blocks DELETE and UPDATE of key columns.
//
//	b.ForKeyShare()
//	// SELECT ... FOR KEY SHARE
func (b *SqlSelectBuilder) ForKeyShare() *SqlSelectBuilder {
	b2 := *b
	b2.lock = "FOR KEY SHARE"
	return &b2
}

// Nowait adds the NOWAIT option to the locking clause.
// The query fails immediately if any selected row cannot be locked.
//
//	b.ForUpdate().Nowait()
//	// SELECT ... FOR UPDATE NOWAIT
func (b *SqlSelectBuilder) Nowait() *SqlSelectBuilder {
	b2 := *b
	b2.lockOption = "NOWAIT"
	return &b2
}

// SkipLocked adds the SKIP LOCKED option to the locking clause.
// Rows that cannot be immediately locked are silently skipped.
//
//	b.ForUpdate().SkipLocked()
//	// SELECT ... FOR UPDATE SKIP LOCKED
func (b *SqlSelectBuilder) SkipLocked() *SqlSelectBuilder {
	b2 := *b
	b2.lockOption = "SKIP LOCKED"
	return &b2
}

// InnerJoin adds an INNER JOIN clause.
//
//	b.InnerJoin(sqbl.As("orders", "o"), "u.id = o.user_id")
//	// INNER JOIN orders AS o ON u.id = o.user_id
func (b *SqlSelectBuilder) InnerJoin(table any, on syntax.SqlFragment) *SqlSelectBuilder {
	b2 := *b
	b2.joins = append(append([]syntax.JoinClause(nil), b.joins...), syntax.JoinClause{
		Kind:  "INNER",
		Table: syntax.ToFragment(table),
		On:    on,
	})
	return &b2
}

// LeftJoin adds a LEFT JOIN clause.
//
//	b.LeftJoin(sqbl.As("profiles", "p"), sqbl.Eq("u.id", "p.user_id"))
//	// LEFT JOIN profiles AS p ON u.id = p.user_id
func (b *SqlSelectBuilder) LeftJoin(table any, on syntax.SqlFragment) *SqlSelectBuilder {
	b2 := *b
	b2.joins = append(append([]syntax.JoinClause(nil), b.joins...), syntax.JoinClause{
		Kind:  "LEFT",
		Table: syntax.ToFragment(table),
		On:    on,
	})
	return &b2
}

// CrossJoin adds a CROSS JOIN clause.
// No ON condition is required for a cross join.
//
//	b.CrossJoin("sizes")
//	// CROSS JOIN sizes
func (b *SqlSelectBuilder) CrossJoin(table any) *SqlSelectBuilder {
	b2 := *b
	b2.joins = append(append([]syntax.JoinClause(nil), b.joins...), syntax.JoinClause{
		Kind:  "CROSS",
		Table: syntax.ToFragment(table),
		On:    nil,
	})
	return &b2
}

// With adds a Common Table Expression (CTE) to the query.
// The CTE is rendered as a WITH clause before the main SELECT.
//
//	recent := sqblpg.From("orders").Where(sqblpg.Gte("created_at", "2025-01-01"))
//	b.With("recent_orders", recent)
//	// WITH recent_orders AS (SELECT ...) SELECT ...
func (b *SqlSelectBuilder) With(name string, q *SqlSelectBuilder) *SqlSelectBuilder {
	b2 := *b
	b2.ctes = append(append([]cteClause(nil), b.ctes...), cteClause{Name: name, Builder: q})
	return &b2
}

// WithRecursive adds a recursive CTE to the query.
// The WITH clause is rendered as WITH RECURSIVE when this is used.
// Use this for hierarchical or iterative queries.
//
//	base := sqblpg.From("employees").Select("id", "manager_id").Where(sqblpg.IsNull("manager_id"))
//	recursive := sqblpg.From("employees").Join(sqblpg.As("org", "o"), sqblpg.Eq("employees.manager_id", "o.id")).Select("employees.id", "employees.manager_id")
//	b.WithRecursive("org", base.Union(recursive))
//	// WITH RECURSIVE org AS (...) SELECT ...
func (b *SqlSelectBuilder) WithRecursive(name string, q *SqlSelectBuilder) *SqlSelectBuilder {
	b2 := *b
	b2.ctes = append(append([]cteClause(nil), b.ctes...), cteClause{Name: name, Builder: q, Recursive: true})
	return &b2
}

// LeftLateralJoin adds a LEFT JOIN LATERAL clause (PostgreSQL and MySQL).
// The lateral subquery can reference columns from earlier FROM/JOIN items.
// The ON condition is automatically set to TRUE.
//
//	sqblpg.From(sqblpg.As("users", "u")).LeftLateralJoin(
//	    sqblpg.From("orders").Select("amount").Where(sqblpg.Eq("user_id", "u.id")).Limit(1),
//	    "latest",
//	)
//	// FROM "users" AS "u" LEFT LATERAL JOIN (SELECT ...) AS "latest" ON TRUE
func (b *SqlSelectBuilder) LeftLateralJoin(subquery *SqlSelectBuilder, alias string) *SqlSelectBuilder {
	b2 := *b
	b2.joins = append(append([]syntax.JoinClause(nil), b.joins...), syntax.JoinClause{
		Kind:    "LEFT",
		Table:   syntax.As(subquery, alias),
		Lateral: true,
	})
	return &b2
}

// CrossLateralJoin adds a CROSS JOIN LATERAL clause (PostgreSQL and MySQL).
// Equivalent to an INNER JOIN LATERAL (rows with no match are excluded).
//
//	sqblpg.From(sqblpg.As("users", "u")).CrossLateralJoin(
//	    sqblpg.From("orders").Select("amount").Where(sqblpg.Eq("user_id", "u.id")).Limit(1),
//	    "latest",
//	)
//	// FROM "users" AS "u" CROSS LATERAL JOIN (SELECT ...) AS "latest"
func (b *SqlSelectBuilder) CrossLateralJoin(subquery *SqlSelectBuilder, alias string) *SqlSelectBuilder {
	b2 := *b
	b2.joins = append(append([]syntax.JoinClause(nil), b.joins...), syntax.JoinClause{
		Kind:    "CROSS",
		Table:   syntax.As(subquery, alias),
		Lateral: true,
	})
	return &b2
}

// Union appends a UNION clause with another query.
// Duplicate rows are removed from the combined result set.
//
//	b.Union(sqblpg.From("archived_users").Select("id", "name"))
//	// SELECT ... UNION SELECT ...
func (b *SqlSelectBuilder) Union(q *SqlSelectBuilder) *SqlSelectBuilder {
	b2 := *b
	b2.setOps = append(append([]setOpClause(nil), b.setOps...), setOpClause{Op: "UNION", All: false, Builder: q})
	return &b2
}

// UnionAll appends a UNION ALL clause with another query.
// All rows including duplicates are included in the combined result set.
//
//	b.UnionAll(sqblpg.From("archived_users").Select("id", "name"))
//	// SELECT ... UNION ALL SELECT ...
func (b *SqlSelectBuilder) UnionAll(q *SqlSelectBuilder) *SqlSelectBuilder {
	b2 := *b
	b2.setOps = append(append([]setOpClause(nil), b.setOps...), setOpClause{Op: "UNION", All: true, Builder: q})
	return &b2
}

// RightJoin adds a RIGHT JOIN clause.
//
//	b.RightJoin(sqbl.As("orders", "o"), sqbl.Eq("u.id", "o.user_id"))
//	// RIGHT JOIN orders AS o ON u.id = o.user_id
func (b *SqlSelectBuilder) RightJoin(table any, on syntax.SqlFragment) *SqlSelectBuilder {
	b2 := *b
	b2.joins = append(append([]syntax.JoinClause(nil), b.joins...), syntax.JoinClause{
		Kind:  "RIGHT",
		Table: syntax.ToFragment(table),
		On:    on,
	})
	return &b2
}

// FullJoin adds a FULL OUTER JOIN clause.
//
//	b.FullJoin(sqbl.As("orders", "o"), sqbl.Eq("u.id", "o.user_id"))
//	// FULL OUTER JOIN orders AS o ON u.id = o.user_id
func (b *SqlSelectBuilder) FullJoin(table any, on syntax.SqlFragment) *SqlSelectBuilder {
	b2 := *b
	b2.joins = append(append([]syntax.JoinClause(nil), b.joins...), syntax.JoinClause{
		Kind:  "FULL OUTER",
		Table: syntax.ToFragment(table),
		On:    on,
	})
	return &b2
}

// Intersect appends an INTERSECT clause with another query.
// Returns only rows present in both result sets.
//
//	b.Intersect(sqblpg.From("premium_users").Select("id"))
//	// SELECT ... INTERSECT SELECT ...
func (b *SqlSelectBuilder) Intersect(q *SqlSelectBuilder) *SqlSelectBuilder {
	b2 := *b
	b2.setOps = append(append([]setOpClause(nil), b.setOps...), setOpClause{Op: "INTERSECT", All: false, Builder: q})
	return &b2
}

// IntersectAll appends an INTERSECT ALL clause with another query.
// Returns all rows present in both result sets, including duplicates.
//
//	b.IntersectAll(sqblpg.From("premium_users").Select("id"))
//	// SELECT ... INTERSECT ALL SELECT ...
func (b *SqlSelectBuilder) IntersectAll(q *SqlSelectBuilder) *SqlSelectBuilder {
	b2 := *b
	b2.setOps = append(append([]setOpClause(nil), b.setOps...), setOpClause{Op: "INTERSECT", All: true, Builder: q})
	return &b2
}

// Except appends an EXCEPT clause with another query.
// Returns rows from the first result set that are not in the second.
//
//	b.Except(sqblpg.From("banned_users").Select("id"))
//	// SELECT ... EXCEPT SELECT ...
func (b *SqlSelectBuilder) Except(q *SqlSelectBuilder) *SqlSelectBuilder {
	b2 := *b
	b2.setOps = append(append([]setOpClause(nil), b.setOps...), setOpClause{Op: "EXCEPT", All: false, Builder: q})
	return &b2
}

// ExceptAll appends an EXCEPT ALL clause with another query.
// Returns all rows from the first result set not in the second, including duplicates.
//
//	b.ExceptAll(sqblpg.From("banned_users").Select("id"))
//	// SELECT ... EXCEPT ALL SELECT ...
func (b *SqlSelectBuilder) ExceptAll(q *SqlSelectBuilder) *SqlSelectBuilder {
	b2 := *b
	b2.setOps = append(append([]setOpClause(nil), b.setOps...), setOpClause{Op: "EXCEPT", All: true, Builder: q})
	return &b2
}
