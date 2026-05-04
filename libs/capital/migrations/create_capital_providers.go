package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateCapitalProviders struct{ sql.BaseMigration }

func (m *CreateCapitalProviders) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "CapitalProviders" (
			"id"         varchar(36)  NOT NULL,
			"created_at" timestamp    DEFAULT NULL,
			"updated_at" timestamp    DEFAULT NULL,
			"deleted_at" timestamp    DEFAULT NULL,
			"account_id" varchar(36)  DEFAULT NULL,
			"name"       varchar(255) NOT NULL,
			"type"       varchar(255) NOT NULL,
			"signature"  varchar(255) DEFAULT NULL,
			"email"      varchar(255) DEFAULT NULL,
			"token"      varchar(255) DEFAULT NULL,
			"primary"    boolean      DEFAULT NULL,
			"currency"   varchar(255) NOT NULL,
			"country"    varchar(255) NOT NULL,
			"meta"       jsonb        DEFAULT NULL,
			"status"     varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_capital_providers_created_at" ON "CapitalProviders" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_capital_providers_deleted_at" ON "CapitalProviders" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_capital_providers_account_id" ON "CapitalProviders" ("account_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_capital_providers_currency"   ON "CapitalProviders" ("currency")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
