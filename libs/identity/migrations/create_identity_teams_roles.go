package migrations

import "github.com/awesome-goose/goose/modules/sql"

// CreateIdentityTeamsRoles adds a workspace's organisation: teams (a team, a
// group or a department, optionally nested), who is in them, company-defined
// roles holding permission strings, and who or which team holds each role.
// Additive: nothing existing is altered.
type CreateIdentityTeamsRoles struct{ sql.BaseMigration }

func (m *CreateIdentityTeamsRoles) Run(q *sql.Query) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS "IdentityTeams" (
			"id"           varchar(36)  PRIMARY KEY,
			"workspace_id" varchar(36)  NOT NULL,
			"parent_id"    varchar(36),
			"name"         varchar(80)  NOT NULL,
			"kind"         varchar(16)  NOT NULL CHECK ("kind" IN ('team','group','department')),
			"description"  varchar(500) NOT NULL DEFAULT '',
			"created_at"   timestamptz  NOT NULL DEFAULT NOW()
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS "idx_identity_teams_name" ON "IdentityTeams" ("workspace_id", lower("name"))`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_teams_parent" ON "IdentityTeams" ("parent_id")`,
		`CREATE TABLE IF NOT EXISTS "IdentityTeamMembers" (
			"team_id"      varchar(36) NOT NULL,
			"user_id"      varchar(36) NOT NULL,
			"workspace_id" varchar(36) NOT NULL,
			"role"         varchar(16) NOT NULL DEFAULT 'member' CHECK ("role" IN ('lead','member')),
			"created_at"   timestamptz NOT NULL DEFAULT NOW(),
			PRIMARY KEY ("team_id", "user_id")
		)`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_team_members_user" ON "IdentityTeamMembers" ("workspace_id", "user_id")`,
		`CREATE TABLE IF NOT EXISTS "IdentityCustomRoles" (
			"id"           varchar(36)  PRIMARY KEY,
			"workspace_id" varchar(36)  NOT NULL,
			"name"         varchar(80)  NOT NULL,
			"description"  varchar(500) NOT NULL DEFAULT '',
			"permissions"  jsonb        NOT NULL DEFAULT '[]'::jsonb,
			"created_at"   timestamptz  NOT NULL DEFAULT NOW()
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS "idx_identity_custom_roles_name" ON "IdentityCustomRoles" ("workspace_id", lower("name"))`,
		`CREATE TABLE IF NOT EXISTS "IdentityRoleAssignments" (
			"role_id"      varchar(36) NOT NULL,
			"subject_type" varchar(8)  NOT NULL CHECK ("subject_type" IN ('user','team')),
			"subject_id"   varchar(36) NOT NULL,
			"workspace_id" varchar(36) NOT NULL,
			"created_at"   timestamptz NOT NULL DEFAULT NOW(),
			PRIMARY KEY ("role_id", "subject_type", "subject_id")
		)`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_role_assignments_subject" ON "IdentityRoleAssignments" ("workspace_id", "subject_type", "subject_id")`,
	}
	for _, s := range stmts {
		if _, err := q.Exec(s); err != nil {
			return err
		}
	}
	return nil
}
