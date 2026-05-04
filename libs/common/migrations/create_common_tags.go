package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateCommonTags struct{ sql.BaseMigration }

func (m *CreateCommonTags) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "CommonTags" (
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
		`CREATE INDEX IF NOT EXISTS "idx_common_tags_created_at"   ON "CommonTags" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_common_tags_deleted_at"   ON "CommonTags" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_common_tags_entity_name"  ON "CommonTags" ("entity_name")`,
		`CREATE INDEX IF NOT EXISTS "idx_common_tags_type_id"      ON "CommonTags" ("type_id")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
