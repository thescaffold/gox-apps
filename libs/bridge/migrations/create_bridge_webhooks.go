package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateBridgeWebhooks struct{ sql.BaseMigration }

func (m *CreateBridgeWebhooks) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "BridgeWebhooks" (
			"id"         varchar(36)  NOT NULL,
			"created_at" timestamp    DEFAULT NULL,
			"updated_at" timestamp    DEFAULT NULL,
			"deleted_at" timestamp    DEFAULT NULL,
			"license_id" varchar(36)  NOT NULL,
			"type"       varchar(255) DEFAULT NULL,
			"url"        varchar(255) DEFAULT NULL,
			"meta"       jsonb        DEFAULT NULL,
			"status"     varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_bridge_webhooks_created_at" ON "BridgeWebhooks" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_bridge_webhooks_deleted_at" ON "BridgeWebhooks" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_bridge_webhooks_license_id" ON "BridgeWebhooks" ("license_id")`,
		// Mirrors TS BridgeWebhook migration: index on status for filtering.
		`CREATE INDEX IF NOT EXISTS "idx_bridge_webhooks_status"     ON "BridgeWebhooks" ("status")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
