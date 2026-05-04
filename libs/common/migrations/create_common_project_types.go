package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateCommonProjectTypes struct{ sql.BaseMigration }

func (m *CreateCommonProjectTypes) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "CommonProjectTypes" (
			"id"            varchar(36)  NOT NULL,
			"created_at"    timestamp    DEFAULT NULL,
			"updated_at"    timestamp    DEFAULT NULL,
			"deleted_at"    timestamp    DEFAULT NULL,
			"user_id"       varchar(36)  DEFAULT NULL,
			"client_id"     varchar(255) DEFAULT NULL,
			"workspace_id"  varchar(36)  DEFAULT NULL,
			"name"          varchar(255) NOT NULL,
			"desc"          varchar(255) DEFAULT NULL,
			"thumbnail_url" varchar(255) DEFAULT NULL,
			"status"        varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_common_project_types_created_at" ON "CommonProjectTypes" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_common_project_types_deleted_at" ON "CommonProjectTypes" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_common_project_types_name"       ON "CommonProjectTypes" ("name")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
