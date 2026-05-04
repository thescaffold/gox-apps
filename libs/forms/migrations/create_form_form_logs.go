package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateFormFormLogs struct{ sql.BaseMigration }

func (m *CreateFormFormLogs) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "FormFormLogs" (
			"id"           varchar(36)  NOT NULL,
			"created_at"   timestamp    DEFAULT NULL,
			"updated_at"   timestamp    DEFAULT NULL,
			"deleted_at"   timestamp    DEFAULT NULL,
			"form_id"      varchar(36)  NOT NULL,
			"form_field_id" varchar(36) NOT NULL,
			"key"          varchar(255) NOT NULL,
			"value"        varchar(255) NOT NULL,
			"meta"         jsonb        DEFAULT NULL,
			"status"       varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_form_form_logs_created_at" ON "FormFormLogs" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_form_form_logs_deleted_at" ON "FormFormLogs" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_form_form_logs_form_id"    ON "FormFormLogs" ("form_id")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
