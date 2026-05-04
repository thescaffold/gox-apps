package migrations

import "github.com/awesome-goose/goose/modules/sql"

type CreateAuditLogs struct{ sql.BaseMigration }

func (m *CreateAuditLogs) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		CREATE TABLE IF NOT EXISTS "AuditLogs" (
			"id"           varchar(36)  NOT NULL,
			"created_at"   timestamp    DEFAULT NULL,
			"updated_at"   timestamp    DEFAULT NULL,
			"deleted_at"   timestamp    DEFAULT NULL,
			"user_id"      varchar(36)  NOT NULL,
			"client_id"    varchar(255) NOT NULL,
			"workspace_id" varchar(36)  NOT NULL,
			"group"        varchar(255) NOT NULL,
			"service"      varchar(255) NOT NULL,
			"entity_id"    varchar(36)  NOT NULL,
			"entity_name"  varchar(255) NOT NULL,
			"action"       varchar(255) NOT NULL,
			"desc"         varchar(255) DEFAULT NULL,
			"meta"         jsonb        DEFAULT NULL,
			"status"       varchar(255) DEFAULT NULL,
			PRIMARY KEY ("id")
		)
	`); err != nil {
		return err
	}

	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS "idx_audit_logs_created_at"   ON "AuditLogs" ("created_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_audit_logs_updated_at"   ON "AuditLogs" ("updated_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_audit_logs_deleted_at"   ON "AuditLogs" ("deleted_at")`,
		`CREATE INDEX IF NOT EXISTS "idx_audit_logs_user_id"      ON "AuditLogs" ("user_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_audit_logs_client_id"    ON "AuditLogs" ("client_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_audit_logs_workspace_id" ON "AuditLogs" ("workspace_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_audit_logs_group"        ON "AuditLogs" ("group")`,
		`CREATE INDEX IF NOT EXISTS "idx_audit_logs_service"      ON "AuditLogs" ("service")`,
		`CREATE INDEX IF NOT EXISTS "idx_audit_logs_entity_id"    ON "AuditLogs" ("entity_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_audit_logs_entity_name"  ON "AuditLogs" ("entity_name")`,
		`CREATE INDEX IF NOT EXISTS "idx_audit_logs_action"       ON "AuditLogs" ("action")`,
		`CREATE INDEX IF NOT EXISTS "idx_audit_logs_status"       ON "AuditLogs" ("status")`,
	} {
		if _, err := q.Exec(idx); err != nil {
			return err
		}
	}

	return nil
}
