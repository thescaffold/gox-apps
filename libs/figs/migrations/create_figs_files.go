package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateFigsFiles struct{ sql.BaseMigration }

func (m *CreateFigsFiles) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "FigsFiles" (
			"id"         varchar(36)  NOT NULL,
			"created_at" timestamp    DEFAULT NULL,
			"updated_at" timestamp    DEFAULT NULL,
			"deleted_at" timestamp    DEFAULT NULL,
			"name"       varchar(255) NOT NULL,
			"input"      jsonb        NOT NULL,
			"output"     jsonb        NOT NULL,
			"meta"       jsonb        NOT NULL,
			"url"        varchar(255) DEFAULT NULL,
			"raw"        text,
			"tags"       jsonb        DEFAULT NULL,
			"status"     varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}

	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_figs_files_created_at" ON "FigsFiles" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_figs_files_updated_at" ON "FigsFiles" ("updated_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_figs_files_deleted_at" ON "FigsFiles" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_figs_files_status"     ON "FigsFiles" ("status")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}

	return nil
}
