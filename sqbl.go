// Package sqbl provides a fluent SQL query builder with multi-dialect support.
// It supports SELECT, INSERT, UPDATE, DELETE, and DDL statements.
//
// The builders returned by this package have no dialect. Render them with
// ToSqlWithDialect, which lets one builder produce SQL for several databases.
// ToSql panics on these builders; to render with ToSql, start from a dialect
// package such as sqblpg, sqblmysql, or sqblsqlite instead.
//
//	q := sqbl.From("users").Select("id").Where(sqbl.Eq("id", sqbl.P(1)))
//	q.ToSqlWithDialect(&dialect.PostgresDialect{}) // SELECT "id" FROM "users" WHERE "id" = $1
//	q.ToSqlWithDialect(&dialect.SqliteDialect{})   // SELECT "id" FROM "users" WHERE "id" = ?
package sqbl

import "github.com/kotofurumiya/sqbl/builder"

// From starts a SELECT query with the given table as the FROM clause.
//
//	sqbl.From("users")
func From(table any) builder.SqlSelectBuilder {
	var b builder.SqlSelectBuilder
	return b.From(table)
}

// Select starts a SELECT query with the given columns, without a FROM clause.
//
//	sqbl.Select(sqbl.Fn("NOW"))
func Select(columns ...any) builder.SqlSelectBuilder {
	var b builder.SqlSelectBuilder
	return b.Select(columns...)
}

// InsertInto starts an INSERT query targeting the given table.
//
//	sqbl.InsertInto("users").Columns("name").Values(sqbl.P(1))
func InsertInto(table string) builder.SqlInsertBuilder {
	var b builder.SqlInsertBuilder
	return b.Into(table)
}

// Update starts an UPDATE query targeting the given table.
//
//	sqbl.Update("users").Set("name", sqbl.P(1)).Where(sqbl.Eq("id", sqbl.P(2)))
func Update(table string) builder.SqlUpdateBuilder {
	var b builder.SqlUpdateBuilder
	return b.Table(table)
}

// DeleteFrom starts a DELETE query targeting the given table.
//
//	sqbl.DeleteFrom("users").Where(sqbl.Eq("id", sqbl.P(1)))
func DeleteFrom(table string) builder.SqlDeleteBuilder {
	var b builder.SqlDeleteBuilder
	return b.From(table)
}

// CreateTable starts a CREATE TABLE builder for the given table.
//
//	sqbl.CreateTable("users").Column("id", "BIGINT", "NOT NULL").PrimaryKey("id")
func CreateTable(table string) builder.SqlCreateTableBuilder {
	var b builder.SqlCreateTableBuilder
	return b.Table(table)
}

// CreateIndex starts a CREATE INDEX builder for the given index name.
//
//	sqbl.CreateIndex("idx_users_email").On("users").Columns("email")
func CreateIndex(name string) builder.SqlCreateIndexBuilder {
	var b builder.SqlCreateIndexBuilder
	return b.Name(name)
}

// DropTable starts a DROP TABLE builder for the given table.
//
//	sqbl.DropTable("users").IfExists()
func DropTable(table string) builder.SqlDropTableBuilder {
	var b builder.SqlDropTableBuilder
	return b.Table(table)
}

// DropIndex starts a DROP INDEX builder for the given index name.
//
//	sqbl.DropIndex("idx_users_email").IfExists()
func DropIndex(name string) builder.SqlDropIndexBuilder {
	var b builder.SqlDropIndexBuilder
	return b.Name(name)
}

// AlterTable starts an ALTER TABLE builder for the given table.
//
//	sqbl.AlterTable("users").AddColumn("bio", "TEXT")
func AlterTable(table string) builder.SqlAlterTableBuilder {
	var b builder.SqlAlterTableBuilder
	return b.Table(table)
}
