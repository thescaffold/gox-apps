package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateBridgePreferences struct{ sql.BaseMigration }

func (m *CreateBridgePreferences) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "BridgePreferences" (
			"id"         varchar(36)  NOT NULL,
			"created_at" timestamp    DEFAULT NULL,
			"updated_at" timestamp    DEFAULT NULL,
			"deleted_at" timestamp    DEFAULT NULL,
			"license_id" varchar(36)  NOT NULL,
			"key"        varchar(255) NOT NULL,
			"value"      varchar(255) DEFAULT NULL,
			"type"       varchar(255) DEFAULT NULL,
			"meta"       jsonb        DEFAULT NULL,
			"status"     varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_bridge_preferences_created_at" ON "BridgePreferences" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_bridge_preferences_deleted_at" ON "BridgePreferences" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_bridge_preferences_license_id" ON "BridgePreferences" ("license_id")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
