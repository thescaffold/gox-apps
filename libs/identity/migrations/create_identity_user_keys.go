package migrations

import "github.com/awesome-goose/goose/modules/sql"

// CreateIdentityUserKeys holds a person's public keys (U-S22, pkg/userkey):
// SSH keys that authenticate git over SSH, and SSH or GPG keys that sign
// commits. Only public keys are kept. A key belongs to one person across the
// whole service for each use: git over SSH tells who is connecting by it, and a
// signature verifies as one person's.
type CreateIdentityUserKeys struct{ sql.BaseMigration }

func (m *CreateIdentityUserKeys) Run(q *sql.Query) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS "IdentityUserKeys" (
			"id"           varchar(36)  PRIMARY KEY,
			"user_id"      varchar(36)  NOT NULL,
			"usage"        varchar(16)  NOT NULL,
			"format"       varchar(8)   NOT NULL,
			"title"        varchar(120) NOT NULL,
			"public_key"   text         NOT NULL,
			"fingerprint"  varchar(128) NOT NULL,
			"algorithm"    varchar(64)  NOT NULL DEFAULT '',
			"bits"         integer      NOT NULL DEFAULT 0,
			"key_id"       varchar(40)  NOT NULL DEFAULT '',
			"emails"       jsonb,
			"expires_at"   timestamptz,
			"last_used_at" timestamptz,
			"created_at"   timestamptz  NOT NULL DEFAULT NOW(),
			CONSTRAINT "chk_identity_user_keys_usage" CHECK ("usage" IN ('auth','signing')),
			CONSTRAINT "chk_identity_user_keys_format" CHECK ("format" IN ('ssh','gpg')),
			CONSTRAINT "chk_identity_user_keys_auth_ssh" CHECK ("usage" <> 'auth' OR "format" = 'ssh')
		)`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_user_keys_user" ON "IdentityUserKeys" ("user_id", "usage")`,
		`CREATE UNIQUE INDEX IF NOT EXISTS "uq_identity_user_keys_usage_fingerprint" ON "IdentityUserKeys" ("usage", "fingerprint")`,
	}
	for _, s := range stmts {
		if _, err := q.Exec(s); err != nil {
			return err
		}
	}
	return nil
}
