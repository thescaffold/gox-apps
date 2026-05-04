package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateIdentityAttributes struct{ sql.BaseMigration }

func (m *CreateIdentityAttributes) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "IdentityAttributes" (
			"id"           varchar(36)  NOT NULL,
			"created_at"   timestamp    DEFAULT NULL,
			"updated_at"   timestamp    DEFAULT NULL,
			"deleted_at"   timestamp    DEFAULT NULL,
			"user_id"      varchar(36)  NOT NULL,
			"client_id"    varchar(255) DEFAULT NULL,
			"workspace_id" varchar(36)  DEFAULT NULL,
			"key"          varchar(255) NOT NULL,
			"value"        varchar(255) DEFAULT NULL,
			"type"         varchar(255) DEFAULT NULL,
			"meta"         jsonb        DEFAULT NULL,
			"status"       varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_identity_attributes_created_at"   ON "IdentityAttributes" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_attributes_deleted_at"   ON "IdentityAttributes" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_attributes_user_id"      ON "IdentityAttributes" ("user_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_attributes_workspace_id" ON "IdentityAttributes" ("workspace_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_attributes_key"          ON "IdentityAttributes" ("key")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
