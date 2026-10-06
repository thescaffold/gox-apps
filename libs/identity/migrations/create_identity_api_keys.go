package migrations

import "github.com/awesome-goose/goose/modules/sql"

// CreateIdentityApiKeys creates IdentityApiKeys: machine credentials for a
// workspace. Only a SHA-256 of the key is stored (the key is shown once), and
// the key's environment, kind and scopes are recorded with it.
type CreateIdentityApiKeys struct{ sql.BaseMigration }

func (m *CreateIdentityApiKeys) Run(q *sql.Query) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS "IdentityApiKeys" (
			"id"           varchar(36)  PRIMARY KEY,
			"user_id"      varchar(36)  NOT NULL,
			"client_id"    varchar(255) NOT NULL,
			"workspace_id" varchar(36)  NOT NULL,
			"name"         varchar(255) NOT NULL,
			"kind"         varchar(16)  NOT NULL CHECK ("kind" IN ('secret','publishable')),
			"environment"  varchar(16)  NOT NULL CHECK ("environment" IN ('test','live')),
			"prefix"       varchar(32)  NOT NULL,
			"hash"         varchar(64)  NOT NULL,
			"scopes"       jsonb        NOT NULL DEFAULT '["*"]'::jsonb,
			"created_at"   timestamptz  NOT NULL DEFAULT NOW(),
			"expires_at"   timestamptz,
			"revoked_at"   timestamptz,
			"last_used_at" timestamptz
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS "idx_identity_api_keys_hash" ON "IdentityApiKeys" ("hash")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_api_keys_scope" ON "IdentityApiKeys" ("client_id", "workspace_id")`,
	}
	for _, s := range stmts {
		if _, err := q.Exec(s); err != nil {
			return err
		}
	}
	return nil
}
