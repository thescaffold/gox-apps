package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateHealthServices struct{ sql.BaseMigration }

func (m *CreateHealthServices) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "HealthServices" (
			"id"         varchar(36)  NOT NULL,
			"created_at" timestamp    DEFAULT NULL,
			"updated_at" timestamp    DEFAULT NULL,
			"deleted_at" timestamp    DEFAULT NULL,
			"name"       varchar(255) NOT NULL,
			"desc"       varchar(255) DEFAULT NULL,
			"type"       varchar(255) DEFAULT NULL,
			"state"      varchar(255) DEFAULT NULL,
			"status"     varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_health_services_created_at" ON "HealthServices" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_health_services_name"       ON "HealthServices" ("name")`,
		`CREATE INDEX IF NOT EXISTS "idx_health_services_state"      ON "HealthServices" ("state")`,
		// Mirrors TS HealthServices migration: indexes on type and status.
		`CREATE INDEX IF NOT EXISTS "idx_health_services_type"       ON "HealthServices" ("type")`,
		`CREATE INDEX IF NOT EXISTS "idx_health_services_status"     ON "HealthServices" ("status")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
