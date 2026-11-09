// Package sqblmysql provides MySQL-specific entry points for building SQL queries.
// It wires the MySQL dialect into every builder and re-exports the syntax helpers,
// so one import covers a whole query.
//
// Render a statement with ToSQL, which returns an error when the statement holds
// input that has no safe SQL form, such as a value of an unsupported Go type.
// MustSQL panics instead, for statements built from constants.
//
//	sql, err := sqblmysql.From("users").Select("id").Where(sqblmysql.Eq("id", sqblmysql.P())).ToSQL()
//	// SELECT `id` FROM `users` WHERE `id` = ?;
package sqblmysql

import (
	"github.com/kotofurumiya/sqbl/builder"
	"github.com/kotofurumiya/sqbl/dialect"
)

func newSqblBuilder() builder.SelectBuilder {
	var b builder.SelectBuilder
	return b.Dialect(&dialect.MySQL{})
}

// From starts a SELECT query with the given table as the FROM clause.
//
//	sqblmysql.From("users")
//	sqblmysql.From(sqblmysql.As("users", "u"))
//	sqblmysql.From(sqblmysql.As(sqblmysql.From("orders").Select("user_id"), "o"))
func From(table any) builder.SelectBuilder {
	return newSqblBuilder().From(table)
}

// Select starts a SELECT query with the given columns, without a FROM clause.
// Useful for queries like SELECT 1 or SELECT NOW().
//
//	sqblmysql.Select("id", "name")
//	sqblmysql.Select(sqblmysql.As(sqblmysql.Fn("COUNT", "*"), "total"))
func Select(columns ...any) builder.SelectBuilder {
	return newSqblBuilder().Select(columns...)
}

// InsertInto starts an INSERT query targeting the given table.
//
//	sqblmysql.InsertInto("users").Columns("name", "email").Values(sqblmysql.P(), sqblmysql.P())
func InsertInto(table string) builder.InsertBuilder {
	var b builder.InsertBuilder
	return b.Dialect(&dialect.MySQL{}).Into(table)
}

// Update starts an UPDATE query targeting the given table.
//
//	sqblmysql.Update("users").Set("name", sqblmysql.P()).Where(sqblmysql.Eq("id", sqblmysql.P()))
func Update(table string) builder.UpdateBuilder {
	var b builder.UpdateBuilder
	return b.Dialect(&dialect.MySQL{}).Table(table)
}

// DeleteFrom starts a DELETE query targeting the given table.
//
//	sqblmysql.DeleteFrom("users").Where(sqblmysql.Eq("id", sqblmysql.P()))
func DeleteFrom(table string) builder.DeleteBuilder {
	var b builder.DeleteBuilder
	return b.Dialect(&dialect.MySQL{}).From(table)
}

// CreateTable starts a CREATE TABLE builder for the given table.
//
//	sqblmysql.CreateTable("users").Column("id", "BIGINT", "NOT NULL").PrimaryKey("id")
func CreateTable(table string) builder.CreateTableBuilder {
	var b builder.CreateTableBuilder
	return b.Dialect(&dialect.MySQL{}).Table(table)
}

// CreateIndex starts a CREATE INDEX builder for the given index name.
//
//	sqblmysql.CreateIndex("idx_users_email").On("users").Columns("email").Unique()
func CreateIndex(name string) builder.CreateIndexBuilder {
	var b builder.CreateIndexBuilder
	return b.Dialect(&dialect.MySQL{}).Name(name)
}

// DropTable starts a DROP TABLE builder for the given table.
//
//	sqblmysql.DropTable("users").IfExists()
func DropTable(table string) builder.DropTableBuilder {
	var b builder.DropTableBuilder
	return b.Dialect(&dialect.MySQL{}).Table(table)
}

// DropIndex starts a DROP INDEX builder for the given index name.
// MySQL requires specifying the table with On().
//
//	sqblmysql.DropIndex("idx_users_email").On("users")
func DropIndex(name string) builder.DropIndexBuilder {
	var b builder.DropIndexBuilder
	return b.Dialect(&dialect.MySQL{}).Name(name)
}

// AlterTable starts an ALTER TABLE builder for the given table.
//
//	sqblmysql.AlterTable("users").AddColumn("bio", "TEXT")
//	sqblmysql.AlterTable("users").RenameColumn("fullname", "name")
func AlterTable(table string) builder.AlterTableBuilder {
	var b builder.AlterTableBuilder
	return b.Dialect(&dialect.MySQL{}).Table(table)
}

// CreateDatabase starts a CREATE DATABASE builder for the given database name.
//
//	sqblmysql.CreateDatabase("app_test")
func CreateDatabase(name string) builder.CreateDatabaseBuilder {
	var b builder.CreateDatabaseBuilder
	return b.Dialect(&dialect.MySQL{}).Name(name)
}

// DropDatabase starts a DROP DATABASE builder for the given database name.
//
//	sqblmysql.DropDatabase("app_test").IfExists()
func DropDatabase(name string) builder.DropDatabaseBuilder {
	var b builder.DropDatabaseBuilder
	return b.Dialect(&dialect.MySQL{}).Name(name)
}
