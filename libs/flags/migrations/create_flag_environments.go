package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateFlagEnvironments struct{ sql.BaseMigration }

func (m *CreateFlagEnvironments) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "FlagEnvironments" (
			"id"           varchar(36)  NOT NULL,
			"created_at"   timestamp    DEFAULT NULL,
			"updated_at"   timestamp    DEFAULT NULL,
			"deleted_at"   timestamp    DEFAULT NULL,
			"user_id"      varchar(36)  NOT NULL,
			"client_id"    varchar(255) NOT NULL,
			"workspace_id" varchar(36)  NOT NULL,
			"type_id"      varchar(36)  NOT NULL,
			"name"         varchar(255) NOT NULL,
			"desc"         varchar(255) DEFAULT NULL,
			"meta"         jsonb        DEFAULT NULL,
			"status"       varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_flag_environments_created_at"   ON "FlagEnvironments" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_flag_environments_deleted_at"   ON "FlagEnvironments" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_flag_environments_workspace_id" ON "FlagEnvironments" ("workspace_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_flag_environments_type_id"      ON "FlagEnvironments" ("type_id")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
