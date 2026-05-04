package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateCapitalPayments struct{ sql.BaseMigration }

func (m *CreateCapitalPayments) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "CapitalPayments" (
			"id"           varchar(36)  NOT NULL,
			"created_at"   timestamp    DEFAULT NULL,
			"updated_at"   timestamp    DEFAULT NULL,
			"deleted_at"   timestamp    DEFAULT NULL,
			"user_id"      varchar(36)  NOT NULL,
			"client_id"    varchar(255) NOT NULL,
			"workspace_id" varchar(36)  NOT NULL,
			"plan_id"      varchar(36)  NOT NULL,
			"period_type"  varchar(255) DEFAULT NULL,
			"period"       varchar(255) DEFAULT NULL,
			"amount"       integer      DEFAULT NULL,
			"currency"     varchar(255) DEFAULT NULL,
			"paid"         boolean      DEFAULT NULL,
			"invoice_url"  varchar(255) DEFAULT NULL,
			"receipt_url"  varchar(255) DEFAULT NULL,
			"meta"         jsonb        DEFAULT NULL,
			"type"         varchar(255) NOT NULL,
			"reference"    varchar(255) NOT NULL,
			"request"      jsonb        DEFAULT NULL,
			"response"     jsonb        DEFAULT NULL,
			"start_at"     timestamp    DEFAULT NULL,
			"end_at"       timestamp    DEFAULT NULL,
			"status"       varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_capital_payments_created_at"   ON "CapitalPayments" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_capital_payments_deleted_at"   ON "CapitalPayments" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_capital_payments_user_id"      ON "CapitalPayments" ("user_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_capital_payments_workspace_id" ON "CapitalPayments" ("workspace_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_capital_payments_plan_id"      ON "CapitalPayments" ("plan_id")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
