package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateBridgeLicenses struct{ sql.BaseMigration }

func (m *CreateBridgeLicenses) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "BridgeLicenses" (
			"id"           varchar(36)  NOT NULL,
			"created_at"   timestamp    DEFAULT NULL,
			"updated_at"   timestamp    DEFAULT NULL,
			"deleted_at"   timestamp    DEFAULT NULL,
			"user_id"      varchar(36)  NOT NULL,
			"client_id"    varchar(255) NOT NULL,
			"workspace_id" varchar(36)  NOT NULL,
			"type_id"      varchar(36)  NOT NULL,
			"period_type"  varchar(255) DEFAULT NULL,
			"token"        text         DEFAULT NULL,
			"start_at"     timestamp    DEFAULT NULL,
			"renewed_at"   timestamp    DEFAULT NULL,
			"expired_at"   timestamp    DEFAULT NULL,
			"meta"         jsonb        DEFAULT NULL,
			"status"       varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_bridge_licenses_created_at"   ON "BridgeLicenses" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_bridge_licenses_deleted_at"   ON "BridgeLicenses" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_bridge_licenses_workspace_id" ON "BridgeLicenses" ("workspace_id")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
