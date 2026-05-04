package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateCommonProjects struct{ sql.BaseMigration }

func (m *CreateCommonProjects) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "CommonProjects" (
			"id"           varchar(36)  NOT NULL,
			"created_at"   timestamp    DEFAULT NULL,
			"updated_at"   timestamp    DEFAULT NULL,
			"deleted_at"   timestamp    DEFAULT NULL,
			"user_id"      varchar(36)  DEFAULT NULL,
			"client_id"    varchar(255) DEFAULT NULL,
			"workspace_id" varchar(36)  DEFAULT NULL,
			"group_name"   varchar(255) NOT NULL,
			"service_name" varchar(255) NOT NULL,
			"entity_name"  varchar(255) NOT NULL,
			"entity_id"    varchar(36)  NOT NULL,
			"type_id"      varchar(36)  NOT NULL,
			"status"       varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_common_projects_created_at"  ON "CommonProjects" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_common_projects_deleted_at"  ON "CommonProjects" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_common_projects_entity_name" ON "CommonProjects" ("entity_name")`,
		`CREATE INDEX IF NOT EXISTS "idx_common_projects_type_id"     ON "CommonProjects" ("type_id")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
