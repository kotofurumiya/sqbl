// Package sqblsqlite provides SQLite-specific entry points for building SQL queries.
// It wires the SQLite and re-exports syntax helpers for convenience.
package sqblsqlite

import (
	"github.com/kotofurumiya/sqbl/builder"
	"github.com/kotofurumiya/sqbl/dialect"
)

func newSqblBuilder() builder.SelectBuilder {
	var b builder.SelectBuilder
	return b.Dialect(&dialect.SQLite{})
}

// From starts a SELECT query with the given table as the FROM clause.
//
//	sqblsqlite.From("users")
//	sqblsqlite.From(sqblsqlite.As("users", "u"))
//	sqblsqlite.From(sqblsqlite.From("orders").Select("user_id"))
func From(table any) builder.SelectBuilder {
	return newSqblBuilder().From(table)
}

// Select starts a SELECT query with the given columns, without a FROM clause.
// Useful for queries like SELECT 1 or SELECT datetime('now').
//
//	sqblsqlite.Select("id", "name")
//	sqblsqlite.Select(sqblsqlite.As(sqblsqlite.Fn("COUNT", "*"), "total"))
func Select(columns ...any) builder.SelectBuilder {
	return newSqblBuilder().Select(columns...)
}

// InsertInto starts an INSERT query targeting the given table.
//
//	sqblsqlite.InsertInto("users").Columns("name", "email").Values(sqbl.P(1), sqbl.P(2))
func InsertInto(table string) builder.InsertBuilder {
	var b builder.InsertBuilder
	return b.Dialect(&dialect.SQLite{}).Into(table)
}

// Update starts an UPDATE query targeting the given table.
//
//	sqblsqlite.Update("users").Set("name", sqbl.P(1)).Where(sqblsqlite.Eq("id", sqbl.P(2)))
func Update(table string) builder.UpdateBuilder {
	var b builder.UpdateBuilder
	return b.Dialect(&dialect.SQLite{}).Table(table)
}

// DeleteFrom starts a DELETE query targeting the given table.
//
//	sqblsqlite.DeleteFrom("users").Where(sqblsqlite.Eq("id", sqbl.P(1)))
func DeleteFrom(table string) builder.DeleteBuilder {
	var b builder.DeleteBuilder
	return b.Dialect(&dialect.SQLite{}).From(table)
}

// CreateTable starts a CREATE TABLE builder for the given table.
//
//	sqblsqlite.CreateTable("users").Column("id", "INTEGER", "NOT NULL").PrimaryKey("id")
func CreateTable(table string) builder.CreateTableBuilder {
	var b builder.CreateTableBuilder
	return b.Dialect(&dialect.SQLite{}).Table(table)
}

// CreateIndex starts a CREATE INDEX builder for the given index name.
//
//	sqblsqlite.CreateIndex("idx_users_email").On("users").Columns("email").Unique()
func CreateIndex(name string) builder.CreateIndexBuilder {
	var b builder.CreateIndexBuilder
	return b.Dialect(&dialect.SQLite{}).Name(name)
}

// DropTable starts a DROP TABLE builder for the given table.
//
//	sqblsqlite.DropTable("users").IfExists()
func DropTable(table string) builder.DropTableBuilder {
	var b builder.DropTableBuilder
	return b.Dialect(&dialect.SQLite{}).Table(table)
}

// DropIndex starts a DROP INDEX builder for the given index name.
//
//	sqblsqlite.DropIndex("idx_users_email").IfExists()
func DropIndex(name string) builder.DropIndexBuilder {
	var b builder.DropIndexBuilder
	return b.Dialect(&dialect.SQLite{}).Name(name)
}

// AlterTable starts an ALTER TABLE builder for the given table.
//
//	sqblsqlite.AlterTable("users").AddColumn("bio", "TEXT")
func AlterTable(table string) builder.AlterTableBuilder {
	var b builder.AlterTableBuilder
	return b.Dialect(&dialect.SQLite{}).Table(table)
}
