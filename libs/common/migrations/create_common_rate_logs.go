package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateCommonRateLogs struct{ sql.BaseMigration }

func (m *CreateCommonRateLogs) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "CommonRateLogs" (
			"id"         varchar(36)  NOT NULL,
			"created_at" timestamp    DEFAULT NULL,
			"updated_at" timestamp    DEFAULT NULL,
			"deleted_at" timestamp    DEFAULT NULL,
			"rate_id"    varchar(36)  NOT NULL,
			"value"      integer      DEFAULT NULL,
			"delta"      integer      DEFAULT NULL,
			"date"       timestamp    DEFAULT NULL,
			"status"     varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_common_rate_logs_created_at" ON "CommonRateLogs" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_common_rate_logs_deleted_at" ON "CommonRateLogs" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_common_rate_logs_rate_id"    ON "CommonRateLogs" ("rate_id")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
