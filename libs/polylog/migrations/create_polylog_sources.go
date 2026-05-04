package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreatePolylogSources struct{ sql.BaseMigration }

func (m *CreatePolylogSources) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "PolylogSources" (
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
		`CREATE INDEX IF NOT EXISTS "idx_polylog_sources_created_at"   ON "PolylogSources" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_polylog_sources_deleted_at"   ON "PolylogSources" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_polylog_sources_workspace_id" ON "PolylogSources" ("workspace_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_polylog_sources_type_id"      ON "PolylogSources" ("type_id")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
