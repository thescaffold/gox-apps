package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateIdentityProviders struct{ sql.BaseMigration }

func (m *CreateIdentityProviders) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "IdentityProviders" (
			"id"          varchar(36)  NOT NULL,
			"created_at"  timestamp    DEFAULT NULL,
			"updated_at"  timestamp    DEFAULT NULL,
			"deleted_at"  timestamp    DEFAULT NULL,
			"user_id"     varchar(36)  NOT NULL,
			"client_id"   varchar(255) DEFAULT NULL,
			"type"        varchar(255) NOT NULL,
			"reference"   varchar(255) NOT NULL,
			"email"       varchar(255) DEFAULT NULL,
			"token"       text         DEFAULT NULL,
			"meta"        jsonb        DEFAULT NULL,
			"status"      varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_identity_providers_created_at" ON "IdentityProviders" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_providers_deleted_at" ON "IdentityProviders" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_providers_user_id"    ON "IdentityProviders" ("user_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_providers_type"       ON "IdentityProviders" ("type")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_providers_reference"  ON "IdentityProviders" ("reference")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
