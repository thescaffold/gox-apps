package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateNotificationRules struct{ sql.BaseMigration }

func (m *CreateNotificationRules) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "NotificationRules" (
			"id"         varchar(36)  NOT NULL,
			"created_at" timestamp    DEFAULT NULL,
			"updated_at" timestamp    DEFAULT NULL,
			"deleted_at" timestamp    DEFAULT NULL,
			"group"      varchar(255) NOT NULL,
			"key"        varchar(255) NOT NULL,
			"rules"      jsonb        NOT NULL,
			"status"     varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_notification_rules_created_at" ON "NotificationRules" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_notification_rules_deleted_at" ON "NotificationRules" ("deleted_at")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
