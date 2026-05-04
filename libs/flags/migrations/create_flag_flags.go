package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateFlagFlags struct{ sql.BaseMigration }

func (m *CreateFlagFlags) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "FlagFlags" (
			"id"             varchar(36)  NOT NULL,
			"created_at"     timestamp    DEFAULT NULL,
			"updated_at"     timestamp    DEFAULT NULL,
			"deleted_at"     timestamp    DEFAULT NULL,
			"user_id"        varchar(36)  NOT NULL,
			"client_id"      varchar(255) NOT NULL,
			"workspace_id"   varchar(36)  NOT NULL,
			"environment_id" varchar(36)  NOT NULL,
			"name"           varchar(255) NOT NULL,
			"limit"          int          NOT NULL DEFAULT 0,
			"priority"       int          NOT NULL DEFAULT 0,
			"level"          varchar(255) NOT NULL,
			"meta"           jsonb        NOT NULL,
			"status"         varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_flag_flags_created_at"     ON "FlagFlags" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_flag_flags_deleted_at"     ON "FlagFlags" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_flag_flags_workspace_id"   ON "FlagFlags" ("workspace_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_flag_flags_environment_id" ON "FlagFlags" ("environment_id")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
