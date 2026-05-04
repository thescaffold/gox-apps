package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreatePolylogChannels struct{ sql.BaseMigration }

func (m *CreatePolylogChannels) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "PolylogChannels" (
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
			"type"         varchar(255) DEFAULT NULL,
			"source_id"    varchar(36)  NOT NULL,
			"sink_id"      varchar(36)  DEFAULT NULL,
			"tags"         jsonb        DEFAULT NULL,
			"meta"         jsonb        DEFAULT NULL,
			"status"       varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_polylog_channels_created_at"   ON "PolylogChannels" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_polylog_channels_deleted_at"   ON "PolylogChannels" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_polylog_channels_workspace_id" ON "PolylogChannels" ("workspace_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_polylog_channels_source_id"    ON "PolylogChannels" ("source_id")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
