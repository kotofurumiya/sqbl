package main

import (
	"fmt"

	sqbl "github.com/kotofurumiya/sqbl/sqblpg"
)

func main() {
	// SELECT
	//     "id",
	//     "name"
	// FROM
	//     "users";
	sql1 := sqbl.From("users").Select("id", "name").ToSql()
	fmt.Println(sql1)

	// SELECT
	//     "id",
	//     "name"
	// FROM
	//     "users"
	// WHERE
	//     "active" = TRUE
	// ORDER BY
	//     "name" ASC
	// LIMIT 10;
	sql2 := sqbl.From("users").
		Select("id", "name").
		Where(sqbl.Eq("active", true)).
		OrderBy(sqbl.Asc("name")).
		Limit(10).
		ToSql()
	fmt.Println(sql2)

	// SELECT
	//     "id",
	//     "name"
	// FROM
	//     "users"
	// WHERE
	//     "role" = 'admin' OR "role" = 'moderator';
	sql3 := sqbl.From("users").
		Select("id", "name").
		Where(sqbl.Or(
			sqbl.Eq("role", "'admin'"),
			sqbl.Eq("role", "'moderator'"),
		)).
		ToSql()
	fmt.Println(sql3)
}
