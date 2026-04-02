// Package sqblpg provides PostgreSQL-specific entry points for building SQL queries.
// It wires the PostgresDialect and re-exports syntax helpers for convenience.
package sqblpg

import (
	"github.com/kotofurumiya/sqbl/builder"
	"github.com/kotofurumiya/sqbl/dialect"
)

func newSqblBuilder() *builder.SqlSelectBuilder {
	b := &builder.SqlSelectBuilder{}
	return b.Dialect(&dialect.PostgresDialect{})
}

// From starts a SELECT query with the given table as the FROM clause.
//
//	sqblpg.From("users")
//	sqblpg.From(sqblpg.As("users", "u"))
//	sqblpg.From(sqblpg.From("orders").Select("user_id"))
func From(table any) *builder.SqlSelectBuilder {
	return newSqblBuilder().From(table)
}

// Select starts a SELECT query with the given columns, without a FROM clause.
// Useful for queries like SELECT 1 or SELECT NOW().
//
//	sqblpg.Select("id", "name")
//	sqblpg.Select(sqblpg.As("COUNT(*)", "total"))
func Select(columns ...any) *builder.SqlSelectBuilder {
	return newSqblBuilder().Select(columns...)
}

// InsertInto starts an INSERT query targeting the given table.
//
//	sqblpg.InsertInto("users").Columns("name", "email").Values(sqbl.P(1), sqbl.P(2))
func InsertInto(table string) *builder.SqlInsertBuilder {
	b := &builder.SqlInsertBuilder{}
	return b.Dialect(&dialect.PostgresDialect{}).Into(table)
}

// Update starts an UPDATE query targeting the given table.
//
//	sqblpg.Update("users").Set("name", sqbl.P(1)).Where(sqblpg.Eq("id", sqbl.P(2)))
func Update(table string) *builder.SqlUpdateBuilder {
	b := &builder.SqlUpdateBuilder{}
	return b.Dialect(&dialect.PostgresDialect{}).Table(table)
}

// DeleteFrom starts a DELETE query targeting the given table.
//
//	sqblpg.DeleteFrom("users").Where(sqblpg.Eq("id", sqbl.P(1)))
func DeleteFrom(table string) *builder.SqlDeleteBuilder {
	b := &builder.SqlDeleteBuilder{}
	return b.Dialect(&dialect.PostgresDialect{}).From(table)
}

// CreateTable starts a CREATE TABLE builder for the given table.
//
//	sqblpg.CreateTable("users").Column("id", "BIGINT", "NOT NULL").PrimaryKey("id")
func CreateTable(table string) *builder.SqlCreateTableBuilder {
	b := &builder.SqlCreateTableBuilder{}
	return b.Dialect(&dialect.PostgresDialect{}).Table(table)
}

// CreateIndex starts a CREATE INDEX builder for the given index name.
//
//	sqblpg.CreateIndex("idx_users_email").On("users").Columns("email").Unique()
func CreateIndex(name string) *builder.SqlCreateIndexBuilder {
	b := &builder.SqlCreateIndexBuilder{}
	return b.Dialect(&dialect.PostgresDialect{}).Name(name)
}

// DropTable starts a DROP TABLE builder for the given table.
//
//	sqblpg.DropTable("users").IfExists()
//	sqblpg.DropTable("users").IfExists().Cascade()
func DropTable(table string) *builder.SqlDropTableBuilder {
	b := &builder.SqlDropTableBuilder{}
	return b.Dialect(&dialect.PostgresDialect{}).Table(table)
}

// DropIndex starts a DROP INDEX builder for the given index name.
//
//	sqblpg.DropIndex("idx_users_email").IfExists()
func DropIndex(name string) *builder.SqlDropIndexBuilder {
	b := &builder.SqlDropIndexBuilder{}
	return b.Dialect(&dialect.PostgresDialect{}).Name(name)
}

// AlterTable starts an ALTER TABLE builder for the given table.
//
//	sqblpg.AlterTable("users").AddColumn("bio", "TEXT")
//	sqblpg.AlterTable("users").RenameColumn("fullname", "name")
func AlterTable(table string) *builder.SqlAlterTableBuilder {
	b := &builder.SqlAlterTableBuilder{}
	return b.Dialect(&dialect.PostgresDialect{}).Table(table)
}
