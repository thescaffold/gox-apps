package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreatePolylogEventLogs struct{ sql.BaseMigration }

func (m *CreatePolylogEventLogs) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "PolylogEventLogs" (
			"id"         varchar(36) NOT NULL,
			"created_at" timestamp   DEFAULT NULL,
			"updated_at" timestamp   DEFAULT NULL,
			"deleted_at" timestamp   DEFAULT NULL,
			"event_id"   varchar(36) NOT NULL,
			"meta"       jsonb       DEFAULT NULL,
			"request"    jsonb       DEFAULT NULL,
			"response"   jsonb       DEFAULT NULL,
			"status"     varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_polylog_event_logs_created_at" ON "PolylogEventLogs" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_polylog_event_logs_updated_at" ON "PolylogEventLogs" ("updated_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_polylog_event_logs_deleted_at" ON "PolylogEventLogs" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_polylog_event_logs_event_id"   ON "PolylogEventLogs" ("event_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_polylog_event_logs_status"     ON "PolylogEventLogs" ("status")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
