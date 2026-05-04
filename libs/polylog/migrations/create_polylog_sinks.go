package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreatePolylogSinks struct{ sql.BaseMigration }

func (m *CreatePolylogSinks) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "PolylogSinks" (
			"id"           varchar(36)  NOT NULL,
			"created_at"   timestamp    DEFAULT NULL,
			"updated_at"   timestamp    DEFAULT NULL,
			"deleted_at"   timestamp    DEFAULT NULL,
			"user_id"      varchar(36)  NOT NULL,
			"client_id"    varchar(255) NOT NULL,
			"workspace_id" varchar(36)  NOT NULL,
			"category"     varchar(255) NOT NULL,
			"key"          varchar(255) DEFAULT NULL,
			"visibility"   varchar(255) DEFAULT NULL,
			"type_id"      varchar(36)  NOT NULL,
			"name"         varchar(255) NOT NULL,
			"desc"         varchar(255) DEFAULT NULL,
			"meta"         jsonb        DEFAULT NULL,
			"status"       varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_polylog_sinks_created_at"   ON "PolylogSinks" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_polylog_sinks_deleted_at"   ON "PolylogSinks" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_polylog_sinks_workspace_id" ON "PolylogSinks" ("workspace_id")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
