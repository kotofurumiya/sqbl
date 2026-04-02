package shared

import (
	"context"
	"errors"
	"testing"

	"github.com/kotofurumiya/sqbl/syntax"
)

// RunInjection_ValueAsParameter verifies that a classic SQL injection payload
// passed as a bound parameter value is treated as a literal string and does not
// execute as SQL.
func RunInjection_ValueAsParameter(t testing.TB, conn Conn, s Suite) {
	t.Helper()
	ctx := context.Background()

	mustExec(t, conn, ctx, "create table", `
		CREATE TEMPORARY TABLE victims (
			id   INTEGER PRIMARY KEY,
			name TEXT NOT NULL
		)
	`)

	insertSQL := s.InsertInto("victims").
		Columns("id", "name").
		Values(syntax.P(1), syntax.P(2)).
		ToSql()
	mustExec(t, conn, ctx, "insert", insertSQL, 1, "Alice")

	// The payload would break out of a string literal in a concatenated query.
	// Because the builder uses bind parameters the whole payload is sent as a
	// single string value and the database never parses it as SQL.
	injectionPayload := "'; DROP TABLE victims; --"

	selectSQL := s.From("victims").
		Select("name").
		Where(syntax.Eq("name", syntax.P(1))).
		ToSql()

	var matched string
	err := conn.QueryRow(ctx, selectSQL, injectionPayload).Scan(&matched)

	// ErrNoRows is acceptable; any other error indicates the query itself broke.
	if err != nil && !errors.Is(err, ErrNoRows) {
		t.Fatalf("unexpected error from injection attempt: %v", err)
	}
	if matched == "Alice" {
		t.Error("injection payload unexpectedly matched a legitimate row")
	}

	// The table must still exist and contain the original row.
	var count int
	countSQL := s.From("victims").Select("COUNT(*)").ToSql()
	if err := conn.QueryRow(ctx, countSQL).Scan(&count); err != nil {
		t.Fatalf("count after injection attempt: %v", err)
	}
	if count != 1 {
		t.Errorf("row count = %d after injection attempt; want 1 (table may have been dropped)", count)
	}
}

// RunInjection_StoredValue verifies that an injection payload stored as a column
// value is retrieved verbatim. Each payload is inserted with an explicit id so
// the test does not rely on auto-increment behaviour.
func RunInjection_StoredValue(t testing.TB, conn Conn, s Suite) {
	t.Helper()
	ctx := context.Background()

	mustExec(t, conn, ctx, "create table", `
		CREATE TEMPORARY TABLE notes (
			id      INTEGER PRIMARY KEY,
			content TEXT NOT NULL
		)
	`)

	payloads := []string{
		"'; DROP TABLE notes; --",
		"' OR '1'='1",
		"' UNION SELECT version() --",
		`\'; DROP TABLE notes; --`,
	}

	insertSQL := s.InsertInto("notes").
		Columns("id", "content").
		Values(syntax.P(1), syntax.P(2)).
		ToSql()
	selectSQL := s.From("notes").
		Select("content").
		Where(syntax.Eq("id", syntax.P(1))).
		ToSql()

	for i, payload := range payloads {
		id := i + 1
		mustExec(t, conn, ctx, "insert payload", insertSQL, id, payload)

		var got string
		if err := conn.QueryRow(ctx, selectSQL, id).Scan(&got); err != nil {
			t.Fatalf("payload[%d] select: %v", i, err)
		}
		if got != payload {
			t.Errorf("payload[%d]: stored %q; want %q", i, got, payload)
		}
	}

	// Confirm all rows still exist — none of the payloads triggered a DROP.
	var count int
	countSQL := s.From("notes").Select("COUNT(*)").ToSql()
	if err := conn.QueryRow(ctx, countSQL).Scan(&count); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != len(payloads) {
		t.Errorf("row count = %d; want %d (a payload may have mutated the table)", count, len(payloads))
	}
}

// RunInjection_LikeWildcardsInEqValue verifies that SQL wildcard characters
// (% and _) inside a value used with Eq are treated as literals, not as
// pattern characters.
func RunInjection_LikeWildcardsInEqValue(t testing.TB, conn Conn, s Suite) {
	t.Helper()
	ctx := context.Background()

	mustExec(t, conn, ctx, "create table", `
		CREATE TEMPORARY TABLE tags (
			id   INTEGER PRIMARY KEY,
			name TEXT NOT NULL
		)
	`)

	insertSQL := s.InsertInto("tags").
		Columns("id", "name").
		Values(syntax.P(1), syntax.P(2)).
		Values(syntax.P(3), syntax.P(4)).
		Values(syntax.P(5), syntax.P(6)).
		ToSql()
	mustExec(t, conn, ctx, "insert", insertSQL, 1, "go", 2, "golang", 3, "go%lang")

	selectSQL := s.From("tags").
		Select("COUNT(*)").
		Where(syntax.Eq("name", syntax.P(1))).
		ToSql()

	var count int

	// "%" is not stored as a row, so the result must be 0.
	if err := conn.QueryRow(ctx, selectSQL, "%").Scan(&count); err != nil {
		t.Fatalf("select with %% as Eq value: %v", err)
	}
	if count != 0 {
		t.Errorf("Eq('%%') matched %d rows; want 0 (wildcard must not expand in Eq)", count)
	}

	// Eq with "go%lang" must match exactly one row, not "golang".
	if err := conn.QueryRow(ctx, selectSQL, "go%lang").Scan(&count); err != nil {
		t.Fatalf("select with go%%lang as Eq value: %v", err)
	}
	if count != 1 {
		t.Errorf("Eq('go%%lang') matched %d rows; want 1", count)
	}
}

// RunInjection_OrderByIdentifier verifies that a column name containing
// injection characters passed to Asc/Desc is quoted by QuoteIdentifier and
// does not execute as SQL.
//
// A naive assumption: "ORDER BY column comes from a UI dropdown, so it's safe
// to embed it directly." sqbl quotes it, turning it into an unknown identifier
// that the database rejects — but never executes the payload.
func RunInjection_OrderByIdentifier(t testing.TB, conn Conn, s Suite) {
	t.Helper()
	ctx := context.Background()

	mustExec(t, conn, ctx, "create table", `
		CREATE TEMPORARY TABLE victims (
			id   INTEGER PRIMARY KEY,
			name TEXT NOT NULL
		)
	`)
	insertSQL := s.InsertInto("victims").
		Columns("id", "name").
		Values(syntax.P(1), syntax.P(2)).
		ToSql()
	mustExec(t, conn, ctx, "insert", insertSQL, 1, "Alice")

	// The payload contains a semicolon and DROP TABLE. Asc() quotes the whole
	// string as an identifier, so the database sees an unknown column name and
	// returns an error — it never reaches the DROP.
	evilCol := "id; DROP TABLE victims; --"
	orderSQL := s.From("victims").
		Select("id").
		OrderBy(syntax.Asc(evilCol)).
		ToSql()

	rows, err := conn.Query(ctx, orderSQL)
	// PostgreSQL and MySQL reject the unknown quoted identifier with an error.
	// SQLite silently falls back to treating an unresolved double-quoted token
	// as a string literal, so err may be nil — but in both cases the DROP TABLE
	// embedded in the payload is never executed.
	// Safety is verified by the table-survival check below.
	if err != nil {
		t.Logf("query correctly rejected unknown ORDER BY identifier: %v", err)
	} else {
		t.Log("query succeeded with unknown ORDER BY identifier (SQLite string-literal fallback)")
	}
	if rows != nil {
		rows.Close()
	}

	var count int
	countSQL := s.From("victims").Select("COUNT(*)").ToSql()
	if err := conn.QueryRow(ctx, countSQL).Scan(&count); err != nil {
		t.Fatalf("table inaccessible after ORDER BY injection attempt: %v", err)
	}
	if count != 1 {
		t.Errorf("row count = %d; want 1 (table may have been dropped)", count)
	}
}

// RunInjection_TableNameIdentifier verifies that a table name containing
// injection characters passed to From is quoted and does not execute as SQL.
//
// A naive assumption: "The table name comes from app config, not user input,
// so it's safe." sqbl quotes it; the database returns table-not-found instead
// of executing the payload.
func RunInjection_TableNameIdentifier(t testing.TB, conn Conn, s Suite) {
	t.Helper()
	ctx := context.Background()

	mustExec(t, conn, ctx, "create table", `
		CREATE TEMPORARY TABLE victims (
			id INTEGER PRIMARY KEY
		)
	`)
	insertSQL := s.InsertInto("victims").
		Columns("id").
		Values(syntax.P(1)).
		ToSql()
	mustExec(t, conn, ctx, "insert", insertSQL, 1)

	// The payload is used as the table name. QuoteIdentifier wraps it in
	// dialect-specific quotes; the database cannot find a table with that
	// literal name and returns an error without executing the DROP.
	evilTable := "victims; DROP TABLE victims; --"
	evilSQL := s.From(evilTable).Select("id").ToSql()

	rows, err := conn.Query(ctx, evilSQL)
	// PostgreSQL and MySQL reject the unknown quoted table name with an error.
	// SQLite has the same behavior for tables (no fallback to string literals
	// in the FROM clause), so err should also be non-nil there.
	// Safety is verified by the table-survival check below.
	if err != nil {
		t.Logf("query correctly rejected unknown table name: %v", err)
	} else {
		t.Log("query succeeded with unknown table name (unexpected; table-survival check still applies)")
	}
	if rows != nil {
		rows.Close()
	}

	var count int
	countSQL := s.From("victims").Select("COUNT(*)").ToSql()
	if err := conn.QueryRow(ctx, countSQL).Scan(&count); err != nil {
		t.Fatalf("table inaccessible after table-name injection attempt: %v", err)
	}
	if count != 1 {
		t.Errorf("row count = %d; want 1 (table may have been dropped)", count)
	}
}

// RunInjection_SelectColumnIdentifier verifies that a column name containing
// injection characters passed to Select is quoted and does not execute as SQL.
//
// A naive assumption: "The caller controls which columns to fetch via an API
// parameter — what could go wrong?" sqbl quotes it; unknown column → error,
// not a DROP.
func RunInjection_SelectColumnIdentifier(t testing.TB, conn Conn, s Suite) {
	t.Helper()
	ctx := context.Background()

	mustExec(t, conn, ctx, "create table", `
		CREATE TEMPORARY TABLE victims (
			id   INTEGER PRIMARY KEY,
			name TEXT NOT NULL
		)
	`)
	insertSQL := s.InsertInto("victims").
		Columns("id", "name").
		Values(syntax.P(1), syntax.P(2)).
		ToSql()
	mustExec(t, conn, ctx, "insert", insertSQL, 1, "Alice")

	evilCol := "id; DROP TABLE victims; --"
	selectSQL := s.From("victims").Select(evilCol).ToSql()

	rows, err := conn.Query(ctx, selectSQL)
	// PostgreSQL and MySQL reject the unknown quoted column with an error.
	// SQLite falls back to treating an unresolved double-quoted token as a
	// string literal, so err may be nil — but in both cases the DROP TABLE
	// embedded in the payload is never executed.
	// Safety is verified by the table-survival check below.
	if err != nil {
		t.Logf("query correctly rejected unknown SELECT column: %v", err)
	} else {
		t.Log("query succeeded with unknown SELECT column (SQLite string-literal fallback)")
	}
	if rows != nil {
		rows.Close()
	}

	var count int
	countSQL := s.From("victims").Select("COUNT(*)").ToSql()
	if err := conn.QueryRow(ctx, countSQL).Scan(&count); err != nil {
		t.Fatalf("table inaccessible after SELECT column injection attempt: %v", err)
	}
	if count != 1 {
		t.Errorf("row count = %d; want 1 (table may have been dropped)", count)
	}
}

// RunInjection_AliasIdentifier verifies that an AS alias containing injection
// characters is quoted by Quote and does not execute as SQL.
//
// A naive assumption: "Aliases are just display labels — nobody can inject
// through a column alias." Quote wraps it; the query succeeds and the value
// is returned correctly.
func RunInjection_AliasIdentifier(t testing.TB, conn Conn, s Suite) {
	t.Helper()
	ctx := context.Background()

	mustExec(t, conn, ctx, "create table", `
		CREATE TEMPORARY TABLE victims (
			id INTEGER PRIMARY KEY
		)
	`)
	insertSQL := s.InsertInto("victims").
		Columns("id").
		Values(syntax.P(1)).
		ToSql()
	mustExec(t, conn, ctx, "insert", insertSQL, 1)

	// The alias contains semicolons and DROP TABLE. Quote() wraps the whole
	// string so the database treats it as a single alias name. The query
	// succeeds and returns the correct value.
	evilAlias := "alias; DROP TABLE victims; --"
	selectSQL := s.From("victims").
		Select(syntax.As("id", evilAlias)).
		ToSql()

	var id int
	if err := conn.QueryRow(ctx, selectSQL).Scan(&id); err != nil {
		t.Fatalf("select with evil alias failed: %v", err)
	}
	if id != 1 {
		t.Errorf("id = %d; want 1", id)
	}

	var count int
	countSQL := s.From("victims").Select("COUNT(*)").ToSql()
	if err := conn.QueryRow(ctx, countSQL).Scan(&count); err != nil {
		t.Fatalf("table inaccessible after alias injection attempt: %v", err)
	}
	if count != 1 {
		t.Errorf("row count = %d; want 1 (table may have been dropped)", count)
	}
}

// RunInjection_CommentBasedPayloads verifies that comment-based injection
// payloads passed as bound parameter values are stored and retrieved verbatim,
// never interpreted as SQL.
//
// These payloads cover patterns commonly used to truncate queries or bypass
// authentication logic when user input is concatenated into SQL strings.
func RunInjection_CommentBasedPayloads(t testing.TB, conn Conn, s Suite) {
	t.Helper()
	ctx := context.Background()

	mustExec(t, conn, ctx, "create table", `
		CREATE TEMPORARY TABLE victims (
			id      INTEGER PRIMARY KEY,
			content TEXT NOT NULL
		)
	`)

	payloads := []string{
		"admin'--",              // comment truncation without space
		"' OR '1'='1' /*",      // OR injection + block comment opener
		"'; SELECT 1; --",      // stacked query via comment (SELECT 1 is DB-agnostic)
		"1 OR 1=1",             // boolean injection without quotes (integer-column assumption)
	}

	insertSQL := s.InsertInto("victims").
		Columns("id", "content").
		Values(syntax.P(1), syntax.P(2)).
		ToSql()
	selectSQL := s.From("victims").
		Select("content").
		Where(syntax.Eq("id", syntax.P(1))).
		ToSql()

	for i, payload := range payloads {
		id := i + 1
		mustExec(t, conn, ctx, "insert payload", insertSQL, id, payload)

		var got string
		if err := conn.QueryRow(ctx, selectSQL, id).Scan(&got); err != nil {
			t.Fatalf("payload[%d] select: %v", i, err)
		}
		if got != payload {
			t.Errorf("payload[%d]: stored %q; want %q", i, got, payload)
		}
	}

	var count int
	countSQL := s.From("victims").Select("COUNT(*)").ToSql()
	if err := conn.QueryRow(ctx, countSQL).Scan(&count); err != nil {
		t.Fatalf("count after comment-payload inserts: %v", err)
	}
	if count != len(payloads) {
		t.Errorf("row count = %d; want %d (a payload may have mutated the table)", count, len(payloads))
	}
}

// RunInjection_StackedInsertPayload verifies that a stacked-query payload
// attempting a secondary INSERT, passed as a bound parameter value, does not
// insert any additional rows.
//
// The payload tries to inject a second INSERT statement via a semicolon.
// Because the value is sent as a bind parameter the database treats the entire
// string as a literal and never parses the embedded INSERT.
func RunInjection_StackedInsertPayload(t testing.TB, conn Conn, s Suite) {
	t.Helper()
	ctx := context.Background()

	mustExec(t, conn, ctx, "create table", `
		CREATE TEMPORARY TABLE victims (
			id   INTEGER PRIMARY KEY,
			name TEXT NOT NULL
		)
	`)
	insertSQL := s.InsertInto("victims").
		Columns("id", "name").
		Values(syntax.P(1), syntax.P(2)).
		ToSql()
	mustExec(t, conn, ctx, "seed", insertSQL, 1, "Alice")

	// The payload embeds a stacked INSERT targeting id=999. If the driver were
	// to pass it as raw SQL the extra row would appear; bind parameters prevent
	// this entirely.
	stackedPayload := "'; INSERT INTO victims (id, name) VALUES (999, 'injected'); --"

	selectSQL := s.From("victims").
		Select("name").
		Where(syntax.Eq("name", syntax.P(1))).
		ToSql()

	var matched string
	err := conn.QueryRow(ctx, selectSQL, stackedPayload).Scan(&matched)
	if err != nil && !errors.Is(err, ErrNoRows) {
		t.Fatalf("unexpected error from stacked-payload attempt: %v", err)
	}

	// Row count must still be 1 — the injected INSERT must not have executed.
	var count int
	countSQL := s.From("victims").Select("COUNT(*)").ToSql()
	if err := conn.QueryRow(ctx, countSQL).Scan(&count); err != nil {
		t.Fatalf("count after stacked-insert attempt: %v", err)
	}
	if count != 1 {
		t.Errorf("row count = %d; want 1 (stacked INSERT may have executed)", count)
	}
}

// RunInjection_QuotedIdentifier verifies that identifier strings containing
// SQL-injection-like characters are safely escaped by QuoteIdentifier.
//
// The column name contains a double-quote, semicolons, comment markers, and
// spaces — all characters that could break unquoted SQL. After QuoteIdentifier
// the whole string is treated as a single identifier by the database.
func RunInjection_QuotedIdentifier(t testing.TB, conn Conn, s Suite) {
	t.Helper()
	ctx := context.Background()

	// Column name that exercises all major injection vectors including spaces,
	// which must also be preserved inside the quoted identifier.
	evilCol := `evil"col; DROP TABLE t; --`

	mustExec(t, conn, ctx, "create table",
		`CREATE TEMPORARY TABLE t (id INTEGER PRIMARY KEY, "evil""col; DROP TABLE t; --" TEXT NOT NULL)`)

	insertSQL := s.InsertInto("t").
		Columns("id", evilCol).
		Values(syntax.P(1), syntax.P(2)).
		ToSql()
	mustExec(t, conn, ctx, "insert", insertSQL, 1, "safe")

	selectSQL := s.From("t").
		Select(evilCol).
		Where(syntax.Eq(evilCol, syntax.P(1))).
		ToSql()

	var val string
	if err := conn.QueryRow(ctx, selectSQL, "safe").Scan(&val); err != nil {
		t.Fatalf("select from injection-like column: %v", err)
	}
	if val != "safe" {
		t.Errorf("got %q; want %q", val, "safe")
	}

	// The table must still exist, confirming no DROP was executed.
	var count int
	countSQL := s.From("t").Select("COUNT(*)").ToSql()
	if err := conn.QueryRow(ctx, countSQL).Scan(&count); err != nil {
		t.Fatalf("table no longer accessible after identifier injection attempt: %v", err)
	}
	if count != 1 {
		t.Errorf("row count = %d; want 1", count)
	}
}
