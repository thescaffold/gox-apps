package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateCapitalTransactions struct{ sql.BaseMigration }

func (m *CreateCapitalTransactions) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "CapitalTransactions" (
			"id"                varchar(36)  NOT NULL,
			"created_at"        timestamp    DEFAULT NULL,
			"updated_at"        timestamp    DEFAULT NULL,
			"deleted_at"        timestamp    DEFAULT NULL,
			"user_id"           varchar(36)  NOT NULL,
			"account_id"        varchar(36)  NOT NULL,
			"amount"            integer      NOT NULL,
			"book_balance"      integer      NOT NULL,
			"available_balance" integer      NOT NULL,
			"currency"          varchar(255) NOT NULL,
			"reference"         varchar(255) NOT NULL,
			"narration"         varchar(255) DEFAULT NULL,
			"desc"              text         DEFAULT NULL,
			"type"              varchar(255) NOT NULL,
			"status"            varchar(255) DEFAULT NULL,
			"transaction_at"    timestamp    NOT NULL,
			"verified_at"       timestamp    DEFAULT NULL,
			"invoiced_at"       timestamp    DEFAULT NULL,
			"paid_at"           timestamp    DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_capital_transactions_created_at"  ON "CapitalTransactions" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_capital_transactions_deleted_at"  ON "CapitalTransactions" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_capital_transactions_user_id"     ON "CapitalTransactions" ("user_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_capital_transactions_account_id"  ON "CapitalTransactions" ("account_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_capital_transactions_currency"    ON "CapitalTransactions" ("currency")`,
		`CREATE INDEX IF NOT EXISTS "idx_capital_transactions_reference"   ON "CapitalTransactions" ("reference")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
