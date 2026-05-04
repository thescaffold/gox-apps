package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateStaticsLists struct{ sql.BaseMigration }

func (m *CreateStaticsLists) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "StaticsLists" (
			"id"        varchar(36)  NOT NULL,
			"created_at" timestamp   DEFAULT NULL,
			"updated_at" timestamp   DEFAULT NULL,
			"deleted_at" timestamp   DEFAULT NULL,
			"parent_id" varchar(36)  DEFAULT NULL,
			"key"       varchar(255) NOT NULL,
			"code"      varchar(255) DEFAULT NULL,
			"value"     varchar(255) NOT NULL,
			"meta"      jsonb        DEFAULT NULL,
			"status"    varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}

	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_statics_lists_created_at" ON "StaticsLists" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_statics_lists_updated_at" ON "StaticsLists" ("updated_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_statics_lists_deleted_at" ON "StaticsLists" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_statics_lists_parent_id"  ON "StaticsLists" ("parent_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_statics_lists_key"        ON "StaticsLists" ("key")`,
		`CREATE INDEX IF NOT EXISTS "idx_statics_lists_status"     ON "StaticsLists" ("status")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}

	return nil
}
