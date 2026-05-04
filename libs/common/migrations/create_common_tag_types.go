package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateCommonTagTypes struct{ sql.BaseMigration }

func (m *CreateCommonTagTypes) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "CommonTagTypes" (
			"id"           varchar(36)  NOT NULL,
			"created_at"   timestamp    DEFAULT NULL,
			"updated_at"   timestamp    DEFAULT NULL,
			"deleted_at"   timestamp    DEFAULT NULL,
			"user_id"      varchar(36)  DEFAULT NULL,
			"client_id"    varchar(255) DEFAULT NULL,
			"workspace_id" varchar(36)  DEFAULT NULL,
			"name"         varchar(255) NOT NULL,
			"desc"         varchar(255) DEFAULT NULL,
			"status"       varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_common_tag_types_created_at" ON "CommonTagTypes" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_common_tag_types_deleted_at" ON "CommonTagTypes" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_common_tag_types_name"       ON "CommonTagTypes" ("name")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
