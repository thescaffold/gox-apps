package migrations

import (
	"errors"

	"github.com/awesome-goose/goose/modules/sql"
	"github.com/thescaffold/gox-apps/libs/identity/pkg/tokenstore"
)

// HashRefreshTokens rewrites every plaintext refresh token to sha256:<hex>
// (PLAN M1-03). Idempotent: already-hashed rows are skipped.
type HashRefreshTokens struct{ sql.BaseMigration }

func (m *HashRefreshTokens) Run(q *sql.Query) error {
	_, err := q.Exec(`
		UPDATE "IdentityTokens"
		   SET "token" = 'sha256:' || encode(sha256(convert_to("token", 'UTF8')), 'hex')
		 WHERE "type" = 'refresh' AND "token" NOT LIKE 'sha256:%'`)
	return err
}

// EncryptProviderTokens seals every plaintext IdentityProviders.token with
// AES-GCM. Key defaults to IDENTITY_TOKEN_KEY. If plaintext tokens exist and
// no valid key is configured it FAILS rather than reporting success with
// plaintext still in the table. Idempotent.
type EncryptProviderTokens struct {
	sql.BaseMigration
	Key string
}

func (m *EncryptProviderTokens) Run(q *sql.Query) error {
	key := m.Key
	if key == "" {
		key = tokenstore.Key()
	}
	rows, err := q.Raw(`SELECT id, token FROM "IdentityProviders" WHERE token IS NOT NULL AND token <> '' AND token NOT LIKE 'v1:%'`)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}
	if len(key) != 32 {
		return errors.New("identity: plaintext provider tokens exist but " + tokenstore.KeyEnv + " is not a 32-byte key; refusing to leave them unencrypted")
	}
	for _, r := range rows {
		id, _ := r["id"].(string)
		plain, _ := r["token"].(string)
		sealed, err := tokenstore.SealProviderToken(plain, key)
		if err != nil {
			return err
		}
		if _, err := q.Exec(`UPDATE "IdentityProviders" SET token = ? WHERE id = ?`, sealed, id); err != nil {
			return err
		}
	}
	return nil
}

// ScrubProviderLogTokens removes the OAuth access token the old code wrote
// into IdentityProviderLogs.request.
type ScrubProviderLogTokens struct{ sql.BaseMigration }

func (m *ScrubProviderLogTokens) Run(q *sql.Query) error {
	_, err := q.Exec(`UPDATE "IdentityProviderLogs" SET request = request - 'accessToken' WHERE request ? 'accessToken'`)
	return err
}
