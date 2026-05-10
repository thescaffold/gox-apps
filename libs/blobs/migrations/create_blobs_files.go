package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateBlobsFiles struct{ sql.BaseMigration }

func (m *CreateBlobsFiles) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "BlobsFiles" (
			"id"           varchar(36)  NOT NULL,
			"created_at"   timestamp    DEFAULT NULL,
			"updated_at"   timestamp    DEFAULT NULL,
			"deleted_at"   timestamp    DEFAULT NULL,
			"user_id"      varchar(36)  NOT NULL,
			"client_id"    varchar(255) NOT NULL,
			"workspace_id" varchar(36)  NOT NULL,
			"type"         varchar(255) NOT NULL,
			"parent_id"    varchar(36)  DEFAULT NULL,
			"name"         varchar(255) NOT NULL,
			"tags"         jsonb        DEFAULT NULL,
			"size"         bigint       NOT NULL,
			"mime"         varchar(255) NOT NULL,
			"status"       varchar(255) DEFAULT NULL,
			"meta"         jsonb        DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}

	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_blobs_files_created_at"   ON "BlobsFiles" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_blobs_files_updated_at"   ON "BlobsFiles" ("updated_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_blobs_files_deleted_at"   ON "BlobsFiles" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_blobs_files_user_id"      ON "BlobsFiles" ("user_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_blobs_files_client_id"    ON "BlobsFiles" ("client_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_blobs_files_workspace_id" ON "BlobsFiles" ("workspace_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_blobs_files_type"         ON "BlobsFiles" ("type")`,
		`CREATE INDEX IF NOT EXISTS "idx_blobs_files_parent_id"    ON "BlobsFiles" ("parent_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_blobs_files_mime"         ON "BlobsFiles" ("mime")`,
		`CREATE INDEX IF NOT EXISTS "idx_blobs_files_status"       ON "BlobsFiles" ("status")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}

	return nil
}
