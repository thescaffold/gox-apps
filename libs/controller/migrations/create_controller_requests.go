package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateControllerRequests struct{ sql.BaseMigration }

func (m *CreateControllerRequests) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "ControllerRequests" (
			"id"         varchar(36)  NOT NULL,
			"created_at" timestamp    DEFAULT NULL,
			"updated_at" timestamp    DEFAULT NULL,
			"deleted_at" timestamp    DEFAULT NULL,
			"group"      varchar(255) DEFAULT NULL,
			"service"    varchar(255) DEFAULT NULL,
			"type"       varchar(255) NOT NULL DEFAULT 'http',
			"method"     varchar(255) NOT NULL,
			"url"        varchar(255) NOT NULL,
			"queries"    jsonb        DEFAULT NULL,
			"body"       jsonb        DEFAULT NULL,
			"response"   jsonb        DEFAULT NULL,
			"status"     varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}

	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_controller_requests_created_at" ON "ControllerRequests" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_controller_requests_updated_at" ON "ControllerRequests" ("updated_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_controller_requests_deleted_at" ON "ControllerRequests" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_controller_requests_group"      ON "ControllerRequests" ("group")`,
		`CREATE INDEX IF NOT EXISTS "idx_controller_requests_service"    ON "ControllerRequests" ("service")`,
		`CREATE INDEX IF NOT EXISTS "idx_controller_requests_method"     ON "ControllerRequests" ("method")`,
		`CREATE INDEX IF NOT EXISTS "idx_controller_requests_url"        ON "ControllerRequests" ("url")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}

	return nil
}
