package builder_test

import (
	"testing"

	"github.com/kotofurumiya/sqbl/syntax"
)

func TestSqlDeleteBuilder_Clauses(t *testing.T) {
	t.Parallel()
	sessions := newDelete().From("sessions")

	runSQLCases(t, []sqlCase{
		{
			name: "without where deletes every row",
			b:    sessions,
			want: `DELETE FROM "sessions";`,
		},
		{
			name: "where and returning",
			b:    sessions.Where(syntax.Lt("expires_at", syntax.P(1))).Returning("id", "user_id"),
			want: `DELETE FROM "sessions" WHERE "expires_at" < $1 RETURNING "id", "user_id";`,
		},
		{
			name: "where and returning are replaced by later calls",
			b: sessions.
				Where(syntax.Eq("old", 1)).Where(syntax.Eq("id", syntax.P(1))).
				Returning("old").Returning("id"),
			want: `DELETE FROM "sessions" WHERE "id" = $1 RETURNING "id";`,
		},
	})
}
