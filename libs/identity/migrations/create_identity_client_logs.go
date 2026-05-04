package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateIdentityClientLogs struct{ sql.BaseMigration }

func (m *CreateIdentityClientLogs) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "IdentityClientLogs" (
			"id"          varchar(36)  NOT NULL,
			"created_at"  timestamp    DEFAULT NULL,
			"updated_at"  timestamp    DEFAULT NULL,
			"deleted_at"  timestamp    DEFAULT NULL,
			"client_id"   varchar(255) NOT NULL,
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
		`CREATE INDEX IF NOT EXISTS "idx_identity_client_logs_created_at" ON "IdentityClientLogs" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_client_logs_deleted_at" ON "IdentityClientLogs" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_client_logs_client_id"  ON "IdentityClientLogs" ("client_id")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
