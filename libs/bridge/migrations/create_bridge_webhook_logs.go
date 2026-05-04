package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateBridgeWebhookLogs struct{ sql.BaseMigration }

func (m *CreateBridgeWebhookLogs) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "BridgeWebhookLogs" (
			"id"         varchar(36)  NOT NULL,
			"created_at" timestamp    DEFAULT NULL,
			"updated_at" timestamp    DEFAULT NULL,
			"deleted_at" timestamp    DEFAULT NULL,
			"webhook_id" varchar(36)  NOT NULL,
			"type"       varchar(255) DEFAULT NULL,
			"meta"       jsonb        DEFAULT NULL,
			"request"    jsonb        DEFAULT NULL,
			"response"   jsonb        DEFAULT NULL,
			"status"     varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_bridge_webhook_logs_created_at" ON "BridgeWebhookLogs" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_bridge_webhook_logs_deleted_at" ON "BridgeWebhookLogs" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_bridge_webhook_logs_webhook_id" ON "BridgeWebhookLogs" ("webhook_id")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
