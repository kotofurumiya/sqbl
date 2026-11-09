package builder_test

import (
	"testing"

	"github.com/kotofurumiya/sqbl/syntax"
)

func TestDeleteBuilder_Clauses(t *testing.T) {
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
			// Each Where narrows the delete further; Returning is replaced.
			name: "where accumulates with and while returning is replaced",
			b: sessions.
				Where(syntax.Eq("user_id", syntax.P(1))).Where(nil).Where(syntax.Lt("expires_at", syntax.P(2))).
				Returning("old").Returning("*"),
			want: `DELETE FROM "sessions" WHERE "user_id" = $1 AND "expires_at" < $2 RETURNING *;`,
		},
	})
}
