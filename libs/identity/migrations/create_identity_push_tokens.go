package migrations

import "github.com/awesome-goose/goose/modules/sql"

// CreateIdentityPushTokens holds each app install's push token (U5,
// pkg/pushtoken): FCM for Android, APNs for iOS, Web Push for browsers. A
// token belongs to one install, so it is unique: registering it again (after
// a sign-in on that install by someone else) moves it.
type CreateIdentityPushTokens struct{ sql.BaseMigration }

func (m *CreateIdentityPushTokens) Run(q *sql.Query) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS "IdentityPushTokens" (
			"id"           varchar(36)   PRIMARY KEY,
			"user_id"      varchar(36)   NOT NULL,
			"device_id"    varchar(36)   NOT NULL DEFAULT '',
			"platform"     varchar(16)   NOT NULL,
			"provider"     varchar(16)   NOT NULL,
			"token"        varchar(4096) NOT NULL,
			"app_version"  varchar(64)   NOT NULL DEFAULT '',
			"locale"       varchar(35)   NOT NULL DEFAULT '',
			"created_at"   timestamptz   NOT NULL DEFAULT NOW(),
			"updated_at"   timestamptz   NOT NULL DEFAULT NOW(),
			CONSTRAINT "chk_identity_push_tokens_platform" CHECK ("platform" IN ('ios','android','web'))
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS "uq_identity_push_tokens_token" ON "IdentityPushTokens" ("token")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_push_tokens_user" ON "IdentityPushTokens" ("user_id")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_push_tokens_device" ON "IdentityPushTokens" ("device_id")`,
	}
	for _, s := range stmts {
		if _, err := q.Exec(s); err != nil {
			return err
		}
	}
	return nil
}
