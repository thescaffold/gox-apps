package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateNotificationMessages struct{ sql.BaseMigration }

func (m *CreateNotificationMessages) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "NotificationMessages" (
			"id"           varchar(36)  NOT NULL,
			"created_at"   timestamp    DEFAULT NULL,
			"updated_at"   timestamp    DEFAULT NULL,
			"deleted_at"   timestamp    DEFAULT NULL,
			"reference"    varchar(255) DEFAULT NULL,
			"user_id"      varchar(36)  DEFAULT NULL,
			"client_id"    varchar(255) DEFAULT NULL,
			"workspace_id" varchar(36)  DEFAULT NULL,
			"key"          varchar(255) DEFAULT NULL,
			"subject"      varchar(255) DEFAULT NULL,
			"channel"      varchar(255) DEFAULT NULL,
			"message"      jsonb        DEFAULT NULL,
			"data"         jsonb        DEFAULT NULL,
			"read_at"      timestamp    DEFAULT NULL,
			"publish_at"   timestamp    DEFAULT NULL,
			"expire_at"    timestamp    DEFAULT NULL,
			"type"         varchar(255) DEFAULT NULL,
			"priority"     varchar(255) DEFAULT NULL,
			"theme"        varchar(255) DEFAULT NULL,
			"scope"        varchar(255) DEFAULT NULL,
			"position"     varchar(255) DEFAULT NULL,
			"status"       varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}
	// Idempotently add the `data` column for DBs created before it existed
	// (TS Message carries data: log.data on web sends).
	if _, err := q.Exec(`ALTER TABLE "NotificationMessages" ADD COLUMN IF NOT EXISTS "data" jsonb`); err != nil {
		return err
	}
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_notification_messages_created_at"   ON "NotificationMessages" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_notification_messages_deleted_at"   ON "NotificationMessages" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_notification_messages_workspace_id" ON "NotificationMessages" ("workspace_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_notification_messages_user_id"      ON "NotificationMessages" ("user_id")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
