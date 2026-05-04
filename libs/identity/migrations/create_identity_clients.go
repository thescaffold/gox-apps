package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateIdentityClients struct{ sql.BaseMigration }

func (m *CreateIdentityClients) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "IdentityClients" (
			"id"         varchar(36)  NOT NULL,
			"created_at" timestamp    DEFAULT NULL,
			"updated_at" timestamp    DEFAULT NULL,
			"deleted_at" timestamp    DEFAULT NULL,
			"name"       varchar(255) NOT NULL,
			"type"       varchar(255) NOT NULL,
			"key"        varchar(255) NOT NULL,
			"secret"     text         DEFAULT NULL,
			"desc"       varchar(255) DEFAULT NULL,
			"status"     varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_identity_clients_created_at" ON "IdentityClients" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_clients_deleted_at" ON "IdentityClients" ("deleted_at")`,
		`CREATE UNIQUE INDEX IF NOT EXISTS "idx_identity_clients_key" ON "IdentityClients" ("key")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
