package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateIdentityTokens struct{ sql.BaseMigration }

func (m *CreateIdentityTokens) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "IdentityTokens" (
			"id"           varchar(36)  NOT NULL,
			"created_at"   timestamp    DEFAULT NULL,
			"updated_at"   timestamp    DEFAULT NULL,
			"deleted_at"   timestamp    DEFAULT NULL,
			"user_id"      varchar(36)  NOT NULL,
			"client_id"    varchar(255) NOT NULL,
			"workspace_id" varchar(36)  DEFAULT NULL,
			"type"         varchar(255) NOT NULL,
			"token"        text         NOT NULL,
			"device_id"    varchar(36)  DEFAULT NULL,
			"expires_at"   timestamp    DEFAULT NULL,
			"status"       varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_identity_tokens_created_at" ON "IdentityTokens" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_tokens_deleted_at" ON "IdentityTokens" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_tokens_user_id"    ON "IdentityTokens" ("user_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_tokens_client_id"  ON "IdentityTokens" ("client_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_tokens_type"       ON "IdentityTokens" ("type")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
