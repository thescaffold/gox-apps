package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateCapitalRates struct{ sql.BaseMigration }

func (m *CreateCapitalRates) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "CapitalRates" (
			"id"         varchar(36)  NOT NULL,
			"created_at" timestamp    DEFAULT NULL,
			"updated_at" timestamp    DEFAULT NULL,
			"deleted_at" timestamp    DEFAULT NULL,
			"type"       varchar(255) NOT NULL,
			"per_unit"   integer      DEFAULT NULL,
			"currency"   varchar(255) NOT NULL,
			"status"     varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_capital_rates_created_at" ON "CapitalRates" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_capital_rates_deleted_at" ON "CapitalRates" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_capital_rates_type"       ON "CapitalRates" ("type")`,
		`CREATE INDEX IF NOT EXISTS "idx_capital_rates_currency"   ON "CapitalRates" ("currency")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
