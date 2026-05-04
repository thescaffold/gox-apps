package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateCapitalVouchers struct{ sql.BaseMigration }

func (m *CreateCapitalVouchers) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "CapitalVouchers" (
			"id"               varchar(36)  NOT NULL,
			"created_at"       timestamp    DEFAULT NULL,
			"updated_at"       timestamp    DEFAULT NULL,
			"deleted_at"       timestamp    DEFAULT NULL,
			"user_id"          varchar(36)  NOT NULL,
			"client_id"        varchar(255) NOT NULL,
			"workspace_id"     varchar(36)  NOT NULL,
			"voucher_type_id"  varchar(36)  NOT NULL,
			"status"           varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_capital_vouchers_created_at"      ON "CapitalVouchers" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_capital_vouchers_deleted_at"      ON "CapitalVouchers" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_capital_vouchers_user_id"         ON "CapitalVouchers" ("user_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_capital_vouchers_workspace_id"    ON "CapitalVouchers" ("workspace_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_capital_vouchers_voucher_type_id" ON "CapitalVouchers" ("voucher_type_id")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
