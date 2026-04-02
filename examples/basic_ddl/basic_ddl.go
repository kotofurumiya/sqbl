package main

import (
	"fmt"

	sqbl "github.com/kotofurumiya/sqbl/sqblpg"
)

func main() {
	// CREATE TABLE with columns and constraints
	// CREATE TABLE "users" (
	//   "id" BIGSERIAL NOT NULL,
	//   "email" TEXT NOT NULL,
	//   "name" TEXT NOT NULL,
	//   "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	//   PRIMARY KEY ("id"),
	//   UNIQUE ("email")
	// );
	sql1 := sqbl.CreateTable("users").
		Column("id", "BIGSERIAL", "NOT NULL").
		Column("email", "TEXT", "NOT NULL").
		Column("name", "TEXT", "NOT NULL").
		Column("created_at", "TIMESTAMPTZ", "NOT NULL", "DEFAULT NOW()").
		PrimaryKey("id").
		Unique("email").
		ToSql()
	fmt.Println(sql1)

	// CREATE TABLE with foreign key
	// CREATE TABLE IF NOT EXISTS "posts" (
	//   "id" BIGSERIAL NOT NULL,
	//   "user_id" BIGINT NOT NULL,
	//   "title" TEXT NOT NULL,
	//   PRIMARY KEY ("id"),
	//   FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE
	// );
	sql2 := sqbl.CreateTable("posts").
		IfNotExists().
		Column("id", "BIGSERIAL", "NOT NULL").
		Column("user_id", "BIGINT", "NOT NULL").
		Column("title", "TEXT", "NOT NULL").
		PrimaryKey("id").
		ForeignKey([]string{"user_id"}, "users", []string{"id"}, "ON DELETE CASCADE").
		ToSql()
	fmt.Println(sql2)

	// CREATE INDEX
	// CREATE INDEX "idx_posts_user_id" ON "posts" ("user_id");
	sql3 := sqbl.CreateIndex("idx_posts_user_id").On("posts").Columns("user_id").ToSql()
	fmt.Println(sql3)

	// CREATE UNIQUE INDEX
	// CREATE UNIQUE INDEX "idx_users_email" ON "users" ("email");
	sql4 := sqbl.CreateIndex("idx_users_email").On("users").Columns("email").Unique().ToSql()
	fmt.Println(sql4)

	// ALTER TABLE: add column
	// ALTER TABLE "users" ADD COLUMN "bio" TEXT;
	sql5 := sqbl.AlterTable("users").AddColumn("bio", "TEXT").ToSql()
	fmt.Println(sql5)

	// ALTER TABLE: rename column
	// ALTER TABLE "users" RENAME COLUMN "name" TO "full_name";
	sql6 := sqbl.AlterTable("users").RenameColumn("name", "full_name").ToSql()
	fmt.Println(sql6)

	// DROP TABLE
	// DROP TABLE IF EXISTS "posts" CASCADE;
	sql7 := sqbl.DropTable("posts").IfExists().Cascade().ToSql()
	fmt.Println(sql7)

	// DROP INDEX
	// DROP INDEX IF EXISTS "idx_users_email";
	sql8 := sqbl.DropIndex("idx_users_email").IfExists().ToSql()
	fmt.Println(sql8)
}
