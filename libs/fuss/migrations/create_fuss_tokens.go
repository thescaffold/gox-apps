package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateFussTokens struct{ sql.BaseMigration }

func (m *CreateFussTokens) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "FussTokens" (
			"id"           varchar(36)  NOT NULL,
			"created_at"   timestamp    DEFAULT NULL,
			"updated_at"   timestamp    DEFAULT NULL,
			"deleted_at"   timestamp    DEFAULT NULL,
			"user_id"      varchar(36)  NOT NULL,
			"client_id"    varchar(255) NOT NULL,
			"workspace_id" varchar(36)  NOT NULL,
			"group"        varchar(255) NOT NULL,
			"service"      varchar(255) NOT NULL,
			"entity_id"    varchar(36)  NOT NULL,
			"entity_name"  varchar(255) NOT NULL,
			"meta"         jsonb        DEFAULT NULL,
			"status"       varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}

	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_fuss_tokens_created_at"   ON "FussTokens" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_fuss_tokens_updated_at"   ON "FussTokens" ("updated_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_fuss_tokens_deleted_at"   ON "FussTokens" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_fuss_tokens_user_id"      ON "FussTokens" ("user_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_fuss_tokens_workspace_id" ON "FussTokens" ("workspace_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_fuss_tokens_group"        ON "FussTokens" ("group")`,
		`CREATE INDEX IF NOT EXISTS "idx_fuss_tokens_service"      ON "FussTokens" ("service")`,
		`CREATE INDEX IF NOT EXISTS "idx_fuss_tokens_entity_id"    ON "FussTokens" ("entity_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_fuss_tokens_entity_name"  ON "FussTokens" ("entity_name")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}

	return nil
}
