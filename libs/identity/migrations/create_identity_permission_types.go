package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateIdentityPermissionTypes struct{ sql.BaseMigration }

func (m *CreateIdentityPermissionTypes) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "IdentityPermissionTypes" (
			"id"         varchar(36)  NOT NULL,
			"created_at" timestamp    DEFAULT NULL,
			"updated_at" timestamp    DEFAULT NULL,
			"deleted_at" timestamp    DEFAULT NULL,
			"name"       varchar(255) NOT NULL,
			"key"        varchar(255) NOT NULL,
			"desc"       varchar(255) DEFAULT NULL,
			"resource"   varchar(255) NOT NULL,
			"action"     varchar(255) NOT NULL,
			"scope"      varchar(255) NOT NULL,
			"status"     varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_identity_permission_types_created_at" ON "IdentityPermissionTypes" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_permission_types_deleted_at" ON "IdentityPermissionTypes" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_permission_types_key"        ON "IdentityPermissionTypes" ("key")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
