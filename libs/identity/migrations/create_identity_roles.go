package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateIdentityRoles struct{ sql.BaseMigration }

func (m *CreateIdentityRoles) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "IdentityRoles" (
			"id"           varchar(36)  NOT NULL,
			"created_at"   timestamp    DEFAULT NULL,
			"updated_at"   timestamp    DEFAULT NULL,
			"deleted_at"   timestamp    DEFAULT NULL,
			"name"         varchar(255) NOT NULL,
			"client_id"    varchar(255) NOT NULL,
			"workspace_id" varchar(36)  DEFAULT NULL,
			"type"         varchar(255) NOT NULL,
			"status"       varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_identity_roles_created_at"   ON "IdentityRoles" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_roles_deleted_at"   ON "IdentityRoles" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_roles_client_id"    ON "IdentityRoles" ("client_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_roles_workspace_id" ON "IdentityRoles" ("workspace_id")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
