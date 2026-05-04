package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateCapitalPlanTypes struct{ sql.BaseMigration }

func (m *CreateCapitalPlanTypes) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "CapitalPlanTypes" (
			"id"         varchar(36)  NOT NULL,
			"created_at" timestamp    DEFAULT NULL,
			"updated_at" timestamp    DEFAULT NULL,
			"deleted_at" timestamp    DEFAULT NULL,
			"client_id"  varchar(255) NOT NULL,
			"key"        varchar(255) NOT NULL,
			"name"       varchar(255) NOT NULL,
			"desc"       varchar(255) DEFAULT NULL,
			"detail"     text         DEFAULT NULL,
			"type"       varchar(255) DEFAULT NULL,
			"currency"   varchar(255) NOT NULL,
			"daily"      integer      DEFAULT NULL,
			"weekly"     integer      DEFAULT NULL,
			"monthly"    integer      NOT NULL DEFAULT 0,
			"yearly"     integer      DEFAULT NULL,
			"meta"       jsonb        DEFAULT NULL,
			"status"     varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_capital_plan_types_created_at" ON "CapitalPlanTypes" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_capital_plan_types_deleted_at" ON "CapitalPlanTypes" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_capital_plan_types_client_id"  ON "CapitalPlanTypes" ("client_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_capital_plan_types_key"        ON "CapitalPlanTypes" ("key")`,
		`CREATE INDEX IF NOT EXISTS "idx_capital_plan_types_name"       ON "CapitalPlanTypes" ("name")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
