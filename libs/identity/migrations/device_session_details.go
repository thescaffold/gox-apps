package migrations

import "github.com/awesome-goose/goose/modules/sql"

// DeviceSessionDetails is what Account › Sessions shows of a device session and how single logout
// finds it (V0-15, V4-03): the address and place it was last used from, when, and, for a session
// opened through SAML, the connection, NameID and SessionIndex the identity provider gave it.
type DeviceSessionDetails struct{ sql.BaseMigration }

func (m *DeviceSessionDetails) Run(q *sql.Query) error {
	for _, s := range []string{
		`ALTER TABLE "IdentityDeviceSessions" ADD COLUMN IF NOT EXISTS "ip" varchar(64) NOT NULL DEFAULT ''`,
		`ALTER TABLE "IdentityDeviceSessions" ADD COLUMN IF NOT EXISTS "place" varchar(255) NOT NULL DEFAULT ''`,
		`ALTER TABLE "IdentityDeviceSessions" ADD COLUMN IF NOT EXISTS "last_seen_at" timestamp DEFAULT NULL`,
		`ALTER TABLE "IdentityDeviceSessions" ADD COLUMN IF NOT EXISTS "sso_connection_id" varchar(36) NOT NULL DEFAULT ''`,
		`ALTER TABLE "IdentityDeviceSessions" ADD COLUMN IF NOT EXISTS "sso_name_id" text NOT NULL DEFAULT ''`,
		`ALTER TABLE "IdentityDeviceSessions" ADD COLUMN IF NOT EXISTS "sso_session_index" text NOT NULL DEFAULT ''`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_device_sessions_user" ON "IdentityDeviceSessions" ("user_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_device_sessions_sso" ON "IdentityDeviceSessions" ("sso_connection_id", "sso_name_id") WHERE "sso_connection_id" <> ''`,
	} {
		if _, err := q.Exec(s); err != nil {
			return err
		}
	}
	return nil
}
