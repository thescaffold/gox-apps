package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreatePolylogEvents struct{ sql.BaseMigration }

func (m *CreatePolylogEvents) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "PolylogEvents" (
			"id"           varchar(36)  NOT NULL,
			"created_at"   timestamp    DEFAULT NULL,
			"updated_at"   timestamp    DEFAULT NULL,
			"deleted_at"   timestamp    DEFAULT NULL,
			"user_id"      varchar(36)  NOT NULL,
			"client_id"    varchar(255) NOT NULL,
			"workspace_id" varchar(36)  NOT NULL,
			"entity_id"    varchar(36)  NOT NULL,
			"entity_name"  varchar(255) NOT NULL,
			"category"     varchar(255) NOT NULL,
			"type"         varchar(255) DEFAULT NULL,
			"reference"    varchar(255) NOT NULL,
			"version"      varchar(255) NOT NULL,
			"payload"      jsonb        NOT NULL,
			"status"       varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_polylog_events_created_at"   ON "PolylogEvents" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_polylog_events_deleted_at"   ON "PolylogEvents" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_polylog_events_workspace_id" ON "PolylogEvents" ("workspace_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_polylog_events_entity_name"  ON "PolylogEvents" ("entity_name")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
