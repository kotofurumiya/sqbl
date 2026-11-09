// Package sqblpg provides PostgreSQL-specific entry points for building SQL queries.
// It wires the Postgres and re-exports syntax helpers for convenience.
package sqblpg

import (
	"github.com/kotofurumiya/sqbl/builder"
	"github.com/kotofurumiya/sqbl/dialect"
)

func newSqblBuilder() builder.SelectBuilder {
	var b builder.SelectBuilder
	return b.Dialect(&dialect.Postgres{})
}

// From starts a SELECT query with the given table as the FROM clause.
//
//	sqblpg.From("users")
//	sqblpg.From(sqblpg.As("users", "u"))
//	sqblpg.From(sqblpg.From("orders").Select("user_id"))
func From(table any) builder.SelectBuilder {
	return newSqblBuilder().From(table)
}

// Select starts a SELECT query with the given columns, without a FROM clause.
// Useful for queries like SELECT 1 or SELECT NOW().
//
//	sqblpg.Select("id", "name")
//	sqblpg.Select(sqblpg.As(sqblpg.Fn("COUNT", "*"), "total"))
func Select(columns ...any) builder.SelectBuilder {
	return newSqblBuilder().Select(columns...)
}

// InsertInto starts an INSERT query targeting the given table.
//
//	sqblpg.InsertInto("users").Columns("name", "email").Values(sqbl.P(1), sqbl.P(2))
func InsertInto(table string) builder.InsertBuilder {
	var b builder.InsertBuilder
	return b.Dialect(&dialect.Postgres{}).Into(table)
}

// Update starts an UPDATE query targeting the given table.
//
//	sqblpg.Update("users").Set("name", sqbl.P(1)).Where(sqblpg.Eq("id", sqbl.P(2)))
func Update(table string) builder.UpdateBuilder {
	var b builder.UpdateBuilder
	return b.Dialect(&dialect.Postgres{}).Table(table)
}

// DeleteFrom starts a DELETE query targeting the given table.
//
//	sqblpg.DeleteFrom("users").Where(sqblpg.Eq("id", sqbl.P(1)))
func DeleteFrom(table string) builder.DeleteBuilder {
	var b builder.DeleteBuilder
	return b.Dialect(&dialect.Postgres{}).From(table)
}

// CreateTable starts a CREATE TABLE builder for the given table.
//
//	sqblpg.CreateTable("users").Column("id", "BIGINT", "NOT NULL").PrimaryKey("id")
func CreateTable(table string) builder.CreateTableBuilder {
	var b builder.CreateTableBuilder
	return b.Dialect(&dialect.Postgres{}).Table(table)
}

// CreateIndex starts a CREATE INDEX builder for the given index name.
//
//	sqblpg.CreateIndex("idx_users_email").On("users").Columns("email").Unique()
func CreateIndex(name string) builder.CreateIndexBuilder {
	var b builder.CreateIndexBuilder
	return b.Dialect(&dialect.Postgres{}).Name(name)
}

// DropTable starts a DROP TABLE builder for the given table.
//
//	sqblpg.DropTable("users").IfExists()
//	sqblpg.DropTable("users").IfExists().Cascade()
func DropTable(table string) builder.DropTableBuilder {
	var b builder.DropTableBuilder
	return b.Dialect(&dialect.Postgres{}).Table(table)
}

// DropIndex starts a DROP INDEX builder for the given index name.
//
//	sqblpg.DropIndex("idx_users_email").IfExists()
func DropIndex(name string) builder.DropIndexBuilder {
	var b builder.DropIndexBuilder
	return b.Dialect(&dialect.Postgres{}).Name(name)
}

// AlterTable starts an ALTER TABLE builder for the given table.
//
//	sqblpg.AlterTable("users").AddColumn("bio", "TEXT")
//	sqblpg.AlterTable("users").RenameColumn("fullname", "name")
func AlterTable(table string) builder.AlterTableBuilder {
	var b builder.AlterTableBuilder
	return b.Dialect(&dialect.Postgres{}).Table(table)
}

// CreateDatabase starts a CREATE DATABASE builder for the given database name.
//
//	sqblpg.CreateDatabase("app_test")
func CreateDatabase(name string) builder.CreateDatabaseBuilder {
	var b builder.CreateDatabaseBuilder
	return b.Dialect(&dialect.Postgres{}).Name(name)
}

// DropDatabase starts a DROP DATABASE builder for the given database name.
//
//	sqblpg.DropDatabase("app_test").IfExists()
func DropDatabase(name string) builder.DropDatabaseBuilder {
	var b builder.DropDatabaseBuilder
	return b.Dialect(&dialect.Postgres{}).Name(name)
}
