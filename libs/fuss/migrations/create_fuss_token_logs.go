package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateFussTokenLogs struct{ sql.BaseMigration }

func (m *CreateFussTokenLogs) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "FussTokenLogs" (
			"id"         varchar(36)  NOT NULL,
			"created_at" timestamp    DEFAULT NULL,
			"updated_at" timestamp    DEFAULT NULL,
			"deleted_at" timestamp    DEFAULT NULL,
			"token_id"   varchar(36)  NOT NULL,
			"value"      varchar(255) NOT NULL,
			"status"     varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}

	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_fuss_token_logs_created_at" ON "FussTokenLogs" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_fuss_token_logs_updated_at" ON "FussTokenLogs" ("updated_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_fuss_token_logs_deleted_at" ON "FussTokenLogs" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_fuss_token_logs_token_id"   ON "FussTokenLogs" ("token_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_fuss_token_logs_value"      ON "FussTokenLogs" ("value")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}

	return nil
}
