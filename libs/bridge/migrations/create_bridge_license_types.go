package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateBridgeLicenseTypes struct{ sql.BaseMigration }

func (m *CreateBridgeLicenseTypes) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "BridgeLicenseTypes" (
			"id"         varchar(36)    NOT NULL,
			"created_at" timestamp      DEFAULT NULL,
			"updated_at" timestamp      DEFAULT NULL,
			"deleted_at" timestamp      DEFAULT NULL,
			"key"        varchar(255)   NOT NULL,
			"name"       varchar(255)   NOT NULL,
			"desc"       varchar(255)   DEFAULT NULL,
			"detail"     text           DEFAULT NULL,
			"type"       varchar(255)   DEFAULT NULL,
			"currency"   varchar(255)   NOT NULL,
			"daily"      decimal(18,2)  DEFAULT NULL,
			"weekly"     decimal(18,2)  DEFAULT NULL,
			"monthly"    decimal(18,2)  NOT NULL DEFAULT 0,
			"yearly"     decimal(18,2)  DEFAULT NULL,
			"meta"       jsonb          DEFAULT NULL,
			"status"     varchar(255)   DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_bridge_license_types_created_at" ON "BridgeLicenseTypes" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_bridge_license_types_deleted_at" ON "BridgeLicenseTypes" ("deleted_at")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
