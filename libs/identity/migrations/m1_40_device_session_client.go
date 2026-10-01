package migrations

import "github.com/awesome-goose/goose/modules/sql"

// DeviceSessionClient adds the client a device session belongs to (PLAN M1-40).
// The browser shell lists a device's sessions to show its signed-in accounts and
// to re-mint a token, and a token is minted for one client, so the session has
// to remember which. Nullable: sessions made before this keep working and fall
// back to the client the caller names.
type DeviceSessionClient struct{ sql.BaseMigration }

func (m *DeviceSessionClient) Run(q *sql.Query) error {
	_, err := q.Exec(`ALTER TABLE "IdentityDeviceSessions" ADD COLUMN IF NOT EXISTS "client_id" varchar(255) DEFAULT NULL`)
	return err
}
