package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreatePolylogSinkTypes struct{ sql.BaseMigration }

func (m *CreatePolylogSinkTypes) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "PolylogSinkTypes" (
			"id"            varchar(36)  NOT NULL,
			"created_at"    timestamp    DEFAULT NULL,
			"updated_at"    timestamp    DEFAULT NULL,
			"deleted_at"    timestamp    DEFAULT NULL,
			"category"      varchar(255) NOT NULL,
			"key"           varchar(255) DEFAULT NULL,
			"visibility"    varchar(255) DEFAULT NULL,
			"name"          varchar(255) NOT NULL,
			"desc"          varchar(255) DEFAULT NULL,
			"detail"        text         DEFAULT NULL,
			"thumbnail_url" varchar(255) DEFAULT NULL,
			"banner_url"    varchar(255) DEFAULT NULL,
			"tags"          jsonb        DEFAULT NULL,
			"meta"          jsonb        DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_polylog_sink_types_created_at" ON "PolylogSinkTypes" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_polylog_sink_types_deleted_at" ON "PolylogSinkTypes" ("deleted_at")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
