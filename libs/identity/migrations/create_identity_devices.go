package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateIdentityDevices struct{ sql.BaseMigration }

func (m *CreateIdentityDevices) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "IdentityDevices" (
			"id"          varchar(36)  NOT NULL,
			"created_at"  timestamp    DEFAULT NULL,
			"updated_at"  timestamp    DEFAULT NULL,
			"deleted_at"  timestamp    DEFAULT NULL,
			"user_id"     varchar(36)  NOT NULL,
			"fingerprint" varchar(255) NOT NULL,
			"type"        varchar(255) NOT NULL,
			"platform"    varchar(255) DEFAULT NULL,
			"meta"        jsonb        DEFAULT NULL,
			"status"      varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_identity_devices_created_at"  ON "IdentityDevices" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_devices_deleted_at"  ON "IdentityDevices" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_devices_user_id"     ON "IdentityDevices" ("user_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_devices_fingerprint" ON "IdentityDevices" ("fingerprint")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
