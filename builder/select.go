package builder

import (
	"bytes"
	"strconv"

	"github.com/kotofurumiya/sqbl/dialect"
	"github.com/kotofurumiya/sqbl/internal/sqlbuf"
	"github.com/kotofurumiya/sqbl/syntax"
)

// cteClause represents a single WITH ... AS (...) entry.
type cteClause struct {
	Name      string
	Builder   SqlSelectBuilder
	Recursive bool
}

// setOpClause represents a single set operation entry (UNION, INTERSECT, or EXCEPT).
// Op is one of "UNION", "INTERSECT", or "EXCEPT".
type setOpClause struct {
	Op      string
	All     bool
	Builder SqlSelectBuilder
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
	dialect    dialect.SqlDialect
	ctes       []cteClause          // WITH ...
	distinct   bool                 // DISTINCT
	distinctOn []string             // DISTINCT ON (col1, col2) — PostgreSQL only
	columns    []syntax.SqlFragment // SELECT ... (StringExpr | Aliased)
	from       syntax.SqlFragment   // FROM (StringExpr | SqlSelectBuilder | Aliased)
	joins      []syntax.JoinClause  // JOIN ...
	where      syntax.SqlFragment   // WHERE ...
	groups     []syntax.SqlFragment // GROUP BY ...
	having     syntax.SqlFragment   // HAVING ...
	orders     []syntax.SqlFragment // ORDER BY ...
	limit      *int                 // LIMIT
	offset     *int                 // OFFSET
	lock       string               // FOR UPDATE, FOR NO KEY UPDATE, FOR SHARE, FOR KEY SHARE
	lockOption string               // NOWAIT, SKIP LOCKED
	setOps     []setOpClause        // UNION, INTERSECT, EXCEPT (and ALL variants)
}

var _ syntax.SqlFragment = SqlSelectBuilder{}
var _ SqlBuilder = SqlSelectBuilder{}

// ToSql renders the query using the dialect set on this builder.
//
//	sql := sqblpg.From("users").Select("id").ToSql()
func (b SqlSelectBuilder) ToSql() string {
	buf := sqlbuf.GetStringBuffer()
	b.renderSQL(buf, b.dialect)
	buf.WriteByte(';')
	s := buf.String()
	sqlbuf.PutStringBuffer(buf)
	return s
}

// ToSqlWithDialect renders the statement using the given dialect, without a trailing semicolon.
// Implements SqlBuilder.
func (b SqlSelectBuilder) ToSqlWithDialect(d dialect.SqlDialect) string {
	buf := sqlbuf.GetStringBuffer()
	b.renderSQL(buf, d)
	s := buf.String()
	sqlbuf.PutStringBuffer(buf)
	return s
}

// AppendSQL implements syntax.SqlFragment for subquery embedding.
// Writes the query wrapped in parentheses directly into buf.
func (b SqlSelectBuilder) AppendSQL(buf *bytes.Buffer, d dialect.SqlDialect) {
	buf.WriteByte('(')
	b.renderSQL(buf, d)
	buf.WriteByte(')')
}

// renderSQL writes the query (without trailing semicolon) into buf.
// WriteString/Write calls append directly to the buffer, eliminating intermediate
// string values and the []string slice that strings.Join would require.
func (b SqlSelectBuilder) renderSQL(buf *bytes.Buffer, d dialect.SqlDialect) {
	// WITH [RECURSIVE] name AS (...), ...
	// Scan once to determine whether RECURSIVE is needed, then render each CTE.
	if len(b.ctes) > 0 {
		withKw := "WITH"
		for _, c := range b.ctes {
			if c.Recursive {
				withKw = "WITH RECURSIVE"
				break
			}
		}
		buf.WriteString(withKw)
		buf.WriteByte(' ')
		for i, c := range b.ctes {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(c.Name)
			buf.WriteString(" AS (")
			c.Builder.renderSQL(buf, d)
			buf.WriteByte(')')
		}
		buf.WriteByte(' ') // space before SELECT
	}

	// SELECT [DISTINCT [ON (col1, col2)]]
	// DISTINCT ON is PostgreSQL-only; fall back to plain SELECT on other dialects.
	if len(b.distinctOn) > 0 {
		if _, isPg := d.(*dialect.PostgresDialect); isPg {
			buf.WriteString("SELECT DISTINCT ON (")
			for i, col := range b.distinctOn {
				if i > 0 {
					buf.WriteString(", ")
				}
				d.QuoteIdentifier(buf, col)
			}
			buf.WriteByte(')')
		} else {
			buf.WriteString("SELECT")
		}
	} else if b.distinct {
		buf.WriteString("SELECT DISTINCT")
	} else {
		buf.WriteString("SELECT")
	}

	// Column list: col1, col2, ... — or * when none are specified.
	if len(b.columns) > 0 {
		buf.WriteByte(' ')
		for i, f := range b.columns {
			if i > 0 {
				buf.WriteString(", ")
			}
			f.AppendSQL(buf, d)
		}
	} else {
		buf.WriteString(" *")
	}

	// FROM table | subquery | aliased
	if b.from != nil {
		buf.WriteString(" FROM ")
		b.from.AppendSQL(buf, d)
	}

	// JOIN clauses — each rendered as "KIND [LATERAL] JOIN table [ON cond]".
	// CROSS JOIN has no ON clause; lateral joins use ON TRUE unless CROSS.
	for _, j := range b.joins {
		buf.WriteByte(' ')
		switch {
		case j.Lateral && j.Kind == "CROSS":
			buf.WriteString("CROSS LATERAL JOIN ")
			j.Table.AppendSQL(buf, d)
		case j.Lateral:
			buf.WriteString(j.Kind)
			buf.WriteString(" LATERAL JOIN ")
			j.Table.AppendSQL(buf, d)
			buf.WriteString(" ON TRUE")
		case j.Kind == "CROSS":
			buf.WriteString("CROSS JOIN ")
			j.Table.AppendSQL(buf, d)
		default:
			buf.WriteString(j.Kind)
			buf.WriteString(" JOIN ")
			j.Table.AppendSQL(buf, d)
			buf.WriteString(" ON ")
			j.On.AppendSQL(buf, d)
		}
	}

	// WHERE condition
	if b.where != nil {
		buf.WriteString(" WHERE ")
		b.where.AppendSQL(buf, d)
	}

	// GROUP BY col1, col2, ...
	if len(b.groups) > 0 {
		buf.WriteString(" GROUP BY ")
		for i, f := range b.groups {
			if i > 0 {
				buf.WriteString(", ")
			}
			f.AppendSQL(buf, d)
		}
	}

	// HAVING condition (requires GROUP BY in practice)
	if b.having != nil {
		buf.WriteString(" HAVING ")
		b.having.AppendSQL(buf, d)
	}

	// ORDER BY col1 ASC, col2 DESC, ...
	if len(b.orders) > 0 {
		buf.WriteString(" ORDER BY ")
		for i, f := range b.orders {
			if i > 0 {
				buf.WriteString(", ")
			}
			f.AppendSQL(buf, d)
		}
	}

	// LIMIT / OFFSET — AppendInt writes digits directly into buf without
	// allocating an intermediate string the way strconv.Itoa would.
	var tmp [20]byte
	if b.limit != nil {
		buf.WriteString(" LIMIT ")
		buf.Write(strconv.AppendInt(tmp[:0], int64(*b.limit), 10))
	}
	if b.offset != nil {
		buf.WriteString(" OFFSET ")
		buf.Write(strconv.AppendInt(tmp[:0], int64(*b.offset), 10))
	}

	// Locking clause: FOR UPDATE / FOR SHARE / FOR NO KEY UPDATE / FOR KEY SHARE
	// with optional NOWAIT or SKIP LOCKED modifier.
	if b.lock != "" {
		buf.WriteByte(' ')
		buf.WriteString(b.lock)
		if b.lockOption != "" {
			buf.WriteByte(' ')
			buf.WriteString(b.lockOption)
		}
	}

	// Set operations: UNION [ALL] / INTERSECT [ALL] / EXCEPT [ALL]
	for _, u := range b.setOps {
		buf.WriteByte(' ')
		buf.WriteString(u.Op)
		if u.All {
			buf.WriteString(" ALL")
		}
		buf.WriteByte(' ')
		u.Builder.renderSQL(buf, d)
	}
}

// Dialect sets the SQL dialect used when rendering the query.
// Dialect-specific constructors (e.g. sqblpg.From) call this automatically.
//
//	b.Dialect(pgDialect)
func (b SqlSelectBuilder) Dialect(d dialect.SqlDialect) SqlSelectBuilder {
	b.dialect = d
	return b
}

// From sets the FROM clause.
// Accepts a table name string, an aliased expression, or a subquery.
//
//	sqblpg.From("users")
//	sqblpg.From(sqbl.As("users", "u"))
//	sqblpg.From(sqbl.As(subquery, "sub"))
func (b SqlSelectBuilder) From(table any) SqlSelectBuilder {
	b.from = syntax.ToFragment(table)
	return b
}

// Select sets the columns to retrieve.
// Accepts column name strings, aliased expressions, or raw SQL fragments.
// Calling Select again replaces the previous column list.
//
//	b.Select("id", "name")
//	b.Select(sqbl.As("SUM(amount)", "total"))
func (b SqlSelectBuilder) Select(columns ...any) SqlSelectBuilder {
	cols := make([]syntax.SqlFragment, len(columns))
	for i, col := range columns {
		cols[i] = syntax.ToFragment(col)
	}
	b.columns = cols
	return b
}

// Join adds an INNER JOIN clause. It is an alias for InnerJoin.
//
//	b.Join(sqbl.As("orders", "o"), "u.id = o.user_id")
//	// INNER JOIN orders AS o ON u.id = o.user_id
func (b SqlSelectBuilder) Join(table any, on syntax.SqlFragment) SqlSelectBuilder {
	return b.InnerJoin(table, on)
}

// Where sets the WHERE condition.
// Use dialect-specific helpers (e.g. sqblpg.Eq, sqblpg.Gt) to build conditions.
// Combine multiple conditions explicitly with And or Or.
//
//	b.Where(sqblpg.Eq("status", "active"))
//	b.Where(sqblpg.And(sqblpg.Eq("status", "active"), sqblpg.Gt("age", 18)))
func (b SqlSelectBuilder) Where(expr syntax.SqlFragment) SqlSelectBuilder {
	b.where = expr
	return b
}

// Distinct adds the DISTINCT keyword to the SELECT clause.
//
//	sqblpg.From("users").Select("country").Distinct()
//	// SELECT DISTINCT country FROM users
func (b SqlSelectBuilder) Distinct() SqlSelectBuilder {
	b.distinct = true
	return b
}

// DistinctOn sets the DISTINCT ON columns (PostgreSQL only).
// Rows are deduplicated based on the specified columns; the first row for each group is kept.
// ORDER BY should be used to control which row is considered "first".
//
//	sqblpg.From("orders").Select("user_id", "amount").DistinctOn("user_id").OrderBy(sqblpg.Desc("amount"))
//	// SELECT DISTINCT ON ("user_id") "user_id", "amount" FROM "orders" ORDER BY "amount" DESC
func (b SqlSelectBuilder) DistinctOn(cols ...string) SqlSelectBuilder {
	b.distinctOn = cols
	return b
}

// GroupBy sets the GROUP BY columns or expressions.
// Accepts column name strings, aliased expressions, or raw SQL fragments.
//
//	b.GroupBy("department", "role")
//	b.GroupBy(sqbl.As("EXTRACT(year FROM created_at)", "year"))
func (b SqlSelectBuilder) GroupBy(cols ...any) SqlSelectBuilder {
	groups := make([]syntax.SqlFragment, len(cols))
	for i, col := range cols {
		groups[i] = syntax.ToFragment(col)
	}
	b.groups = groups
	return b
}

// Having sets the HAVING condition, applied after grouping.
// Use dialect-specific helpers to build conditions.
// Combine multiple conditions explicitly with And or Or.
//
//	b.Having(sqblpg.Gt("SUM(amount)", 1000))
//	// HAVING SUM(amount) > 1000
func (b SqlSelectBuilder) Having(expr syntax.SqlFragment) SqlSelectBuilder {
	b.having = expr
	return b
}

// OrderBy sets the ORDER BY columns.
// Accepts column name strings or Order expressions (syntax.Asc / syntax.Desc).
//
//	b.OrderBy(syntax.Asc("name"), syntax.Desc("created_at"))
//	// ORDER BY name ASC, created_at DESC
func (b SqlSelectBuilder) OrderBy(cols ...any) SqlSelectBuilder {
	orders := make([]syntax.SqlFragment, len(cols))
	for i, col := range cols {
		orders[i] = syntax.ToFragment(col)
	}
	b.orders = orders
	return b
}

// Limit sets the maximum number of rows to return.
//
//	b.Limit(20)
//	// LIMIT 20
func (b SqlSelectBuilder) Limit(n int) SqlSelectBuilder {
	b.limit = &n
	return b
}

// Offset sets the number of rows to skip before returning results.
//
//	b.Limit(20).Offset(40)
//	// LIMIT 20 OFFSET 40
func (b SqlSelectBuilder) Offset(n int) SqlSelectBuilder {
	b.offset = &n
	return b
}

// ForUpdate appends a FOR UPDATE locking clause to the query.
// Locks selected rows for the duration of the current transaction.
//
//	b.ForUpdate()
//	// SELECT ... FOR UPDATE
func (b SqlSelectBuilder) ForUpdate() SqlSelectBuilder {
	b.lock = "FOR UPDATE"
	return b
}

// ForShare appends a FOR SHARE locking clause to the query.
// Allows concurrent reads but prevents modification of selected rows.
//
//	b.ForShare()
//	// SELECT ... FOR SHARE
func (b SqlSelectBuilder) ForShare() SqlSelectBuilder {
	b.lock = "FOR SHARE"
	return b
}

// ForNoKeyUpdate appends a FOR NO KEY UPDATE locking clause.
// Weaker than FOR UPDATE; does not block INSERT of rows referencing this table.
//
//	b.ForNoKeyUpdate()
//	// SELECT ... FOR NO KEY UPDATE
func (b SqlSelectBuilder) ForNoKeyUpdate() SqlSelectBuilder {
	b.lock = "FOR NO KEY UPDATE"
	return b
}

// ForKeyShare appends a FOR KEY SHARE locking clause.
// Weakest lock; only blocks DELETE and UPDATE of key columns.
//
//	b.ForKeyShare()
//	// SELECT ... FOR KEY SHARE
func (b SqlSelectBuilder) ForKeyShare() SqlSelectBuilder {
	b.lock = "FOR KEY SHARE"
	return b
}

// Nowait adds the NOWAIT option to the locking clause.
// The query fails immediately if any selected row cannot be locked.
//
//	b.ForUpdate().Nowait()
//	// SELECT ... FOR UPDATE NOWAIT
func (b SqlSelectBuilder) Nowait() SqlSelectBuilder {
	b.lockOption = "NOWAIT"
	return b
}

// SkipLocked adds the SKIP LOCKED option to the locking clause.
// Rows that cannot be immediately locked are silently skipped.
//
//	b.ForUpdate().SkipLocked()
//	// SELECT ... FOR UPDATE SKIP LOCKED
func (b SqlSelectBuilder) SkipLocked() SqlSelectBuilder {
	b.lockOption = "SKIP LOCKED"
	return b
}

// InnerJoin adds an INNER JOIN clause.
//
//	b.InnerJoin(sqbl.As("orders", "o"), "u.id = o.user_id")
//	// INNER JOIN orders AS o ON u.id = o.user_id
func (b SqlSelectBuilder) InnerJoin(table any, on syntax.SqlFragment) SqlSelectBuilder {
	b.joins = append(append([]syntax.JoinClause(nil), b.joins...), syntax.JoinClause{
		Kind:  "INNER",
		Table: syntax.ToFragment(table),
		On:    on,
	})
	return b
}

// LeftJoin adds a LEFT JOIN clause.
//
//	b.LeftJoin(sqbl.As("profiles", "p"), sqbl.Eq("u.id", "p.user_id"))
//	// LEFT JOIN profiles AS p ON u.id = p.user_id
func (b SqlSelectBuilder) LeftJoin(table any, on syntax.SqlFragment) SqlSelectBuilder {
	b.joins = append(append([]syntax.JoinClause(nil), b.joins...), syntax.JoinClause{
		Kind:  "LEFT",
		Table: syntax.ToFragment(table),
		On:    on,
	})
	return b
}

// CrossJoin adds a CROSS JOIN clause.
// No ON condition is required for a cross join.
//
//	b.CrossJoin("sizes")
//	// CROSS JOIN sizes
func (b SqlSelectBuilder) CrossJoin(table any) SqlSelectBuilder {
	b.joins = append(append([]syntax.JoinClause(nil), b.joins...), syntax.JoinClause{
		Kind:  "CROSS",
		Table: syntax.ToFragment(table),
		On:    nil,
	})
	return b
}

// With adds a Common Table Expression (CTE) to the query.
// The CTE is rendered as a WITH clause before the main SELECT.
//
//	recent := sqblpg.From("orders").Where(sqblpg.Gte("created_at", "2025-01-01"))
//	b.With("recent_orders", recent)
//	// WITH recent_orders AS (SELECT ...) SELECT ...
func (b SqlSelectBuilder) With(name string, q SqlSelectBuilder) SqlSelectBuilder {
	b.ctes = append(append([]cteClause(nil), b.ctes...), cteClause{Name: name, Builder: q})
	return b
}

// WithRecursive adds a recursive CTE to the query.
// The WITH clause is rendered as WITH RECURSIVE when this is used.
// Use this for hierarchical or iterative queries.
//
//	base := sqblpg.From("employees").Select("id", "manager_id").Where(sqblpg.IsNull("manager_id"))
//	recursive := sqblpg.From("employees").Join(sqblpg.As("org", "o"), sqblpg.Eq("employees.manager_id", "o.id")).Select("employees.id", "employees.manager_id")
//	b.WithRecursive("org", base.Union(recursive))
//	// WITH RECURSIVE org AS (...) SELECT ...
func (b SqlSelectBuilder) WithRecursive(name string, q SqlSelectBuilder) SqlSelectBuilder {
	b.ctes = append(append([]cteClause(nil), b.ctes...), cteClause{Name: name, Builder: q, Recursive: true})
	return b
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
func (b SqlSelectBuilder) LeftLateralJoin(subquery SqlSelectBuilder, alias string) SqlSelectBuilder {
	b.joins = append(append([]syntax.JoinClause(nil), b.joins...), syntax.JoinClause{
		Kind:    "LEFT",
		Table:   syntax.As(subquery, alias),
		Lateral: true,
	})
	return b
}

// CrossLateralJoin adds a CROSS JOIN LATERAL clause (PostgreSQL and MySQL).
// Equivalent to an INNER JOIN LATERAL (rows with no match are excluded).
//
//	sqblpg.From(sqblpg.As("users", "u")).CrossLateralJoin(
//	    sqblpg.From("orders").Select("amount").Where(sqblpg.Eq("user_id", "u.id")).Limit(1),
//	    "latest",
//	)
//	// FROM "users" AS "u" CROSS LATERAL JOIN (SELECT ...) AS "latest"
func (b SqlSelectBuilder) CrossLateralJoin(subquery SqlSelectBuilder, alias string) SqlSelectBuilder {
	b.joins = append(append([]syntax.JoinClause(nil), b.joins...), syntax.JoinClause{
		Kind:    "CROSS",
		Table:   syntax.As(subquery, alias),
		Lateral: true,
	})
	return b
}

// Union appends a UNION clause with another query.
// Duplicate rows are removed from the combined result set.
//
//	b.Union(sqblpg.From("archived_users").Select("id", "name"))
//	// SELECT ... UNION SELECT ...
func (b SqlSelectBuilder) Union(q SqlSelectBuilder) SqlSelectBuilder {
	b.setOps = append(append([]setOpClause(nil), b.setOps...), setOpClause{Op: "UNION", All: false, Builder: q})
	return b
}

// UnionAll appends a UNION ALL clause with another query.
// All rows including duplicates are included in the combined result set.
//
//	b.UnionAll(sqblpg.From("archived_users").Select("id", "name"))
//	// SELECT ... UNION ALL SELECT ...
func (b SqlSelectBuilder) UnionAll(q SqlSelectBuilder) SqlSelectBuilder {
	b.setOps = append(append([]setOpClause(nil), b.setOps...), setOpClause{Op: "UNION", All: true, Builder: q})
	return b
}

// RightJoin adds a RIGHT JOIN clause.
//
//	b.RightJoin(sqbl.As("orders", "o"), sqbl.Eq("u.id", "o.user_id"))
//	// RIGHT JOIN orders AS o ON u.id = o.user_id
func (b SqlSelectBuilder) RightJoin(table any, on syntax.SqlFragment) SqlSelectBuilder {
	b.joins = append(append([]syntax.JoinClause(nil), b.joins...), syntax.JoinClause{
		Kind:  "RIGHT",
		Table: syntax.ToFragment(table),
		On:    on,
	})
	return b
}

// FullJoin adds a FULL OUTER JOIN clause.
//
//	b.FullJoin(sqbl.As("orders", "o"), sqbl.Eq("u.id", "o.user_id"))
//	// FULL OUTER JOIN orders AS o ON u.id = o.user_id
func (b SqlSelectBuilder) FullJoin(table any, on syntax.SqlFragment) SqlSelectBuilder {
	b.joins = append(append([]syntax.JoinClause(nil), b.joins...), syntax.JoinClause{
		Kind:  "FULL OUTER",
		Table: syntax.ToFragment(table),
		On:    on,
	})
	return b
}

// Intersect appends an INTERSECT clause with another query.
// Returns only rows present in both result sets.
//
//	b.Intersect(sqblpg.From("premium_users").Select("id"))
//	// SELECT ... INTERSECT SELECT ...
func (b SqlSelectBuilder) Intersect(q SqlSelectBuilder) SqlSelectBuilder {
	b.setOps = append(append([]setOpClause(nil), b.setOps...), setOpClause{Op: "INTERSECT", All: false, Builder: q})
	return b
}

// IntersectAll appends an INTERSECT ALL clause with another query.
// Returns all rows present in both result sets, including duplicates.
//
//	b.IntersectAll(sqblpg.From("premium_users").Select("id"))
//	// SELECT ... INTERSECT ALL SELECT ...
func (b SqlSelectBuilder) IntersectAll(q SqlSelectBuilder) SqlSelectBuilder {
	b.setOps = append(append([]setOpClause(nil), b.setOps...), setOpClause{Op: "INTERSECT", All: true, Builder: q})
	return b
}

// Except appends an EXCEPT clause with another query.
// Returns rows from the first result set that are not in the second.
//
//	b.Except(sqblpg.From("banned_users").Select("id"))
//	// SELECT ... EXCEPT SELECT ...
func (b SqlSelectBuilder) Except(q SqlSelectBuilder) SqlSelectBuilder {
	b.setOps = append(append([]setOpClause(nil), b.setOps...), setOpClause{Op: "EXCEPT", All: false, Builder: q})
	return b
}

// ExceptAll appends an EXCEPT ALL clause with another query.
// Returns all rows from the first result set not in the second, including duplicates.
//
//	b.ExceptAll(sqblpg.From("banned_users").Select("id"))
//	// SELECT ... EXCEPT ALL SELECT ...
func (b SqlSelectBuilder) ExceptAll(q SqlSelectBuilder) SqlSelectBuilder {
	b.setOps = append(append([]setOpClause(nil), b.setOps...), setOpClause{Op: "EXCEPT", All: true, Builder: q})
	return b
}
