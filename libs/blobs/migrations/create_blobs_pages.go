package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateBlobsPages struct{ sql.BaseMigration }

func (m *CreateBlobsPages) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "BlobsPages" (
			"id"         varchar(36) NOT NULL,
			"created_at" timestamp   DEFAULT NULL,
			"updated_at" timestamp   DEFAULT NULL,
			"deleted_at" timestamp   DEFAULT NULL,
			"file_id"    varchar(36) NOT NULL,
			"index"      int         NOT NULL,
			"raw"        bytea,
			"meta"       jsonb       DEFAULT NULL,
			"status"     varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}

	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_blobs_pages_created_at" ON "BlobsPages" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_blobs_pages_updated_at" ON "BlobsPages" ("updated_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_blobs_pages_deleted_at" ON "BlobsPages" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_blobs_pages_file_id"          ON "BlobsPages" ("file_id")`,
		// Composite index on (file_id, index) for efficient ordered chunk lookup, matching TS BlobsPage migration.
		`CREATE INDEX IF NOT EXISTS "idx_blobs_pages_file_id_index"    ON "BlobsPages" ("file_id", "index")`,
		`CREATE INDEX IF NOT EXISTS "idx_blobs_pages_status"           ON "BlobsPages" ("status")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}

	return nil
}
