package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateIdentityWorkspaces struct{ sql.BaseMigration }

func (m *CreateIdentityWorkspaces) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "IdentityWorkspaces" (
			"id"         varchar(36)  NOT NULL,
			"created_at" timestamp    DEFAULT NULL,
			"updated_at" timestamp    DEFAULT NULL,
			"deleted_at" timestamp    DEFAULT NULL,
			"user_id"    varchar(36)  NOT NULL,
			"client_id"  varchar(255) NOT NULL,
			"name"       varchar(255) NOT NULL,
			"desc"       varchar(255) DEFAULT NULL,
			"slug"       varchar(255) NOT NULL,
			"logo_url"   varchar(255) DEFAULT NULL,
			"banner_url" varchar(255) DEFAULT NULL,
			"meta"       jsonb        DEFAULT NULL,
			"status"     varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_identity_workspaces_created_at" ON "IdentityWorkspaces" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_workspaces_deleted_at" ON "IdentityWorkspaces" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_workspaces_user_id"    ON "IdentityWorkspaces" ("user_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_workspaces_client_id"  ON "IdentityWorkspaces" ("client_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_workspaces_slug"       ON "IdentityWorkspaces" ("slug")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
