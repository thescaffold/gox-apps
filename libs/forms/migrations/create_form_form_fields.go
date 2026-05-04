package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateFormFormFields struct{ sql.BaseMigration }

func (m *CreateFormFormFields) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "FormFormFields" (
			"id"            varchar(36)  NOT NULL,
			"created_at"    timestamp    DEFAULT NULL,
			"updated_at"    timestamp    DEFAULT NULL,
			"deleted_at"    timestamp    DEFAULT NULL,
			"form_id"       varchar(36)  NOT NULL,
			"key"           varchar(255) NOT NULL,
			"name"          varchar(255) NOT NULL,
			"type"          varchar(255) NOT NULL,
			"desc"          varchar(255) DEFAULT NULL,
			"placeholder"   varchar(255) DEFAULT NULL,
			"disabled"      boolean      DEFAULT NULL,
			"required"      boolean      DEFAULT NULL,
			"default_value" varchar(255) DEFAULT NULL,
			"meta"          jsonb        DEFAULT NULL,
			"status"        varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_form_form_fields_created_at" ON "FormFormFields" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_form_form_fields_deleted_at" ON "FormFormFields" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_form_form_fields_form_id"    ON "FormFormFields" ("form_id")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
