package migrations

import "github.com/awesome-goose/goose/modules/sql"

// CreateIdentitySsoConnections holds each workspace's OpenID Connect
// registration: the provider, the client credentials (the secret is sealed,
// never stored in clear) and the email domains the workspace has claimed.
// One connection per workspace.
type CreateIdentitySsoConnections struct{ sql.BaseMigration }

func (m *CreateIdentitySsoConnections) Run(q *sql.Query) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS "IdentitySsoConnections" (
			"id"                   varchar(36)  PRIMARY KEY,
			"workspace_id"         varchar(36)  NOT NULL,
			"issuer"               text         NOT NULL,
			"oidc_client_id"       text         NOT NULL,
			"client_secret_sealed" text         NOT NULL,
			"domains"              jsonb        NOT NULL DEFAULT '[]',
			"enforce"              boolean      NOT NULL DEFAULT false,
			"created_at"           timestamptz  NOT NULL DEFAULT NOW(),
			"updated_at"           timestamptz  NOT NULL DEFAULT NOW()
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS "idx_identity_sso_workspace" ON "IdentitySsoConnections" ("workspace_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_sso_domains" ON "IdentitySsoConnections" USING gin ("domains")`,
	}
	for _, s := range stmts {
		if _, err := q.Exec(s); err != nil {
			return err
		}
	}
	return nil
}
