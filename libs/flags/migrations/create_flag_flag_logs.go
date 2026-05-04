package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateFlagFlagLogs struct{ sql.BaseMigration }

func (m *CreateFlagFlagLogs) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "FlagFlagLogs" (
			"id"         varchar(36)  NOT NULL,
			"created_at" timestamp    DEFAULT NULL,
			"updated_at" timestamp    DEFAULT NULL,
			"deleted_at" timestamp    DEFAULT NULL,
			"flag_id"    varchar(36)  NOT NULL,
			"limit"      int          NOT NULL DEFAULT 0,
			"meta"       jsonb        DEFAULT NULL,
			"status"     varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_flag_flag_logs_created_at" ON "FlagFlagLogs" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_flag_flag_logs_deleted_at" ON "FlagFlagLogs" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_flag_flag_logs_flag_id"    ON "FlagFlagLogs" ("flag_id")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
