package migrations

import "github.com/awesome-goose/goose/modules/sql"

// CreateIdentityMfa backs two-factor sign-in (pkg/mfa):
//
//   - IdentityMfaFactors: a person's second factors. An authenticator app
//     keeps its TOTP secret sealed with IDENTITY_TOKEN_KEY and the last time
//     step it accepted (a code is never accepted twice); a security key keeps
//     its WebAuthn credential. A factor counts once confirmed.
//   - IdentityMfaRecoveryCodes: one-time codes, kept only as hashes.
//   - IdentityMfaChallenges: short-lived, single-use proofs in progress (a
//     sign-in waiting for its second factor, a key being registered, a
//     re-verification). Only a hash of the token is kept.
//   - IdentityMfaPolicies: a workspace that requires 2FA of its members.
type CreateIdentityMfa struct{ sql.BaseMigration }

func (m *CreateIdentityMfa) Run(q *sql.Query) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS "IdentityMfaFactors" (
			"id"            varchar(36)  PRIMARY KEY,
			"user_id"       varchar(36)  NOT NULL,
			"kind"          varchar(16)  NOT NULL,
			"name"          varchar(120) NOT NULL DEFAULT '',
			"secret_sealed" text,
			"credential_id" varchar(1400),
			"credential"    jsonb,
			"last_step"     bigint       NOT NULL DEFAULT 0,
			"confirmed_at"  timestamptz,
			"last_used_at"  timestamptz,
			"created_at"    timestamptz  NOT NULL DEFAULT NOW(),
			CONSTRAINT "chk_identity_mfa_factor_kind" CHECK ("kind" IN ('totp','webauthn'))
		)`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_mfa_factors_user" ON "IdentityMfaFactors" ("user_id")`,
		`CREATE UNIQUE INDEX IF NOT EXISTS "uq_identity_mfa_factors_credential" ON "IdentityMfaFactors" ("credential_id") WHERE "credential_id" IS NOT NULL`,
		`CREATE TABLE IF NOT EXISTS "IdentityMfaRecoveryCodes" (
			"id"         varchar(36) PRIMARY KEY,
			"user_id"    varchar(36) NOT NULL,
			"code_hash"  varchar(64) NOT NULL,
			"used_at"    timestamptz,
			"created_at" timestamptz NOT NULL DEFAULT NOW()
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS "uq_identity_mfa_recovery_code" ON "IdentityMfaRecoveryCodes" ("user_id", "code_hash")`,
		`CREATE TABLE IF NOT EXISTS "IdentityMfaChallenges" (
			"id"          varchar(36)  PRIMARY KEY,
			"token_hash"  varchar(64)  NOT NULL,
			"purpose"     varchar(24)  NOT NULL,
			"user_id"     varchar(36)  NOT NULL,
			"client_id"   varchar(255) NOT NULL DEFAULT '',
			"user_agent"  text         NOT NULL DEFAULT '',
			"session"     jsonb,
			"attempts"    integer      NOT NULL DEFAULT 0,
			"expires_at"  timestamptz  NOT NULL,
			"consumed_at" timestamptz,
			"created_at"  timestamptz  NOT NULL DEFAULT NOW()
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS "uq_identity_mfa_challenge_token" ON "IdentityMfaChallenges" ("token_hash")`,
		`CREATE INDEX IF NOT EXISTS "idx_identity_mfa_challenges_expiry" ON "IdentityMfaChallenges" ("expires_at")`,
		`CREATE TABLE IF NOT EXISTS "IdentityMfaPolicies" (
			"workspace_id" varchar(36) PRIMARY KEY,
			"require"      boolean     NOT NULL DEFAULT false,
			"updated_by"   varchar(36) NOT NULL DEFAULT '',
			"updated_at"   timestamptz NOT NULL DEFAULT NOW()
		)`,
	}
	for _, s := range stmts {
		if _, err := q.Exec(s); err != nil {
			return err
		}
	}
	return nil
}
