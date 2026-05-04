package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateCapitalPlans struct{ sql.BaseMigration }

func (m *CreateCapitalPlans) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "CapitalPlans" (
			"id"           varchar(36)  NOT NULL,
			"created_at"   timestamp    DEFAULT NULL,
			"updated_at"   timestamp    DEFAULT NULL,
			"deleted_at"   timestamp    DEFAULT NULL,
			"user_id"      varchar(36)  NOT NULL,
			"client_id"    varchar(255) NOT NULL,
			"workspace_id" varchar(36)  NOT NULL,
			"type_id"      varchar(36)  NOT NULL,
			"period_type"  varchar(255) DEFAULT NULL,
			"meta"         jsonb        DEFAULT NULL,
			"status"       varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_capital_plans_created_at"   ON "CapitalPlans" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_capital_plans_deleted_at"   ON "CapitalPlans" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_capital_plans_user_id"      ON "CapitalPlans" ("user_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_capital_plans_workspace_id" ON "CapitalPlans" ("workspace_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_capital_plans_type_id"      ON "CapitalPlans" ("type_id")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
