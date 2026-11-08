package migrations

import "github.com/awesome-goose/goose/modules/sql"

// SsoSaml lets a connection speak SAML 2.0 (U-S21): its protocol (existing
// connections are OpenID Connect), the IdP metadata and the URL it came from.
// IdentitySsoSamlAssertions holds the id of every assertion accepted, until
// it expires, so none is accepted twice.
type SsoSaml struct{ sql.BaseMigration }

func (m *SsoSaml) Run(q *sql.Query) error {
	for _, s := range []string{
		`ALTER TABLE "IdentitySsoConnections" ADD COLUMN IF NOT EXISTS "protocol" varchar(8) NOT NULL DEFAULT 'oidc'`,
		`ALTER TABLE "IdentitySsoConnections" ADD COLUMN IF NOT EXISTS "saml_metadata_url" text NOT NULL DEFAULT ''`,
		`ALTER TABLE "IdentitySsoConnections" ADD COLUMN IF NOT EXISTS "saml_metadata" text NOT NULL DEFAULT ''`,
		`CREATE TABLE IF NOT EXISTS "IdentitySsoSamlAssertions" (
			"connection_id" varchar(36)  NOT NULL,
			"assertion_id"  text         NOT NULL,
			"expires_at"    timestamptz  NOT NULL,
			"created_at"    timestamptz  NOT NULL DEFAULT NOW(),
			PRIMARY KEY ("connection_id", "assertion_id")
		)`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_sso_saml_assertions_expiry" ON "IdentitySsoSamlAssertions" ("expires_at")`,
	} {
		if _, err := q.Exec(s); err != nil {
			return err
		}
	}
	return nil
}
