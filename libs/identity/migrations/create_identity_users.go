package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateIdentityUsers struct{ sql.BaseMigration }

func (m *CreateIdentityUsers) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "IdentityUsers" (
			"id"         varchar(36)  NOT NULL,
			"created_at" timestamp    DEFAULT NULL,
			"updated_at" timestamp    DEFAULT NULL,
			"deleted_at" timestamp    DEFAULT NULL,
			"name"       varchar(255) NOT NULL,
			"email"      varchar(255) NOT NULL,
			"phone"      varchar(255) DEFAULT NULL,
			"secret"     text         DEFAULT NULL,
			"status"     varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_identity_users_created_at" ON "IdentityUsers" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_users_deleted_at" ON "IdentityUsers" ("deleted_at")`,
		`CREATE UNIQUE INDEX IF NOT EXISTS "idx_identity_users_email" ON "IdentityUsers" ("email")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
