package migrations

import "github.com/awesome-goose/goose/modules/sql"

// CreateIdentitySsoRecoveries backs the break-glass for companies that
// require single sign-on: someone who controls the company's DNS can prove it
// with a TXT record and be allowed one ordinary sign-in. Only a hash of the
// proof token is kept.
type CreateIdentitySsoRecoveries struct{ sql.BaseMigration }

func (m *CreateIdentitySsoRecoveries) Run(q *sql.Query) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS "IdentitySsoRecoveries" (
			"id"            varchar(36)  PRIMARY KEY,
			"email"         varchar(320) NOT NULL,
			"domain"        varchar(253) NOT NULL,
			"token_hash"    varchar(64)  NOT NULL,
			"expires_at"    timestamptz  NOT NULL,
			"granted_until" timestamptz,
			"created_at"    timestamptz  NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_sso_recoveries_email" ON "IdentitySsoRecoveries" ("email", "created_at" DESC)`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_sso_recoveries_domain" ON "IdentitySsoRecoveries" ("domain", "created_at" DESC)`,
	}
	for _, s := range stmts {
		if _, err := q.Exec(s); err != nil {
			return err
		}
	}
	return nil
}
