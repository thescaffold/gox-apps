package migrations

import "github.com/awesome-goose/goose/modules/sql"

// CreateIdentityOtps stores one-time sign-in codes durably (the cache is
// per-process, which breaks with more than one server). Only a hash of the code
// is kept, with an attempt counter and expiry.
type CreateIdentityOtps struct{ sql.BaseMigration }

func (m *CreateIdentityOtps) Run(q *sql.Query) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS "IdentityOtps" (
			"id"          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			"ref"         varchar(255) NOT NULL,
			"client_id"   varchar(255) NOT NULL,
			"code_hash"   varchar(64)  NOT NULL,
			"attempts"    int          NOT NULL DEFAULT 0,
			"created_at"  timestamptz  NOT NULL DEFAULT NOW(),
			"expires_at"  timestamptz  NOT NULL,
			"consumed_at" timestamptz
		)`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_otps_ref" ON "IdentityOtps" ("ref", "client_id", "id" DESC)`,
	}
	for _, s := range stmts {
		if _, err := q.Exec(s); err != nil {
			return err
		}
	}
	return nil
}
