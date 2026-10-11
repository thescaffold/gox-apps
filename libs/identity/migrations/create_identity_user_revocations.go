package migrations

import "github.com/awesome-goose/goose/modules/sql"

// CreateIdentityUserRevocations holds, per person, the instant before which
// their access tokens are no longer accepted (sign out everywhere, account
// closure).
type CreateIdentityUserRevocations struct{ sql.BaseMigration }

func (m *CreateIdentityUserRevocations) Run(q *sql.Query) error {
	_, err := q.Exec(`CREATE TABLE IF NOT EXISTS "IdentityUserRevocations" (
		"user_id"        varchar(36)  PRIMARY KEY,
		"valid_after_ms" bigint       NOT NULL,
		"reason"         varchar(32)  NOT NULL DEFAULT '',
		"created_at"     timestamptz  NOT NULL DEFAULT NOW()
	)`)
	return err
}
