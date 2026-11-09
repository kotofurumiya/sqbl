package shared

import (
	"testing"

	"github.com/kotofurumiya/sqbl/syntax"
)

// runIdentifiers reads and writes a table whose name and columns only work
// when quoted: a reserved word, a mixed-case name, a name holding the quote
// character, and a name shaped like an injection payload.
func runIdentifiers(t *testing.T, conn Conn, s Suite) {
	const (
		mixed  = "MixedCase"
		quoted = `my"col`
		evil   = `evil"col; DROP TABLE "order"; --`
	)
	// Raw DDL uses standard double quotes; MySQL runs with ANSI_QUOTES for this.
	MustExec(t, conn, `
		CREATE TEMPORARY TABLE "order" (
			id                                    INTEGER PRIMARY KEY,
			"select"                              VARCHAR(16) NOT NULL,
			"MixedCase"                           VARCHAR(16) NOT NULL,
			"my""col"                             VARCHAR(16) NOT NULL,
			"evil""col; DROP TABLE ""order""; --" VARCHAR(16) NOT NULL
		)
	`)
	MustExec(t, conn,
		s.InsertInto("order").
			Columns("id", "select", mixed, quoted, evil).
			Values(syntax.P(1), syntax.P(2), syntax.P(3), syntax.P(4), syntax.P(5)).
			MustSQL(),
		1, "a", "b", "c", "d",
	)

	// Qualified names exercise per-part quoting of "o"."column".
	sql := s.From(syntax.As("order", "o")).
		Select("o.select", "o."+mixed, "o."+quoted, "o."+evil).
		Where(syntax.And(syntax.Eq("o.select", syntax.P(1)), syntax.Eq("o."+evil, syntax.P(2)))).
		MustSQL()

	var got [4]string
	MustScan(t, conn, sql, []any{"a", "d"}, &got[0], &got[1], &got[2], &got[3])
	if want := [4]string{"a", "b", "c", "d"}; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

// runInjectionValues stores classic injection payloads through bound
// parameters and proves each is kept and matched verbatim: quotes do not
// close, comments do not truncate, stacked statements do not run, and LIKE
// wildcards stay literal under Eq.
func runInjectionValues(t *testing.T, conn Conn, s Suite) {
	MustExec(t, conn, `CREATE TEMPORARY TABLE notes (id INTEGER PRIMARY KEY, content VARCHAR(128) NOT NULL)`)

	payloads := []string{
		"'; DROP TABLE notes; --",
		"' OR '1'='1",
		"' OR '1'='1' /*",
		"admin'--",
		"' UNION SELECT 1 --",
		"'; INSERT INTO notes (id, content) VALUES (999, 'x'); --",
		`\'; DROP TABLE notes; --`,
		"%",
		"_",
		"go%lang",
	}
	insert := s.InsertInto("notes").Columns("id", "content").Values(syntax.P(1), syntax.P(2)).MustSQL()
	for i, p := range payloads {
		MustExec(t, conn, insert, i+1, p)
	}

	all := s.From("notes").Select("content").OrderBy(syntax.Asc("id")).MustSQL()
	AssertSlice(t, QueryColumn[string](t, conn, all), payloads)

	match := s.From("notes").Select("id").Where(syntax.Eq("content", syntax.P(1))).MustSQL()
	for i, p := range payloads {
		AssertSlice(t, QueryColumn[int](t, conn, match, p), []int{i + 1})
	}

	// Values that cannot be written safely as SQL must stop the render instead
	// of reaching the database. A named string type used to be pasted in as-is.
	t.Run("RefusedBeforeExecution", func(t *testing.T) {
		type role string
		refused := []struct {
			name string
			b    interface{ ToSQL() (string, error) }
		}{
			{"named string value", s.From("notes").Where(syntax.Eq("content", role("' OR '1'='1")))},
			{"placeholder name", s.From("notes").Where(syntax.Eq("content", syntax.P("x' OR '1'='1")))},
		}
		for _, tt := range refused {
			if sql, err := tt.b.ToSQL(); err == nil {
				t.Errorf("%s: rendered %q; want an error", tt.name, sql)
			}
		}
		AssertCount(t, conn, "notes", len(payloads))
	})
}

// runInjectionIdentifiers passes a payload as an identifier in each position
// that takes one. Quoting turns it into an unknown name, so the database
// either rejects the query or (SQLite) reads it as a string literal; either
// way the embedded DROP must never run. An alias is always valid, so that
// query must succeed.
func runInjectionIdentifiers(t *testing.T, conn Conn, s Suite) {
	MustExec(t, conn, `CREATE TEMPORARY TABLE victims (id INTEGER PRIMARY KEY)`)
	MustExec(t, conn, s.InsertInto("victims").Columns("id").Values(1).MustSQL())

	const payload = "id; DROP TABLE victims; --"
	unknown := []struct{ name, sql string }{
		{"From", s.From(payload).Select("id").MustSQL()},
		{"Select", s.From("victims").Select(payload).MustSQL()},
		{"Where", s.From("victims").Select("id").Where(syntax.Eq(payload, 1)).MustSQL()},
		{"OrderBy", s.From("victims").Select("id").OrderBy(syntax.Asc(payload)).MustSQL()},
	}
	for _, tt := range unknown {
		t.Run(tt.name, func(t *testing.T) {
			if rows, err := conn.Query(t.Context(), tt.sql); err == nil {
				rows.Close()
			}
			AssertCount(t, conn, "victims", 1)
		})
	}

	t.Run("Alias", func(t *testing.T) {
		sql := s.From("victims").Select(syntax.As("id", payload)).MustSQL()
		AssertSlice(t, QueryColumn[int](t, conn, sql), []int{1})
		AssertCount(t, conn, "victims", 1)
	})
}
