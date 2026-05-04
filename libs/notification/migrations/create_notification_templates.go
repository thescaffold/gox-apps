package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateNotificationTemplates struct{ sql.BaseMigration }

func (m *CreateNotificationTemplates) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "NotificationTemplates" (
			"id"         varchar(36)  NOT NULL,
			"created_at" timestamp    DEFAULT NULL,
			"updated_at" timestamp    DEFAULT NULL,
			"deleted_at" timestamp    DEFAULT NULL,
			"group"      varchar(255) NOT NULL,
			"key"        varchar(255) NOT NULL,
			"email"      text         DEFAULT NULL,
			"sms"        text         DEFAULT NULL,
			"web"        text         DEFAULT NULL,
			"mobile"     text         DEFAULT NULL,
			"status"     varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_notification_templates_created_at" ON "NotificationTemplates" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_notification_templates_deleted_at" ON "NotificationTemplates" ("deleted_at")`,
		`CREATE UNIQUE INDEX IF NOT EXISTS "idx_notification_templates_group_key" ON "NotificationTemplates" ("group","key")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
