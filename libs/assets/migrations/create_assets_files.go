package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateAssetsFiles struct{ sql.BaseMigration }

func (m *CreateAssetsFiles) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "AssetsFiles" (
			"id"         varchar(36)  NOT NULL,
			"created_at" timestamp    DEFAULT NULL,
			"updated_at" timestamp    DEFAULT NULL,
			"deleted_at" timestamp    DEFAULT NULL,
			"type"       varchar(255) NOT NULL,
			"bucket"     varchar(255) NOT NULL,
			"name"       varchar(255) NOT NULL,
			"tags"       jsonb        DEFAULT NULL,
			"raw"        text,
			"status"     varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}

	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_assets_files_created_at" ON "AssetsFiles" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_assets_files_updated_at" ON "AssetsFiles" ("updated_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_assets_files_deleted_at" ON "AssetsFiles" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_assets_files_type"       ON "AssetsFiles" ("type")`,
		`CREATE INDEX IF NOT EXISTS "idx_assets_files_bucket"     ON "AssetsFiles" ("bucket")`,
		`CREATE INDEX IF NOT EXISTS "idx_assets_files_name"       ON "AssetsFiles" ("name")`,
		`CREATE INDEX IF NOT EXISTS "idx_assets_files_status"     ON "AssetsFiles" ("status")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}

	return nil
}
