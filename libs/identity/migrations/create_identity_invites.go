package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateIdentityInvites struct{ sql.BaseMigration }

func (m *CreateIdentityInvites) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "IdentityInvites" (
			"id"           varchar(36)  NOT NULL,
			"created_at"   timestamp    DEFAULT NULL,
			"updated_at"   timestamp    DEFAULT NULL,
			"deleted_at"   timestamp    DEFAULT NULL,
			"user_id"      varchar(36)  NOT NULL,
			"client_id"    varchar(255) NOT NULL,
			"workspace_id" varchar(36)  NOT NULL,
			"email"        varchar(255) NOT NULL,
			"role_id"      varchar(36)  DEFAULT NULL,
			"token"        varchar(255) NOT NULL,
			"expires_at"   timestamp    DEFAULT NULL,
			"status"       varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_identity_invites_created_at"   ON "IdentityInvites" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_invites_deleted_at"   ON "IdentityInvites" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_invites_user_id"      ON "IdentityInvites" ("user_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_invites_workspace_id" ON "IdentityInvites" ("workspace_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_invites_email"        ON "IdentityInvites" ("email")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
