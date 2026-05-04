package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateHealthSummaries struct{ sql.BaseMigration }

func (m *CreateHealthSummaries) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "HealthSummaries" (
			"id"         varchar(36)  NOT NULL,
			"created_at" timestamp    DEFAULT NULL,
			"updated_at" timestamp    DEFAULT NULL,
			"deleted_at" timestamp    DEFAULT NULL,
			"service_id" varchar(36)  NOT NULL,
			"type"       varchar(255) NOT NULL,
			"received"   int          NOT NULL DEFAULT 0,
			"measure"    real         NOT NULL DEFAULT 0,
			"note"       varchar(500) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_health_summaries_created_at"  ON "HealthSummaries" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_health_summaries_service_id"  ON "HealthSummaries" ("service_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_health_summaries_type"        ON "HealthSummaries" ("type")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
