package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateFussHistories struct{ sql.BaseMigration }

func (m *CreateFussHistories) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "FussHistories" (
			"id"           varchar(36)  NOT NULL,
			"created_at"   timestamp    DEFAULT NULL,
			"updated_at"   timestamp    DEFAULT NULL,
			"deleted_at"   timestamp    DEFAULT NULL,
			"user_id"      varchar(36)  NOT NULL,
			"client_id"    varchar(255) NOT NULL,
			"workspace_id" varchar(36)  NOT NULL,
			"type"         varchar(255) NOT NULL DEFAULT 'global',
			"query"        varchar(255) NOT NULL,
			"status"       varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}

	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_fuss_histories_created_at"   ON "FussHistories" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_fuss_histories_updated_at"   ON "FussHistories" ("updated_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_fuss_histories_deleted_at"   ON "FussHistories" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_fuss_histories_user_id"      ON "FussHistories" ("user_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_fuss_histories_client_id"    ON "FussHistories" ("client_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_fuss_histories_workspace_id" ON "FussHistories" ("workspace_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_fuss_histories_type"         ON "FussHistories" ("type")`,
		`CREATE INDEX IF NOT EXISTS "idx_fuss_histories_query"        ON "FussHistories" ("query")`,
		`CREATE INDEX IF NOT EXISTS "idx_fuss_histories_status"       ON "FussHistories" ("status")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}

	return nil
}
