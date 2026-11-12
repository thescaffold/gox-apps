package migrations

import "github.com/awesome-goose/goose/modules/sql"

// CreateIdentityKeyPolicies holds a workspace's maximum API key life (V0-14, pkg/apikey
// expiry.go); a workspace without a row has the default.
type CreateIdentityKeyPolicies struct{ sql.BaseMigration }

func (m *CreateIdentityKeyPolicies) Run(q *sql.Query) error {
	_, err := q.Exec(`CREATE TABLE IF NOT EXISTS "IdentityKeyPolicies" (
		"workspace_id" varchar(36)  PRIMARY KEY,
		"max_days"     integer      NOT NULL,
		"updated_by"   varchar(36)  NOT NULL DEFAULT '',
		"updated_at"   timestamptz  NOT NULL DEFAULT NOW()
	)`)
	return err
}
