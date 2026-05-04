package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateHealthLogs struct{ sql.BaseMigration }

func (m *CreateHealthLogs) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "HealthLogs" (
			"id"         varchar(36)  NOT NULL,
			"created_at" timestamp    DEFAULT NULL,
			"updated_at" timestamp    DEFAULT NULL,
			"deleted_at" timestamp    DEFAULT NULL,
			"service_id" varchar(36)  NOT NULL,
			"state"      varchar(255) DEFAULT NULL,
			"meta"       jsonb        DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_health_logs_created_at"  ON "HealthLogs" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_health_logs_service_id"  ON "HealthLogs" ("service_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_health_logs_state"       ON "HealthLogs" ("state")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
