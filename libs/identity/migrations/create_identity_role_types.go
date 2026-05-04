package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateIdentityRoleTypes struct{ sql.BaseMigration }

func (m *CreateIdentityRoleTypes) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "IdentityRoleTypes" (
			"id"         varchar(36)  NOT NULL,
			"created_at" timestamp    DEFAULT NULL,
			"updated_at" timestamp    DEFAULT NULL,
			"deleted_at" timestamp    DEFAULT NULL,
			"name"       varchar(255) NOT NULL,
			"desc"       varchar(255) DEFAULT NULL,
			"status"     varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_identity_role_types_created_at" ON "IdentityRoleTypes" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_role_types_deleted_at" ON "IdentityRoleTypes" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_role_types_name"       ON "IdentityRoleTypes" ("name")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
