package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateIdentityUserClientWorkspaces struct{ sql.BaseMigration }

func (m *CreateIdentityUserClientWorkspaces) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "IdentityUserClientWorkspaces" (
			"id"           varchar(36)  NOT NULL,
			"created_at"   timestamp    DEFAULT NULL,
			"updated_at"   timestamp    DEFAULT NULL,
			"deleted_at"   timestamp    DEFAULT NULL,
			"user_id"      varchar(36)  NOT NULL,
			"client_id"    varchar(255) NOT NULL,
			"workspace_id" varchar(36)  NOT NULL,
			"role_id"      varchar(36)  DEFAULT NULL,
			"status"       varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_identity_ucw_created_at"   ON "IdentityUserClientWorkspaces" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_ucw_deleted_at"   ON "IdentityUserClientWorkspaces" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_ucw_user_id"      ON "IdentityUserClientWorkspaces" ("user_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_ucw_client_id"    ON "IdentityUserClientWorkspaces" ("client_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_ucw_workspace_id" ON "IdentityUserClientWorkspaces" ("workspace_id")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
