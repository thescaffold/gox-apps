package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateCommonRates struct{ sql.BaseMigration }

func (m *CreateCommonRates) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "CommonRates" (
			"id"         varchar(36)  NOT NULL,
			"created_at" timestamp    DEFAULT NULL,
			"updated_at" timestamp    DEFAULT NULL,
			"deleted_at" timestamp    DEFAULT NULL,
			"currency"   varchar(255) NOT NULL,
			"value"      integer      DEFAULT NULL,
			"delta"      integer      DEFAULT NULL,
			"status"     varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id"),
			UNIQUE ("currency")
		)
	`); err != nil {
		return err
	}
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_common_rates_created_at" ON "CommonRates" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_common_rates_deleted_at" ON "CommonRates" ("deleted_at")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
