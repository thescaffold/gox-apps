package migrations

import "github.com/awesome-goose/goose/modules/sql"

// AddLinkToNotificationMessages gives each in-app message the place it opens
// (link) and indexes the inbox's queries: a person's messages newest first,
// and their unread ones. Additive and idempotent.
type AddLinkToNotificationMessages struct{ sql.BaseMigration }

func (m *AddLinkToNotificationMessages) Run(q *sql.Query) error {
	for _, stmt := range []string{
		`ALTER TABLE "NotificationMessages" ADD COLUMN IF NOT EXISTS "link" varchar(2048) DEFAULT NULL`,
		`CREATE INDEX IF NOT EXISTS "idx_notification_messages_inbox" ON "NotificationMessages" ("user_id", "channel", "created_at" DESC, "id" DESC)`,
		`CREATE INDEX IF NOT EXISTS "idx_notification_messages_unread" ON "NotificationMessages" ("user_id", "channel") WHERE "read_at" IS NULL AND "deleted_at" IS NULL`,
	} {
		if _, err := q.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}
