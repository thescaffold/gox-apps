package migrations

import "github.com/awesome-goose/goose/modules/sql"

// SsoSamlSigning gives each workspace's SAML service provider its own signing key (V4-03): the
// current key and certificate, and a next one a rotation stages before it signs. The key is
// sealed like a provider secret. A connection says whether its AuthnRequests are signed.
type SsoSamlSigning struct{ sql.BaseMigration }

func (m *SsoSamlSigning) Run(q *sql.Query) error {
	for _, s := range []string{
		`ALTER TABLE "IdentitySsoConnections" ADD COLUMN IF NOT EXISTS "sign_requests" boolean NOT NULL DEFAULT false`,
		`CREATE TABLE IF NOT EXISTS "IdentitySsoSpKeys" (
			"workspace_id"    varchar(36)  PRIMARY KEY,
			"key_sealed"      text         NOT NULL,
			"cert"            text         NOT NULL,
			"next_key_sealed" text         NOT NULL DEFAULT '',
			"next_cert"       text         NOT NULL DEFAULT '',
			"created_at"      timestamptz  NOT NULL DEFAULT NOW(),
			"updated_at"      timestamptz  NOT NULL DEFAULT NOW(),
			"rotated_at"      timestamptz  DEFAULT NULL
		)`,
	} {
		if _, err := q.Exec(s); err != nil {
			return err
		}
	}
	return nil
}
