// Package sqblsqlite provides SQLite-specific entry points for building SQL queries.
// It wires the SqliteDialect and re-exports syntax helpers for convenience.
package sqblsqlite

import (
	"github.com/kotofurumiya/sqbl/builder"
	"github.com/kotofurumiya/sqbl/dialect"
)

func newSqblBuilder() *builder.SqlSelectBuilder {
	b := &builder.SqlSelectBuilder{}
	return b.Dialect(&dialect.SqliteDialect{})
}

// From starts a SELECT query with the given table as the FROM clause.
//
//	sqblsqlite.From("users")
//	sqblsqlite.From(sqblsqlite.As("users", "u"))
//	sqblsqlite.From(sqblsqlite.From("orders").Select("user_id"))
func From(table any) *builder.SqlSelectBuilder {
	return newSqblBuilder().From(table)
}

// Select starts a SELECT query with the given columns, without a FROM clause.
// Useful for queries like SELECT 1 or SELECT datetime('now').
//
//	sqblsqlite.Select("id", "name")
//	sqblsqlite.Select(sqblsqlite.As("COUNT(*)", "total"))
func Select(columns ...any) *builder.SqlSelectBuilder {
	return newSqblBuilder().Select(columns...)
}

// InsertInto starts an INSERT query targeting the given table.
//
//	sqblsqlite.InsertInto("users").Columns("name", "email").Values(sqbl.P(1), sqbl.P(2))
func InsertInto(table string) *builder.SqlInsertBuilder {
	b := &builder.SqlInsertBuilder{}
	return b.Dialect(&dialect.SqliteDialect{}).Into(table)
}

// Update starts an UPDATE query targeting the given table.
//
//	sqblsqlite.Update("users").Set("name", sqbl.P(1)).Where(sqblsqlite.Eq("id", sqbl.P(2)))
func Update(table string) *builder.SqlUpdateBuilder {
	b := &builder.SqlUpdateBuilder{}
	return b.Dialect(&dialect.SqliteDialect{}).Table(table)
}

// DeleteFrom starts a DELETE query targeting the given table.
//
//	sqblsqlite.DeleteFrom("users").Where(sqblsqlite.Eq("id", sqbl.P(1)))
func DeleteFrom(table string) *builder.SqlDeleteBuilder {
	b := &builder.SqlDeleteBuilder{}
	return b.Dialect(&dialect.SqliteDialect{}).From(table)
}

// CreateTable starts a CREATE TABLE builder for the given table.
//
//	sqblsqlite.CreateTable("users").Column("id", "INTEGER", "NOT NULL").PrimaryKey("id")
func CreateTable(table string) *builder.SqlCreateTableBuilder {
	b := &builder.SqlCreateTableBuilder{}
	return b.Dialect(&dialect.SqliteDialect{}).Table(table)
}

// CreateIndex starts a CREATE INDEX builder for the given index name.
//
//	sqblsqlite.CreateIndex("idx_users_email").On("users").Columns("email").Unique()
func CreateIndex(name string) *builder.SqlCreateIndexBuilder {
	b := &builder.SqlCreateIndexBuilder{}
	return b.Dialect(&dialect.SqliteDialect{}).Name(name)
}

// DropTable starts a DROP TABLE builder for the given table.
//
//	sqblsqlite.DropTable("users").IfExists()
func DropTable(table string) *builder.SqlDropTableBuilder {
	b := &builder.SqlDropTableBuilder{}
	return b.Dialect(&dialect.SqliteDialect{}).Table(table)
}

// DropIndex starts a DROP INDEX builder for the given index name.
//
//	sqblsqlite.DropIndex("idx_users_email").IfExists()
func DropIndex(name string) *builder.SqlDropIndexBuilder {
	b := &builder.SqlDropIndexBuilder{}
	return b.Dialect(&dialect.SqliteDialect{}).Name(name)
}

// AlterTable starts an ALTER TABLE builder for the given table.
//
//	sqblsqlite.AlterTable("users").AddColumn("bio", "TEXT")
func AlterTable(table string) *builder.SqlAlterTableBuilder {
	b := &builder.SqlAlterTableBuilder{}
	return b.Dialect(&dialect.SqliteDialect{}).Table(table)
}
