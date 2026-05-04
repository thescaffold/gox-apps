package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateIdentityDeviceLogs struct{ sql.BaseMigration }

func (m *CreateIdentityDeviceLogs) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "IdentityDeviceLogs" (
			"id"         varchar(36)  NOT NULL,
			"created_at" timestamp    DEFAULT NULL,
			"updated_at" timestamp    DEFAULT NULL,
			"deleted_at" timestamp    DEFAULT NULL,
			"device_id"  varchar(36)  NOT NULL,
			"event"      varchar(255) NOT NULL,
			"meta"       jsonb        DEFAULT NULL,
			"status"     varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_identity_device_logs_created_at" ON "IdentityDeviceLogs" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_device_logs_deleted_at" ON "IdentityDeviceLogs" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_device_logs_device_id"  ON "IdentityDeviceLogs" ("device_id")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
