package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateCapitalPaymentLogs struct{ sql.BaseMigration }

func (m *CreateCapitalPaymentLogs) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "CapitalPaymentLogs" (
			"id"         varchar(36)  NOT NULL,
			"created_at" timestamp    DEFAULT NULL,
			"updated_at" timestamp    DEFAULT NULL,
			"deleted_at" timestamp    DEFAULT NULL,
			"payment_id" varchar(36)  NOT NULL,
			"type"       varchar(255) NOT NULL,
			"request"    jsonb        NOT NULL,
			"response"   jsonb        DEFAULT NULL,
			"status"     varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_capital_payment_logs_created_at" ON "CapitalPaymentLogs" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_capital_payment_logs_deleted_at" ON "CapitalPaymentLogs" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_capital_payment_logs_payment_id" ON "CapitalPaymentLogs" ("payment_id")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
