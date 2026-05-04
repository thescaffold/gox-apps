package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateControllerRoutes struct{ sql.BaseMigration }

func (m *CreateControllerRoutes) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "ControllerRoutes" (
			"id"         varchar(36)  NOT NULL,
			"created_at" timestamp    DEFAULT NULL,
			"updated_at" timestamp    DEFAULT NULL,
			"deleted_at" timestamp    DEFAULT NULL,
			"group"      varchar(255) NOT NULL,
			"service"    varchar(255) NOT NULL,
			"type"       varchar(255) NOT NULL DEFAULT 'http',
			"name"       varchar(255) NOT NULL,
			"desc"       varchar(255) DEFAULT NULL,
			"upstream"   varchar(255) NOT NULL,
			"status"     varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}

	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_controller_routes_created_at" ON "ControllerRoutes" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_controller_routes_updated_at" ON "ControllerRoutes" ("updated_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_controller_routes_deleted_at" ON "ControllerRoutes" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_controller_routes_group"      ON "ControllerRoutes" ("group")`,
		`CREATE INDEX IF NOT EXISTS "idx_controller_routes_service"    ON "ControllerRoutes" ("service")`,
		`CREATE INDEX IF NOT EXISTS "idx_controller_routes_type"       ON "ControllerRoutes" ("type")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}

	return nil
}
