package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateCacheLists struct{ sql.BaseMigration }

func (m *CreateCacheLists) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "CacheLists" (
			"id"         varchar(36)  NOT NULL,
			"created_at" timestamp    DEFAULT NULL,
			"updated_at" timestamp    DEFAULT NULL,
			"deleted_at" timestamp    DEFAULT NULL,
			"key"        varchar(255) NOT NULL,
			"value"      jsonb        DEFAULT NULL,
			"group"      varchar(255) NOT NULL DEFAULT '',
			"expired_at" timestamp    DEFAULT NULL,
			"meta"       jsonb        DEFAULT NULL,
			"status"     varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id"),
			CONSTRAINT "unq_cache_lists_key_group" UNIQUE ("key", "group")
		)
	`); err != nil {
		return err
	}

	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_cache_lists_created_at" ON "CacheLists" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_cache_lists_updated_at" ON "CacheLists" ("updated_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_cache_lists_deleted_at" ON "CacheLists" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_cache_lists_key"        ON "CacheLists" ("key")`,
		`CREATE INDEX IF NOT EXISTS "idx_cache_lists_group"      ON "CacheLists" ("group")`,
		`CREATE INDEX IF NOT EXISTS "idx_cache_lists_expired_at" ON "CacheLists" ("expired_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_cache_lists_status"     ON "CacheLists" ("status")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}

	return nil
}
