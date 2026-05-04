package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateNotificationLogs struct{ sql.BaseMigration }

func (m *CreateNotificationLogs) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "NotificationLogs" (
			"id"           varchar(36)  NOT NULL,
			"created_at"   timestamp    DEFAULT NULL,
			"updated_at"   timestamp    DEFAULT NULL,
			"deleted_at"   timestamp    DEFAULT NULL,
			"reference"    varchar(255) NOT NULL,
			"user_id"      varchar(36)  DEFAULT NULL,
			"client_id"    varchar(255) DEFAULT NULL,
			"workspace_id" varchar(36)  DEFAULT NULL,
			"key"          varchar(255) NOT NULL,
			"subject"      varchar(255) NOT NULL,
			"channels"     jsonb        NOT NULL,
			"data"         jsonb        DEFAULT NULL,
			"message"      jsonb        DEFAULT NULL,
			"meta"         jsonb        DEFAULT NULL,
			"read_at"      timestamp    DEFAULT NULL,
			"publish_at"   timestamp    DEFAULT NULL,
			"expire_at"    timestamp    DEFAULT NULL,
			"type"         varchar(255) DEFAULT NULL,
			"priority"     varchar(255) DEFAULT NULL,
			"theme"        varchar(255) DEFAULT NULL,
			"scope"        varchar(255) DEFAULT NULL,
			"position"     varchar(255) DEFAULT NULL,
			"status"       varchar(255) DEFAULT NULL,
			"template_id"  varchar(36)  DEFAULT NULL,
			"rule_id"      varchar(36)  DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_notification_logs_created_at"   ON "NotificationLogs" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_notification_logs_deleted_at"   ON "NotificationLogs" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_notification_logs_workspace_id" ON "NotificationLogs" ("workspace_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_notification_logs_user_id"      ON "NotificationLogs" ("user_id")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
