package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateCapitalAccounts struct{ sql.BaseMigration }

func (m *CreateCapitalAccounts) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "CapitalAccounts" (
			"id"                varchar(36)  NOT NULL,
			"created_at"        timestamp    DEFAULT NULL,
			"updated_at"        timestamp    DEFAULT NULL,
			"deleted_at"        timestamp    DEFAULT NULL,
			"user_id"           varchar(36)  NOT NULL,
			"client_id"         varchar(255) NOT NULL,
			"workspace_id"      varchar(36)  NOT NULL,
			"book_balance"      integer      NOT NULL DEFAULT 0,
			"available_balance" integer      NOT NULL DEFAULT 0,
			"currency"          varchar(255) NOT NULL,
			"reference"         varchar(255) NOT NULL,
			"desc"              varchar(255) DEFAULT NULL,
			"label"             varchar(255) DEFAULT NULL,
			"type"              varchar(255) DEFAULT NULL,
			"number"            varchar(255) DEFAULT NULL,
			"daily_limit"       integer      DEFAULT NULL,
			"monthly_limit"     integer      DEFAULT NULL,
			"status"            varchar(255) DEFAULT NULL,
			"suspended_at"      timestamp    DEFAULT NULL,
			"closed_at"         timestamp    DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_capital_accounts_created_at"   ON "CapitalAccounts" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_capital_accounts_deleted_at"   ON "CapitalAccounts" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_capital_accounts_user_id"      ON "CapitalAccounts" ("user_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_capital_accounts_workspace_id" ON "CapitalAccounts" ("workspace_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_capital_accounts_currency"     ON "CapitalAccounts" ("currency")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
