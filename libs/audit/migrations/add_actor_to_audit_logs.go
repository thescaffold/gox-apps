package migrations

import "github.com/awesome-goose/goose/modules/sql"

// AddActorToAuditLogs adds the general (actor_type, actor_id) reference
// (TRD §5, PLAN M1-04/TRD U-S4) alongside the existing user_id/client_id
// columns — additive, both nullable, so every pre-existing row (which never
// had a non-human actor) stays valid without a backfill. user_id/client_id
// stay exactly as they are: they're populated from the real authenticated
// request session, which an AI or system-initiated event usually won't
// have.
type AddActorToAuditLogs struct{ sql.BaseMigration }

func (m *AddActorToAuditLogs) Run(q *sql.Query) error {
	if _, err := q.Exec(`
		ALTER TABLE "AuditLogs"
			ADD COLUMN IF NOT EXISTS "actor_type" varchar(255) DEFAULT NULL,
			ADD COLUMN IF NOT EXISTS "actor_id"   varchar(255) DEFAULT NULL
	`); err != nil {
		return err
	}
	if _, err := q.Exec(`
		CREATE INDEX IF NOT EXISTS "idx_audit_logs_actor_type" ON "AuditLogs" ("actor_type")
	`); err != nil {
		return err
	}
	if _, err := q.Exec(`
		CREATE INDEX IF NOT EXISTS "idx_audit_logs_actor_id" ON "AuditLogs" ("actor_id")
	`); err != nil {
		return err
	}
	return nil
}
