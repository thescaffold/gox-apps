package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateCommonIps struct{ sql.BaseMigration }

func (m *CreateCommonIps) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "CommonIps" (
			"id"         varchar(36)  NOT NULL,
			"created_at" timestamp    DEFAULT NULL,
			"updated_at" timestamp    DEFAULT NULL,
			"deleted_at" timestamp    DEFAULT NULL,
			"value"      varchar(255) NOT NULL,
			"country"    varchar(255) NOT NULL,
			"city"       varchar(255) DEFAULT NULL,
			"status"     varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id"),
			UNIQUE ("value")
		)
	`); err != nil {
		return err
	}
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_common_ips_created_at" ON "CommonIps" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_common_ips_deleted_at" ON "CommonIps" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_common_ips_country"    ON "CommonIps" ("country")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
