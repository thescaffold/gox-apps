package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateIdentityDeviceSessions struct{ sql.BaseMigration }

func (m *CreateIdentityDeviceSessions) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "IdentityDeviceSessions" (
			"id"         varchar(36)  NOT NULL,
			"created_at" timestamp    DEFAULT NULL,
			"updated_at" timestamp    DEFAULT NULL,
			"deleted_at" timestamp    DEFAULT NULL,
			"device_id"  varchar(36)  NOT NULL,
			"user_id"    varchar(36)  NOT NULL,
			"token"      text         NOT NULL,
			"expires_at" timestamp    DEFAULT NULL,
			"status"     varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_identity_device_sessions_created_at" ON "IdentityDeviceSessions" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_device_sessions_deleted_at" ON "IdentityDeviceSessions" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_device_sessions_device_id"  ON "IdentityDeviceSessions" ("device_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_device_sessions_user_id"    ON "IdentityDeviceSessions" ("user_id")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
