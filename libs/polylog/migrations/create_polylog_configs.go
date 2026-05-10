package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreatePolylogConfigs struct{ sql.BaseMigration }

func (m *CreatePolylogConfigs) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "PolylogConfigs" (
			"id"           varchar(36)  NOT NULL,
			"created_at"   timestamp    DEFAULT NULL,
			"updated_at"   timestamp    DEFAULT NULL,
			"deleted_at"   timestamp    DEFAULT NULL,
			"user_id"      varchar(36)  NOT NULL,
			"client_id"    varchar(255) NOT NULL,
			"workspace_id" varchar(36)  NOT NULL,
			"entity_id"    varchar(36)  NOT NULL,
			"entity_name"  varchar(255) NOT NULL,
			"type"         varchar(255) NOT NULL,
			"category"     varchar(255) NOT NULL,
			"value"        jsonb        DEFAULT NULL,
			"status"       varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}

	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_polylog_configs_user_id"      ON "PolylogConfigs" ("user_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_polylog_configs_workspace_id" ON "PolylogConfigs" ("workspace_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_polylog_configs_entity_id"    ON "PolylogConfigs" ("entity_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_polylog_configs_entity_name"  ON "PolylogConfigs" ("entity_name")`,
		`CREATE INDEX IF NOT EXISTS "idx_polylog_configs_type"         ON "PolylogConfigs" ("type")`,
		`CREATE INDEX IF NOT EXISTS "idx_polylog_configs_category"     ON "PolylogConfigs" ("category")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}

	return nil
}
