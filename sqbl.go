// Package sqbl provides a fluent SQL query builder with multi-dialect support.
// It supports SELECT, INSERT, UPDATE, DELETE, and DDL statements.
//
// The builders returned by this package have no dialect. Render them with
// ToSQLWithDialect, which lets one builder produce SQL for several databases.
// ToSQL returns an error on these builders; to render with ToSQL, start from a
// dialect package such as sqblpg, sqblmysql, or sqblsqlite instead.
//
//	q := sqbl.From("users").Select("id").Where(sqbl.Eq("id", sqbl.P(1)))
//	q.MustSQLWithDialect(&dialect.Postgres{}) // SELECT "id" FROM "users" WHERE "id" = $1
//	q.MustSQLWithDialect(&dialect.SQLite{})   // SELECT "id" FROM "users" WHERE "id" = ?1
//
// ToSQLWithDialect returns an error when the statement holds input that has no
// safe SQL form; the Must variants panic instead. The helper examples in this
// package show PostgreSQL output.
package sqbl

import "github.com/kotofurumiya/sqbl/builder"

// From starts a SELECT query with the given table as the FROM clause.
//
//	sqbl.From("users")
func From(table any) builder.SelectBuilder {
	var b builder.SelectBuilder
	return b.From(table)
}

// Select starts a SELECT query with the given columns, without a FROM clause.
//
//	sqbl.Select(sqbl.Fn("NOW"))
func Select(columns ...any) builder.SelectBuilder {
	var b builder.SelectBuilder
	return b.Select(columns...)
}

// InsertInto starts an INSERT query targeting the given table.
//
//	sqbl.InsertInto("users").Columns("name").Values(sqbl.P(1))
func InsertInto(table string) builder.InsertBuilder {
	var b builder.InsertBuilder
	return b.Into(table)
}

// Update starts an UPDATE query targeting the given table.
//
//	sqbl.Update("users").Set("name", sqbl.P(1)).Where(sqbl.Eq("id", sqbl.P(2)))
func Update(table string) builder.UpdateBuilder {
	var b builder.UpdateBuilder
	return b.Table(table)
}

// DeleteFrom starts a DELETE query targeting the given table.
//
//	sqbl.DeleteFrom("users").Where(sqbl.Eq("id", sqbl.P(1)))
func DeleteFrom(table string) builder.DeleteBuilder {
	var b builder.DeleteBuilder
	return b.From(table)
}

// CreateTable starts a CREATE TABLE builder for the given table.
//
//	sqbl.CreateTable("users").Column("id", "BIGINT", "NOT NULL").PrimaryKey("id")
func CreateTable(table string) builder.CreateTableBuilder {
	var b builder.CreateTableBuilder
	return b.Table(table)
}

// CreateIndex starts a CREATE INDEX builder for the given index name.
//
//	sqbl.CreateIndex("idx_users_email").On("users").Columns("email")
func CreateIndex(name string) builder.CreateIndexBuilder {
	var b builder.CreateIndexBuilder
	return b.Name(name)
}

// DropTable starts a DROP TABLE builder for the given table.
//
//	sqbl.DropTable("users").IfExists()
func DropTable(table string) builder.DropTableBuilder {
	var b builder.DropTableBuilder
	return b.Table(table)
}

// DropIndex starts a DROP INDEX builder for the given index name.
//
//	sqbl.DropIndex("idx_users_email").IfExists()
func DropIndex(name string) builder.DropIndexBuilder {
	var b builder.DropIndexBuilder
	return b.Name(name)
}

// AlterTable starts an ALTER TABLE builder for the given table.
//
//	sqbl.AlterTable("users").AddColumn("bio", "TEXT")
func AlterTable(table string) builder.AlterTableBuilder {
	var b builder.AlterTableBuilder
	return b.Table(table)
}

// CreateDatabase starts a CREATE DATABASE builder for the given database name.
//
//	sqbl.CreateDatabase("app_test")
func CreateDatabase(name string) builder.CreateDatabaseBuilder {
	var b builder.CreateDatabaseBuilder
	return b.Name(name)
}

// DropDatabase starts a DROP DATABASE builder for the given database name.
//
//	sqbl.DropDatabase("app_test").IfExists()
func DropDatabase(name string) builder.DropDatabaseBuilder {
	var b builder.DropDatabaseBuilder
	return b.Name(name)
}
