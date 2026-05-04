package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateBlobsFiles struct{ sql.BaseMigration }

func (m *CreateBlobsFiles) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "BlobsFiles" (
			"id"         varchar(36)  NOT NULL,
			"created_at" timestamp    DEFAULT NULL,
			"updated_at" timestamp    DEFAULT NULL,
			"deleted_at" timestamp    DEFAULT NULL,
			"type"       varchar(255) NOT NULL,
			"parent_id"  varchar(36)  DEFAULT NULL,
			"name"       varchar(255) NOT NULL,
			"bucket"     varchar(255) NOT NULL,
			"url"        varchar(500) DEFAULT NULL,
			"size"       bigint       DEFAULT NULL,
			"mime"       varchar(255) DEFAULT NULL,
			"tags"       jsonb        DEFAULT NULL,
			"meta"       jsonb        DEFAULT NULL,
			"status"     varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}

	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_blobs_files_created_at" ON "BlobsFiles" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_blobs_files_updated_at" ON "BlobsFiles" ("updated_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_blobs_files_deleted_at" ON "BlobsFiles" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_blobs_files_parent_id"  ON "BlobsFiles" ("parent_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_blobs_files_bucket"     ON "BlobsFiles" ("bucket")`,
		`CREATE INDEX IF NOT EXISTS "idx_blobs_files_status"     ON "BlobsFiles" ("status")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}

	return nil
}
