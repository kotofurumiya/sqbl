package builder

import (
	"fmt"
	"strings"

	"github.com/kotofurumiya/sqbl/dialect"
	"github.com/kotofurumiya/sqbl/syntax"
)

// SqlInsertBuilder constructs a SQL INSERT statement using a fluent method chain.
// Use a dialect-specific constructor (e.g. sqblpg.InsertInto) to create an instance.
//
//	q := sqblpg.InsertInto("users").
//	    Columns("name", "email").
//	    Values(sqbl.P(1), sqbl.P(2))
//	sql := q.ToSql()
type SqlInsertBuilder struct {
	dialect          dialect.SqlDialect
	into             string
	orConflict       string   // SQLite: "REPLACE", "IGNORE", "ABORT", "FAIL", "ROLLBACK"
	columns          []string
	rows             []syntax.ValueRow
	conflictTarget   []string // columns for ON CONFLICT (col1, col2); nil means no conflict clause
	conflictAction   string   // "NOTHING" or "UPDATE"
	conflictSets     []string // raw SET fragments for DO UPDATE, e.g. "value = EXCLUDED.value"
	duplicateKeySets []string // raw SET fragments for ON DUPLICATE KEY UPDATE (MySQL)
	returning        []string
}

var _ SqlBuilder = (*SqlInsertBuilder)(nil)

// ToSql renders the INSERT statement with a trailing semicolon.
func (b *SqlInsertBuilder) ToSql() string {
	return b.renderSQL(b.dialect) + ";"
}

// ToSqlWithDialect renders the INSERT statement using the given dialect, without a trailing semicolon.
func (b *SqlInsertBuilder) ToSqlWithDialect(d dialect.SqlDialect) string {
	return b.renderSQL(d)
}

func (b *SqlInsertBuilder) renderSQL(d dialect.SqlDialect) string {
	var parts []string

	// INSERT [OR conflict] INTO table
	_, isMysql := d.(*dialect.MysqlDialect)
	_, isSqlite := d.(*dialect.SqliteDialect)

	into := "INSERT"
	if isSqlite && b.orConflict != "" {
		into += " OR " + b.orConflict
	}
	into += " INTO"
	if b.into != "" {
		into += " " + d.QuoteIdentifier(b.into)
	}
	parts = append(parts, into)

	// (col1, col2, ...)
	if cols := quoteIdentifiers(d, b.columns); cols != "" {
		parts = append(parts, "("+cols+")")
	}

	// VALUES (row1), (row2), ...
	if len(b.rows) > 0 {
		rows := mapJoin(b.rows, ", ", func(row syntax.ValueRow) string {
			return "(" + mapJoin(row, ", ", func(v any) string { return renderValue(d, v) }) + ")"
		})
		parts = append(parts, "VALUES "+rows)
	}

	// ON CONFLICT (PostgreSQL/SQLite) or ON DUPLICATE KEY UPDATE (MySQL)
	if isMysql {
		if len(b.duplicateKeySets) > 0 {
			parts = append(parts, "ON DUPLICATE KEY UPDATE "+strings.Join(b.duplicateKeySets, ", "))
		}
	} else {
		switch b.conflictAction {
		case "NOTHING":
			if target := quoteIdentifiers(d, b.conflictTarget); target != "" {
				parts = append(parts, "ON CONFLICT ("+target+") DO NOTHING")
			} else {
				parts = append(parts, "ON CONFLICT DO NOTHING")
			}
		case "UPDATE":
			if len(b.conflictSets) > 0 {
				target := quoteIdentifiers(d, b.conflictTarget)
				if target != "" {
					target = " (" + target + ")"
				}
				parts = append(parts, "ON CONFLICT"+target+" DO UPDATE SET "+strings.Join(b.conflictSets, ", "))
			}
		}
	}

	// RETURNING
	if ret := quoteIdentifiers(d, b.returning); ret != "" {
		parts = append(parts, "RETURNING "+ret)
	}

	return strings.Join(parts, " ")
}

// Dialect sets the SQL dialect used when rendering the query.
func (b *SqlInsertBuilder) Dialect(d dialect.SqlDialect) *SqlInsertBuilder {
	b2 := *b
	b2.dialect = d
	return &b2
}

// Into sets the target table for the INSERT statement.
//
//	b.Into("users")
func (b *SqlInsertBuilder) Into(table string) *SqlInsertBuilder {
	b2 := *b
	b2.into = table
	return &b2
}

// Columns sets the column list for the INSERT statement.
// Calling Columns again replaces the previous column list.
//
//	b.Columns("name", "email")
func (b *SqlInsertBuilder) Columns(cols ...string) *SqlInsertBuilder {
	b2 := *b
	b2.columns = cols
	return &b2
}

// Values adds a row of values to the INSERT statement.
// Call Values multiple times to insert multiple rows.
//
//	b.Values(sqbl.P(1), sqbl.P(2))
//	b.Values(sqbl.P(1), sqbl.P(2)).Values(sqbl.P(3), sqbl.P(4))
func (b *SqlInsertBuilder) Values(vals ...any) *SqlInsertBuilder {
	b2 := *b
	b2.rows = append(append([]syntax.ValueRow(nil), b.rows...), syntax.ValueRow(vals))
	return &b2
}

// OnConflictDoNothing adds an ON CONFLICT DO NOTHING clause.
// Rows that violate a unique or exclusion constraint are silently discarded.
//
//	sqblpg.InsertInto("tags").Columns("name").Values(sqbl.P(1)).OnConflictDoNothing()
func (b *SqlInsertBuilder) OnConflictDoNothing() *SqlInsertBuilder {
	b2 := *b
	b2.conflictAction = "NOTHING"
	return &b2
}

// OnConflict sets the conflict target columns and prepares the builder for a
// DO UPDATE clause. Chain DoUpdate immediately after to complete the upsert.
//
//	sqblpg.InsertInto("settings").Columns("user_id", "key", "value").
//	    Values(sqbl.P(1), sqbl.P(2), sqbl.P(3)).
//	    OnConflict("user_id", "key").DoUpdate("value = EXCLUDED.value")
func (b *SqlInsertBuilder) OnConflict(cols ...string) *SqlInsertBuilder {
	b2 := *b
	b2.conflictTarget = cols
	return &b2
}

// DoUpdate completes an ON CONFLICT … DO UPDATE SET clause.
// Each argument is a raw SQL assignment fragment, e.g. "value = EXCLUDED.value".
// Must be called after OnConflict.
func (b *SqlInsertBuilder) DoUpdate(sets ...string) *SqlInsertBuilder {
	b2 := *b
	b2.conflictAction = "UPDATE"
	b2.conflictSets = sets
	return &b2
}

// OnDuplicateKeyUpdate adds an ON DUPLICATE KEY UPDATE clause (MySQL only).
// Each argument is a raw SQL assignment fragment, e.g. "value = VALUES(value)".
//
//	sqblmysql.InsertInto("tags").Columns("name", "count").Values(sqbl.P(1), sqbl.P(2)).
//	    OnDuplicateKeyUpdate("count = VALUES(count)")
func (b *SqlInsertBuilder) OnDuplicateKeyUpdate(sets ...string) *SqlInsertBuilder {
	b2 := *b
	b2.duplicateKeySets = sets
	return &b2
}

// OrReplace sets the INSERT OR REPLACE conflict algorithm (SQLite only).
// Existing rows that conflict are deleted and replaced by the new row.
func (b *SqlInsertBuilder) OrReplace() *SqlInsertBuilder {
	b2 := *b
	b2.orConflict = "REPLACE"
	return &b2
}

// OrIgnore sets the INSERT OR IGNORE conflict algorithm (SQLite only).
// Rows that violate a constraint are silently discarded.
func (b *SqlInsertBuilder) OrIgnore() *SqlInsertBuilder {
	b2 := *b
	b2.orConflict = "IGNORE"
	return &b2
}

// OrAbort sets the INSERT OR ABORT conflict algorithm (SQLite only).
// The current INSERT statement is aborted and changes are rolled back.
func (b *SqlInsertBuilder) OrAbort() *SqlInsertBuilder {
	b2 := *b
	b2.orConflict = "ABORT"
	return &b2
}

// OrFail sets the INSERT OR FAIL conflict algorithm (SQLite only).
// The INSERT fails with an error but prior changes in the same transaction are retained.
func (b *SqlInsertBuilder) OrFail() *SqlInsertBuilder {
	b2 := *b
	b2.orConflict = "FAIL"
	return &b2
}

// OrRollback sets the INSERT OR ROLLBACK conflict algorithm (SQLite only).
// The entire transaction is rolled back on conflict.
func (b *SqlInsertBuilder) OrRollback() *SqlInsertBuilder {
	b2 := *b
	b2.orConflict = "ROLLBACK"
	return &b2
}

// Returning adds a RETURNING clause to the INSERT statement.
// The specified columns are returned for each inserted row.
// Supported by PostgreSQL and SQLite 3.35+. Ignored for MySQL.
//
//	sqblpg.InsertInto("users").Columns("name").Values(sqbl.P(1)).Returning("id", "created_at")
func (b *SqlInsertBuilder) Returning(cols ...string) *SqlInsertBuilder {
	b2 := *b
	b2.returning = cols
	return &b2
}

// renderValue renders a single value as a SQL string using the given dialect.
// Handles Parameter, bool, string, and other types.
func renderValue(d dialect.SqlDialect, v any) string {
	switch val := v.(type) {
	case syntax.Parameter:
		return val.ToSqlWithDialect(d)
	case bool:
		return d.Bool(val)
	case string:
		return val
	default:
		return fmt.Sprintf("%v", val)
	}
}
