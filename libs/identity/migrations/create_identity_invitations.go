package migrations

import "github.com/awesome-goose/goose/modules/sql"

// CreateIdentityInvitations holds invitations into a workspace (pkg/invitation):
// by email (single use, token kept as a hash) or the workspace's shareable
// link (token also sealed so admins can copy it again).
type CreateIdentityInvitations struct{ sql.BaseMigration }

func (m *CreateIdentityInvitations) Run(q *sql.Query) error {
	for _, s := range []string{
		`CREATE TABLE IF NOT EXISTS "IdentityInvitations" (
			"id"           varchar(36)  PRIMARY KEY,
			"workspace_id" varchar(36)  NOT NULL,
			"client_id"    varchar(255) NOT NULL DEFAULT '',
			"kind"         varchar(8)   NOT NULL,
			"email"        varchar(320) NOT NULL DEFAULT '',
			"seat"         varchar(32)  NOT NULL,
			"token_hash"   varchar(64)  NOT NULL,
			"token_sealed" text,
			"invited_by"   varchar(36)  NOT NULL,
			"uses"         integer      NOT NULL DEFAULT 0,
			"created_at"   timestamptz  NOT NULL DEFAULT NOW(),
			"expires_at"   timestamptz  NOT NULL,
			"accepted_at"  timestamptz,
			"accepted_by"  varchar(36),
			"revoked_at"   timestamptz,
			CONSTRAINT "chk_identity_invitations_kind" CHECK ("kind" IN ('email','link'))
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS "uq_identity_invitations_token" ON "IdentityInvitations" ("token_hash")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_invitations_workspace" ON "IdentityInvitations" ("workspace_id", "kind", "created_at" DESC)`,
	} {
		if _, err := q.Exec(s); err != nil {
			return err
		}
	}
	return nil
}
