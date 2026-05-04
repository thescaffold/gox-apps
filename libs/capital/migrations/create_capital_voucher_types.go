package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateCapitalVoucherTypes struct{ sql.BaseMigration }

func (m *CreateCapitalVoucherTypes) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "CapitalVoucherTypes" (
			"id"         varchar(36)  NOT NULL,
			"created_at" timestamp    DEFAULT NULL,
			"updated_at" timestamp    DEFAULT NULL,
			"deleted_at" timestamp    DEFAULT NULL,
			"name"       varchar(255) NOT NULL,
			"desc"       varchar(255) DEFAULT NULL,
			"token"      varchar(255) NOT NULL,
			"amount"     integer      NOT NULL,
			"currency"   varchar(255) NOT NULL,
			"rules"      jsonb        DEFAULT NULL,
			"status"     varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_capital_voucher_types_created_at" ON "CapitalVoucherTypes" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_capital_voucher_types_deleted_at" ON "CapitalVoucherTypes" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_capital_voucher_types_name"       ON "CapitalVoucherTypes" ("name")`,
		`CREATE INDEX IF NOT EXISTS "idx_capital_voucher_types_token"      ON "CapitalVoucherTypes" ("token")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
