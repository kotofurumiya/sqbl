package builder

import (
	"bytes"
	"slices"
	"strconv"

	"github.com/kotofurumiya/sqbl/dialect"
	"github.com/kotofurumiya/sqbl/syntax"
)

// Query is a SELECT statement, alone or combined with set operations.
// It is what CTEs, set operations and lateral joins accept.
// It is implemented by SelectBuilder and CompoundBuilder only.
type Query interface {
	Statement
	syntax.Subquery
	renderer

	// dialectOf returns the dialect the query was built with.
	dialectOf() dialect.Dialect
	// writeOperand writes the query as an operand of a set operation,
	// parenthesized when it would otherwise change the meaning of the compound.
	// right reports whether the query is to the right of the operator.
	writeOperand(buf *bytes.Buffer, d dialect.Dialect, right bool)
}

// cteClause represents a single WITH ... AS (...) entry.
type cteClause struct {
	name      string
	query     Query
	recursive bool
}

// joinClause represents a single JOIN entry.
// kind is one of "INNER", "LEFT", "RIGHT", "FULL OUTER", or "CROSS".
type joinClause struct {
	kind    string
	table   syntax.SQLFragment
	on      syntax.SQLFragment
	lateral bool
}

// SelectBuilder constructs a SQL SELECT statement using a fluent method chain.
// Use a dialect-specific constructor (e.g. sqblpg.From) to create an instance.
//
//	q := sqblpg.From("users").
//	    Select("id", "name").
//	    Where(sqblpg.Eq("active", true)).
//	    OrderBy(sqblpg.Asc("name")).
//	    Limit(20)
//	sql, err := q.ToSQL()
//	// SELECT "id", "name" FROM "users" WHERE "active" = TRUE ORDER BY "name" ASC LIMIT 20;
type SelectBuilder struct {
	dialect    dialect.Dialect
	ctes       []cteClause          // WITH ...
	distinct   bool                 // DISTINCT
	distinctOn []any                // DISTINCT ON (...) — PostgreSQL only
	columns    []syntax.SQLFragment // SELECT ...
	from       syntax.SQLFragment   // FROM ...
	joins      []joinClause         // JOIN ...
	where      []syntax.SQLFragment // WHERE ..., ANDed
	groups     []syntax.SQLFragment // GROUP BY ...
	having     []syntax.SQLFragment // HAVING ..., ANDed
	orders     []syntax.SQLFragment // ORDER BY ...
	limit      *int                 // LIMIT
	offset     *int                 // OFFSET
	lock       string               // FOR UPDATE, FOR NO KEY UPDATE, FOR SHARE, FOR KEY SHARE
	lockOption string               // NOWAIT, SKIP LOCKED
}

var _ Query = SelectBuilder{}

// ToSQL renders the statement with the builder's dialect and a trailing semicolon.
func (b SelectBuilder) ToSQL() (string, error) { return render(b, b.dialect, true) }

// ToSQLWithDialect renders the statement with d and no trailing semicolon.
func (b SelectBuilder) ToSQLWithDialect(d dialect.Dialect) (string, error) {
	return render(b, d, false)
}

// MustSQL is like ToSQL but panics on error.
func (b SelectBuilder) MustSQL() string { return must(b.ToSQL()) }

// MustSQLWithDialect is like ToSQLWithDialect but panics on error.
func (b SelectBuilder) MustSQLWithDialect(d dialect.Dialect) string {
	return must(b.ToSQLWithDialect(d))
}

// AppendSQL implements syntax.SQLFragment, writing the query in parentheses
// so that it can be embedded as a subquery.
func (b SelectBuilder) AppendSQL(buf *bytes.Buffer, d dialect.Dialect) {
	buf.WriteByte('(')
	b.renderSQL(buf, d)
	buf.WriteByte(')')
}

// IsSubquery implements syntax.Subquery, so IN (subquery) renders single parentheses.
func (b SelectBuilder) IsSubquery() {}

func (b SelectBuilder) dialectOf() dialect.Dialect { return b.dialect }

func (b SelectBuilder) writeOperand(buf *bytes.Buffer, d dialect.Dialect, right bool) {
	// Trailing clauses would otherwise bind to the whole compound, and a WITH on
	// the right is not allowed at all without parentheses. A WITH on the leftmost
	// operand is left bare: it already scopes over the whole compound.
	if len(b.orders) > 0 || b.limit != nil || b.offset != nil || b.lock != "" || (right && len(b.ctes) > 0) {
		b.AppendSQL(buf, d)
		return
	}
	b.renderSQL(buf, d)
}

func (b SelectBuilder) renderSQL(buf *bytes.Buffer, d dialect.Dialect) {
	writeCTEs(buf, d, b.ctes)

	switch {
	case len(b.distinctOn) > 0:
		buf.WriteString("SELECT DISTINCT ON (")
		writeValues(buf, d, b.distinctOn)
		buf.WriteByte(')')
	case b.distinct:
		buf.WriteString("SELECT DISTINCT")
	default:
		buf.WriteString("SELECT")
	}

	if len(b.columns) > 0 {
		buf.WriteByte(' ')
		writeFragments(buf, d, b.columns)
	} else {
		buf.WriteString(" *")
	}

	if b.from != nil {
		buf.WriteString(" FROM ")
		b.from.AppendSQL(buf, d)
	}

	for _, j := range b.joins {
		buf.WriteByte(' ')
		buf.WriteString(j.kind)
		buf.WriteString(" JOIN ")
		if j.lateral {
			buf.WriteString("LATERAL ")
		}
		j.table.AppendSQL(buf, d)
		switch {
		case j.kind == "CROSS":
		case j.lateral:
			buf.WriteString(" ON TRUE")
		default:
			buf.WriteString(" ON ")
			j.on.AppendSQL(buf, d)
		}
	}

	writeConditions(buf, d, " WHERE ", b.where)

	if len(b.groups) > 0 {
		buf.WriteString(" GROUP BY ")
		writeFragments(buf, d, b.groups)
	}

	writeConditions(buf, d, " HAVING ", b.having)
	writeOrderLimitOffset(buf, d, b.orders, b.limit, b.offset)

	if b.lock == "" && b.lockOption != "" {
		method := map[string]string{"NOWAIT": "Nowait", "SKIP LOCKED": "SkipLocked"}[b.lockOption]
		failFragment(method+" needs a locking clause; add ForUpdate, ForShare, ForNoKeyUpdate or ForKeyShare").
			AppendSQL(buf, d)
	}
	if b.lock != "" {
		buf.WriteByte(' ')
		buf.WriteString(b.lock)
		if b.lockOption != "" {
			buf.WriteByte(' ')
			buf.WriteString(b.lockOption)
		}
	}
}

// writeCTEs writes the WITH clause followed by a space, or nothing when ctes is empty.
// One recursive CTE makes the whole clause WITH RECURSIVE, as SQL requires.
func writeCTEs(buf *bytes.Buffer, d dialect.Dialect, ctes []cteClause) {
	if len(ctes) == 0 {
		return
	}
	buf.WriteString("WITH ")
	if slices.ContainsFunc(ctes, func(c cteClause) bool { return c.recursive }) {
		buf.WriteString("RECURSIVE ")
	}
	for i, c := range ctes {
		if i > 0 {
			buf.WriteString(", ")
		}
		d.Quote(buf, c.name)
		buf.WriteString(" AS (")
		c.query.renderSQL(buf, d)
		buf.WriteByte(')')
	}
	buf.WriteByte(' ')
}

// writeFragments writes frags separated by ", ".
func writeFragments(buf *bytes.Buffer, d dialect.Dialect, frags []syntax.SQLFragment) {
	for i, f := range frags {
		if i > 0 {
			buf.WriteString(", ")
		}
		f.AppendSQL(buf, d)
	}
}

// writeOrderLimitOffset writes the ORDER BY, LIMIT and OFFSET clauses that are set.
func writeOrderLimitOffset(buf *bytes.Buffer, d dialect.Dialect, orders []syntax.SQLFragment, limit, offset *int) {
	if len(orders) > 0 {
		buf.WriteString(" ORDER BY ")
		writeFragments(buf, d, orders)
	}
	if limit != nil {
		buf.WriteString(" LIMIT ")
		buf.WriteString(strconv.Itoa(*limit))
	}
	if offset != nil {
		buf.WriteString(" OFFSET ")
		buf.WriteString(strconv.Itoa(*offset))
	}
}

// toFragments converts each value with syntax.ToFragment.
func toFragments(vals []any) []syntax.SQLFragment {
	frags := make([]syntax.SQLFragment, len(vals))
	for i, v := range vals {
		frags[i] = syntax.ToFragment(v)
	}
	return frags
}

// Dialect sets the SQL dialect used when rendering the query.
// Dialect-specific constructors (e.g. sqblpg.From) call this automatically.
func (b SelectBuilder) Dialect(d dialect.Dialect) SelectBuilder {
	b.dialect = d
	return b
}

// From sets the FROM clause, replacing any earlier one.
// It accepts a table name, an aliased source, or a subquery; nil fails the render.
//
//	sqblpg.From("users")
//	sqblpg.From(sqblpg.As("users", "u"))
//	sqblpg.From(sqblpg.As(subquery, "sub"))
func (b SelectBuilder) From(table any) SelectBuilder {
	b.from = source("From", table)
	return b
}

// Select sets the columns to retrieve, replacing any earlier list.
// Each column follows the rules of syntax.AppendValue; a plain string is an identifier.
// With no columns, the query selects *.
//
//	b.Select("id", "u.name")
//	b.Select(sqblpg.As(sqblpg.Fn("SUM", "amount"), "total"))
//	// SELECT SUM("amount") AS "total"
func (b SelectBuilder) Select(columns ...any) SelectBuilder {
	b.columns = toFragments(columns)
	return b
}

// Where adds a WHERE condition.
// Calling Where again ANDs the new condition with the earlier ones, so a shared
// base query can only be narrowed. A nil condition is ignored.
//
//	b.Where(sqblpg.Eq("status", sqblpg.P(1))).Where(sqblpg.Gt("age", 18))
//	// WHERE "status" = $1 AND "age" > 18
func (b SelectBuilder) Where(expr syntax.SQLFragment) SelectBuilder {
	b.where = addCondition(b.where, expr)
	return b
}

// Distinct adds the DISTINCT keyword to the SELECT clause.
// DistinctOn takes precedence when both are set.
//
//	sqblpg.From("users").Select("country").Distinct()
//	// SELECT DISTINCT "country" FROM "users"
func (b SelectBuilder) Distinct() SelectBuilder {
	b.distinct = true
	return b
}

// DistinctOn sets the DISTINCT ON expressions, replacing any earlier list. Supported by PostgreSQL.
// Each item follows the rules of syntax.AppendValue; a plain string is an identifier.
// Rows are deduplicated on these expressions, keeping the first row of each group
// as ordered by ORDER BY. It takes precedence over Distinct.
//
//	sqblpg.From("orders").Select("user_id", "amount").DistinctOn("user_id").OrderBy("user_id", sqblpg.Desc("amount"))
//	// SELECT DISTINCT ON ("user_id") "user_id", "amount" FROM "orders" ORDER BY "user_id", "amount" DESC
func (b SelectBuilder) DistinctOn(cols ...any) SelectBuilder {
	b.distinctOn = slices.Clone(cols)
	return b
}

// GroupBy sets the GROUP BY items, replacing any earlier list.
// Each item follows the rules of syntax.AppendValue; a plain string is an identifier.
//
//	b.GroupBy("department", "role")
//	b.GroupBy(sqblpg.Raw(`EXTRACT(year FROM "created_at")`))
func (b SelectBuilder) GroupBy(cols ...any) SelectBuilder {
	b.groups = toFragments(cols)
	return b
}

// Having adds a HAVING condition, applied after grouping.
// Calling Having again ANDs the new condition with the earlier ones. A nil condition is ignored.
//
//	b.Having(sqblpg.Gt(sqblpg.Fn("SUM", "amount"), 1000))
//	// HAVING SUM("amount") > 1000
func (b SelectBuilder) Having(expr syntax.SQLFragment) SelectBuilder {
	b.having = addCondition(b.having, expr)
	return b
}

// OrderBy sets the ORDER BY items, replacing any earlier list.
// Each item follows the rules of syntax.AppendValue: pass syntax.Asc or syntax.Desc
// for a direction, since a plain string such as "name DESC" is quoted as one identifier.
//
//	b.OrderBy(sqblpg.Asc("name"), sqblpg.Desc("created_at"))
//	// ORDER BY "name" ASC, "created_at" DESC
func (b SelectBuilder) OrderBy(cols ...any) SelectBuilder {
	b.orders = toFragments(cols)
	return b
}

// Limit sets the maximum number of rows to return, replacing any earlier limit.
//
//	b.Limit(20)
//	// LIMIT 20
func (b SelectBuilder) Limit(n int) SelectBuilder {
	b.limit = &n
	return b
}

// Offset sets the number of rows to skip, replacing any earlier offset.
// OFFSET without LIMIT is accepted by PostgreSQL only; MySQL and SQLite need a LIMIT as well.
//
//	b.Limit(20).Offset(40)
//	// LIMIT 20 OFFSET 40
func (b SelectBuilder) Offset(n int) SelectBuilder {
	b.offset = &n
	return b
}

// ForUpdate sets the locking clause to FOR UPDATE, replacing any earlier lock.
// Selected rows are locked for the rest of the transaction.
//
//	b.ForUpdate()
//	// SELECT ... FOR UPDATE
func (b SelectBuilder) ForUpdate() SelectBuilder {
	b.lock = "FOR UPDATE"
	return b
}

// ForShare sets the locking clause to FOR SHARE, replacing any earlier lock.
// Other transactions can still read the rows but cannot modify them.
// Supported by PostgreSQL and MySQL 8.0+; MariaDB uses LOCK IN SHARE MODE instead.
//
//	b.ForShare()
//	// SELECT ... FOR SHARE
func (b SelectBuilder) ForShare() SelectBuilder {
	b.lock = "FOR SHARE"
	return b
}

// ForNoKeyUpdate sets the locking clause to FOR NO KEY UPDATE, replacing any earlier lock.
// Weaker than FOR UPDATE: it does not block inserts that reference the rows. PostgreSQL only.
//
//	b.ForNoKeyUpdate()
//	// SELECT ... FOR NO KEY UPDATE
func (b SelectBuilder) ForNoKeyUpdate() SelectBuilder {
	b.lock = "FOR NO KEY UPDATE"
	return b
}

// ForKeyShare sets the locking clause to FOR KEY SHARE, replacing any earlier lock.
// The weakest lock: it only blocks deletes and key updates. PostgreSQL only.
//
//	b.ForKeyShare()
//	// SELECT ... FOR KEY SHARE
func (b SelectBuilder) ForKeyShare() SelectBuilder {
	b.lock = "FOR KEY SHARE"
	return b
}

// Nowait makes the locking clause fail immediately instead of waiting for locked rows.
// It replaces SkipLocked, and rendering fails unless a locking clause such as ForUpdate is set.
//
//	b.ForUpdate().Nowait()
//	// SELECT ... FOR UPDATE NOWAIT
func (b SelectBuilder) Nowait() SelectBuilder {
	b.lockOption = "NOWAIT"
	return b
}

// SkipLocked makes the locking clause skip rows that are already locked.
// It replaces Nowait, and rendering fails unless a locking clause such as ForUpdate is set.
//
//	b.ForUpdate().SkipLocked()
//	// SELECT ... FOR UPDATE SKIP LOCKED
func (b SelectBuilder) SkipLocked() SelectBuilder {
	b.lockOption = "SKIP LOCKED"
	return b
}

// addJoin returns b with j appended to its joins.
func (b SelectBuilder) addJoin(j joinClause) SelectBuilder {
	b.joins = append(slices.Clone(b.joins), j)
	return b
}

// joinWithOn builds a join that requires an ON condition; a nil condition fails the render.
func joinWithOn(method, kind string, table any, on syntax.SQLFragment) joinClause {
	if on == nil {
		on = failFragment(method + " needs an ON condition; use CrossJoin to join without one")
	}
	return joinClause{kind: kind, table: source(method, table), on: on}
}

// Join adds an INNER JOIN clause. It is an alias for InnerJoin.
//
//	b.Join(sqblpg.As("orders", "o"), sqblpg.Eq("u.id", "o.user_id"))
//	// INNER JOIN "orders" AS "o" ON "u"."id" = "o"."user_id"
func (b SelectBuilder) Join(table any, on syntax.SQLFragment) SelectBuilder {
	return b.addJoin(joinWithOn("Join", "INNER", table, on))
}

// InnerJoin adds an INNER JOIN clause after any earlier joins.
// on is required; rendering fails when it is nil.
//
//	b.InnerJoin(sqblpg.As("orders", "o"), sqblpg.Eq("u.id", "o.user_id"))
//	// INNER JOIN "orders" AS "o" ON "u"."id" = "o"."user_id"
func (b SelectBuilder) InnerJoin(table any, on syntax.SQLFragment) SelectBuilder {
	return b.addJoin(joinWithOn("InnerJoin", "INNER", table, on))
}

// LeftJoin adds a LEFT JOIN clause after any earlier joins.
// on is required; rendering fails when it is nil.
//
//	b.LeftJoin(sqblpg.As("profiles", "p"), sqblpg.Eq("u.id", "p.user_id"))
//	// LEFT JOIN "profiles" AS "p" ON "u"."id" = "p"."user_id"
func (b SelectBuilder) LeftJoin(table any, on syntax.SQLFragment) SelectBuilder {
	return b.addJoin(joinWithOn("LeftJoin", "LEFT", table, on))
}

// RightJoin adds a RIGHT JOIN clause after any earlier joins.
// on is required; rendering fails when it is nil.
//
//	b.RightJoin(sqblpg.As("orders", "o"), sqblpg.Eq("u.id", "o.user_id"))
//	// RIGHT JOIN "orders" AS "o" ON "u"."id" = "o"."user_id"
func (b SelectBuilder) RightJoin(table any, on syntax.SQLFragment) SelectBuilder {
	return b.addJoin(joinWithOn("RightJoin", "RIGHT", table, on))
}

// FullJoin adds a FULL OUTER JOIN clause after any earlier joins.
// on is required; rendering fails when it is nil. Not supported by MySQL.
//
//	b.FullJoin(sqblpg.As("orders", "o"), sqblpg.Eq("u.id", "o.user_id"))
//	// FULL OUTER JOIN "orders" AS "o" ON "u"."id" = "o"."user_id"
func (b SelectBuilder) FullJoin(table any, on syntax.SQLFragment) SelectBuilder {
	return b.addJoin(joinWithOn("FullJoin", "FULL OUTER", table, on))
}

// CrossJoin adds a CROSS JOIN clause after any earlier joins. It has no ON condition.
//
//	b.CrossJoin("sizes")
//	// CROSS JOIN "sizes"
func (b SelectBuilder) CrossJoin(table any) SelectBuilder {
	return b.addJoin(joinClause{kind: "CROSS", table: source("CrossJoin", table)})
}

// LeftLateralJoin adds a LEFT JOIN LATERAL clause after any earlier joins.
// The subquery can reference columns of earlier FROM and JOIN items, and the ON
// condition is always TRUE. Supported by PostgreSQL and MySQL 8.0.14+, not by MariaDB or SQLite.
//
//	sqblpg.From(sqblpg.As("users", "u")).LeftLateralJoin(
//	    sqblpg.From("orders").Select("amount").Where(sqblpg.Eq("user_id", "u.id")).Limit(1),
//	    "latest",
//	)
//	// SELECT * FROM "users" AS "u" LEFT JOIN LATERAL
//	//   (SELECT "amount" FROM "orders" WHERE "user_id" = "u"."id" LIMIT 1) AS "latest" ON TRUE
func (b SelectBuilder) LeftLateralJoin(subquery Query, alias string) SelectBuilder {
	return b.addJoin(joinClause{kind: "LEFT", table: syntax.As(subquery, alias), lateral: true})
}

// CrossLateralJoin adds a CROSS JOIN LATERAL clause after any earlier joins.
// Rows for which the subquery returns nothing are dropped, as with an inner join.
// Supported by PostgreSQL and MySQL 8.0.14+, not by MariaDB or SQLite.
//
//	sqblpg.From(sqblpg.As("users", "u")).CrossLateralJoin(
//	    sqblpg.From("orders").Select("amount").Where(sqblpg.Eq("user_id", "u.id")).Limit(1),
//	    "latest",
//	)
//	// SELECT * FROM "users" AS "u" CROSS JOIN LATERAL
//	//   (SELECT "amount" FROM "orders" WHERE "user_id" = "u"."id" LIMIT 1) AS "latest"
func (b SelectBuilder) CrossLateralJoin(subquery Query, alias string) SelectBuilder {
	return b.addJoin(joinClause{kind: "CROSS", table: syntax.As(subquery, alias), lateral: true})
}

// With adds a common table expression after any earlier ones.
// The name is quoted as a single identifier, and q may be a single SELECT or a compound.
//
//	recent := sqblpg.From("orders").Where(sqblpg.Gte("created_at", sqblpg.P(1)))
//	sqblpg.From("recent_orders").With("recent_orders", recent)
//	// WITH "recent_orders" AS (SELECT * FROM "orders" WHERE "created_at" >= $1) SELECT * FROM "recent_orders"
func (b SelectBuilder) With(name string, q Query) SelectBuilder {
	b.ctes = append(slices.Clone(b.ctes), cteClause{name: name, query: q})
	return b
}

// WithRecursive adds a recursive common table expression after any earlier ones.
// The whole clause becomes WITH RECURSIVE when any CTE is recursive.
// q is usually an anchor query combined with a recursive step by UnionAll.
//
//	anchor := sqblpg.From("employees").Select("id", "manager_id").Where(sqblpg.IsNull("manager_id"))
//	step := sqblpg.From(sqblpg.As("employees", "e")).Select("e.id", "e.manager_id").
//	    InnerJoin(sqblpg.As("org", "o"), sqblpg.Eq("e.manager_id", "o.id"))
//	sqblpg.From("org").WithRecursive("org", anchor.UnionAll(step))
//	// WITH RECURSIVE "org" AS (SELECT ... UNION ALL SELECT ...) SELECT * FROM "org"
func (b SelectBuilder) WithRecursive(name string, q Query) SelectBuilder {
	b.ctes = append(slices.Clone(b.ctes), cteClause{name: name, query: q, recursive: true})
	return b
}

// Union combines the query with q by UNION, removing duplicate rows.
// The result is a CompoundBuilder, whose OrderBy and Limit apply to the combined rows.
//
//	sqblpg.From("users").Select("id").Union(sqblpg.From("archived_users").Select("id"))
//	// SELECT "id" FROM "users" UNION SELECT "id" FROM "archived_users"
func (b SelectBuilder) Union(q Query) CompoundBuilder { return newCompound(b, "UNION", q) }

// UnionAll combines the query with q by UNION ALL, keeping duplicate rows.
// See Union for how the result behaves.
func (b SelectBuilder) UnionAll(q Query) CompoundBuilder { return newCompound(b, "UNION ALL", q) }

// Intersect combines the query with q by INTERSECT, keeping rows present in both.
// See Union for how the result behaves.
func (b SelectBuilder) Intersect(q Query) CompoundBuilder { return newCompound(b, "INTERSECT", q) }

// IntersectAll combines the query with q by INTERSECT ALL, keeping duplicates.
// Not supported by SQLite or MySQL.
func (b SelectBuilder) IntersectAll(q Query) CompoundBuilder {
	return newCompound(b, "INTERSECT ALL", q)
}

// Except combines the query with q by EXCEPT, keeping rows absent from q.
// See Union for how the result behaves.
func (b SelectBuilder) Except(q Query) CompoundBuilder { return newCompound(b, "EXCEPT", q) }

// ExceptAll combines the query with q by EXCEPT ALL, keeping duplicates.
// Not supported by SQLite or MySQL.
func (b SelectBuilder) ExceptAll(q Query) CompoundBuilder {
	return newCompound(b, "EXCEPT ALL", q)
}
