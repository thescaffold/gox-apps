package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateIdentityPermissions struct{ sql.BaseMigration }

func (m *CreateIdentityPermissions) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "IdentityPermissions" (
			"id"           varchar(36)  NOT NULL,
			"created_at"   timestamp    DEFAULT NULL,
			"updated_at"   timestamp    DEFAULT NULL,
			"deleted_at"   timestamp    DEFAULT NULL,
			"user_id"      varchar(36)  DEFAULT NULL,
			"client_id"    varchar(255) DEFAULT NULL,
			"workspace_id" varchar(36)  DEFAULT NULL,
			"role_id"      varchar(36)  DEFAULT NULL,
			"key"          varchar(255) NOT NULL,
			"scope"        varchar(255) NOT NULL,
			"resource"     varchar(255) NOT NULL,
			"action"       varchar(255) NOT NULL,
			"status"       varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_identity_permissions_created_at"   ON "IdentityPermissions" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_permissions_deleted_at"   ON "IdentityPermissions" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_permissions_user_id"      ON "IdentityPermissions" ("user_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_permissions_workspace_id" ON "IdentityPermissions" ("workspace_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_permissions_role_id"      ON "IdentityPermissions" ("role_id")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
