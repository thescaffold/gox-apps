package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateIdentityProviderLogs struct{ sql.BaseMigration }

func (m *CreateIdentityProviderLogs) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "IdentityProviderLogs" (
			"id"          varchar(36)  NOT NULL,
			"created_at"  timestamp    DEFAULT NULL,
			"updated_at"  timestamp    DEFAULT NULL,
			"deleted_at"  timestamp    DEFAULT NULL,
			"provider_id" varchar(36)  NOT NULL,
			"event"       varchar(255) NOT NULL,
			"request"     jsonb        DEFAULT NULL,
			"response"    jsonb        DEFAULT NULL,
			"meta"        jsonb        DEFAULT NULL,
			"status"      varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_identity_provider_logs_created_at"  ON "IdentityProviderLogs" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_provider_logs_deleted_at"  ON "IdentityProviderLogs" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_provider_logs_provider_id" ON "IdentityProviderLogs" ("provider_id")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
