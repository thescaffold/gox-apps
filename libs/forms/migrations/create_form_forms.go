package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateFormForms struct{ sql.BaseMigration }

func (m *CreateFormForms) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "FormForms" (
			"id"           varchar(36)  NOT NULL,
			"created_at"   timestamp    DEFAULT NULL,
			"updated_at"   timestamp    DEFAULT NULL,
			"deleted_at"   timestamp    DEFAULT NULL,
			"user_id"      varchar(36)  NOT NULL,
			"client_id"    varchar(255) NOT NULL,
			"workspace_id" varchar(36)  NOT NULL,
			"type_id"      varchar(36)  NOT NULL,
			"key"          varchar(255) NOT NULL,
			"name"         varchar(255) NOT NULL,
			"desc"         varchar(255) DEFAULT NULL,
			"meta"         jsonb        DEFAULT NULL,
			"status"       varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_form_forms_created_at"   ON "FormForms" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_form_forms_deleted_at"   ON "FormForms" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_form_forms_workspace_id" ON "FormForms" ("workspace_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_form_forms_type_id"      ON "FormForms" ("type_id")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
