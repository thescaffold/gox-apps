package migrations

import "github.com/awesome-goose/goose/modules/sql"

// CreateIdentityOrgEvents adds the audit trail of a workspace's organisation:
// who created, changed or removed a team or a role, added or removed a person,
// or gave or took a role, and when.
type CreateIdentityOrgEvents struct{ sql.BaseMigration }

func (m *CreateIdentityOrgEvents) Run(q *sql.Query) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS "IdentityOrgEvents" (
			"id"           varchar(36)  PRIMARY KEY,
			"workspace_id" varchar(36)  NOT NULL,
			"actor_id"     varchar(36)  NOT NULL DEFAULT '',
			"action"       varchar(40)  NOT NULL,
			"subject_type" varchar(16)  NOT NULL,
			"subject_id"   varchar(36)  NOT NULL DEFAULT '',
			"detail"       jsonb        NOT NULL DEFAULT '{}'::jsonb,
			"created_at"   timestamptz  NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_org_events_ws" ON "IdentityOrgEvents" ("workspace_id", "created_at" DESC)`,
	}
	for _, s := range stmts {
		if _, err := q.Exec(s); err != nil {
			return err
		}
	}
	return nil
}
