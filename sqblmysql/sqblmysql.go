// Package sqblmysql provides MySQL-specific entry points for building SQL queries.
// It wires the MysqlDialect and re-exports syntax helpers for convenience.
package sqblmysql

import (
	"github.com/kotofurumiya/sqbl/builder"
	"github.com/kotofurumiya/sqbl/dialect"
)

func newSqblBuilder() *builder.SqlSelectBuilder {
	b := &builder.SqlSelectBuilder{}
	return b.Dialect(&dialect.MysqlDialect{})
}

// From starts a SELECT query with the given table as the FROM clause.
//
//	sqblmysql.From("users")
//	sqblmysql.From(sqblmysql.As("users", "u"))
//	sqblmysql.From(sqblmysql.From("orders").Select("user_id"))
func From(table any) *builder.SqlSelectBuilder {
	return newSqblBuilder().From(table)
}

// Select starts a SELECT query with the given columns, without a FROM clause.
// Useful for queries like SELECT 1 or SELECT NOW().
//
//	sqblmysql.Select("id", "name")
//	sqblmysql.Select(sqblmysql.As("COUNT(*)", "total"))
func Select(columns ...any) *builder.SqlSelectBuilder {
	return newSqblBuilder().Select(columns...)
}

// InsertInto starts an INSERT query targeting the given table.
//
//	sqblmysql.InsertInto("users").Columns("name", "email").Values(sqbl.P(1), sqbl.P(2))
func InsertInto(table string) *builder.SqlInsertBuilder {
	b := &builder.SqlInsertBuilder{}
	return b.Dialect(&dialect.MysqlDialect{}).Into(table)
}

// Update starts an UPDATE query targeting the given table.
//
//	sqblmysql.Update("users").Set("name", sqbl.P(1)).Where(sqblmysql.Eq("id", sqbl.P(2)))
func Update(table string) *builder.SqlUpdateBuilder {
	b := &builder.SqlUpdateBuilder{}
	return b.Dialect(&dialect.MysqlDialect{}).Table(table)
}

// DeleteFrom starts a DELETE query targeting the given table.
//
//	sqblmysql.DeleteFrom("users").Where(sqblmysql.Eq("id", sqbl.P(1)))
func DeleteFrom(table string) *builder.SqlDeleteBuilder {
	b := &builder.SqlDeleteBuilder{}
	return b.Dialect(&dialect.MysqlDialect{}).From(table)
}

// CreateTable starts a CREATE TABLE builder for the given table.
//
//	sqblmysql.CreateTable("users").Column("id", "BIGINT", "NOT NULL").PrimaryKey("id")
func CreateTable(table string) *builder.SqlCreateTableBuilder {
	b := &builder.SqlCreateTableBuilder{}
	return b.Dialect(&dialect.MysqlDialect{}).Table(table)
}

// CreateIndex starts a CREATE INDEX builder for the given index name.
//
//	sqblmysql.CreateIndex("idx_users_email").On("users").Columns("email").Unique()
func CreateIndex(name string) *builder.SqlCreateIndexBuilder {
	b := &builder.SqlCreateIndexBuilder{}
	return b.Dialect(&dialect.MysqlDialect{}).Name(name)
}

// DropTable starts a DROP TABLE builder for the given table.
//
//	sqblmysql.DropTable("users").IfExists()
func DropTable(table string) *builder.SqlDropTableBuilder {
	b := &builder.SqlDropTableBuilder{}
	return b.Dialect(&dialect.MysqlDialect{}).Table(table)
}

// DropIndex starts a DROP INDEX builder for the given index name.
// MySQL requires specifying the table with On().
//
//	sqblmysql.DropIndex("idx_users_email").On("users")
func DropIndex(name string) *builder.SqlDropIndexBuilder {
	b := &builder.SqlDropIndexBuilder{}
	return b.Dialect(&dialect.MysqlDialect{}).Name(name)
}

// AlterTable starts an ALTER TABLE builder for the given table.
//
//	sqblmysql.AlterTable("users").AddColumn("bio", "TEXT")
//	sqblmysql.AlterTable("users").RenameColumn("fullname", "name")
func AlterTable(table string) *builder.SqlAlterTableBuilder {
	b := &builder.SqlAlterTableBuilder{}
	return b.Dialect(&dialect.MysqlDialect{}).Table(table)
}
