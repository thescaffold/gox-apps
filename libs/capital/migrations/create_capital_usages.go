package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateCapitalUsages struct{ sql.BaseMigration }

func (m *CreateCapitalUsages) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "CapitalUsages" (
			"id"           varchar(36)  NOT NULL,
			"created_at"   timestamp    DEFAULT NULL,
			"updated_at"   timestamp    DEFAULT NULL,
			"deleted_at"   timestamp    DEFAULT NULL,
			"user_id"      varchar(36)  NOT NULL,
			"client_id"    varchar(255) NOT NULL,
			"workspace_id" varchar(36)  NOT NULL,
			"rate_id"      varchar(36)  NOT NULL,
			"entity_id"    varchar(36)  NOT NULL,
			"entity_name"  varchar(255) NOT NULL,
			"quantity"     integer      NOT NULL,
			"start_at"     timestamp    DEFAULT NULL,
			"stop_at"      timestamp    DEFAULT NULL,
			"meta"         jsonb        DEFAULT NULL,
			"status"       varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_capital_usages_created_at"   ON "CapitalUsages" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_capital_usages_deleted_at"   ON "CapitalUsages" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_capital_usages_user_id"      ON "CapitalUsages" ("user_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_capital_usages_workspace_id" ON "CapitalUsages" ("workspace_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_capital_usages_rate_id"      ON "CapitalUsages" ("rate_id")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
