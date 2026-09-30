// Package tokenstore is how identity keeps credentials at rest (PLAN M1-03,
// TRD F7 / U-S3):
//
//   - refresh tokens are stored as sha256:<hex> — the server only ever needs
//     to recognise a presented token, never to read one back, so a database
//     leak yields nothing redeemable;
//   - provider OAuth tokens (which a later task must read back to call the
//     provider) are stored as AES-256-GCM envelopes (core/security).
package tokenstore

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"strings"

	"github.com/thescaffold/gox-packages/libs/core/security"
)

const hashPrefix = "sha256:"

// KeyEnv names the env var holding the 32-byte provider-token key.
const KeyEnv = "IDENTITY_TOKEN_KEY"

// HashRefresh is the at-rest form of a refresh token.
func HashRefresh(token string) string {
	h := sha256.Sum256([]byte(token))
	return hashPrefix + hex.EncodeToString(h[:])
}

// IsHashed reports whether s is already in HashRefresh form.
func IsHashed(s string) bool { return strings.HasPrefix(s, hashPrefix) }

// Key returns the configured provider-token key, or "" when unset.
func Key() string { return os.Getenv(KeyEnv) }

func checkKey(key string) error {
	if len(key) != 32 {
		return errors.New("tokenstore: " + KeyEnv + " must be exactly 32 bytes")
	}
	return nil
}

// SealProviderToken encrypts a provider token for storage. An empty token
// stays empty.
func SealProviderToken(plain, key string) (string, error) {
	if err := checkKey(key); err != nil {
		return "", err
	}
	return security.EncryptGCM(plain, key)
}

// IsSealed reports whether stored is an encrypted envelope.
func IsSealed(stored string) bool { return security.IsGCMEnvelope(stored) }

// OpenProviderToken decrypts a stored provider token. A legacy plaintext
// value (not an envelope) is returned unchanged so callers keep working until
// the EncryptProviderTokens migration has run; a tampered envelope or a wrong
// key is an error.
func OpenProviderToken(stored, key string) (string, error) {
	if !IsSealed(stored) {
		return stored, nil
	}
	if err := checkKey(key); err != nil {
		return "", err
	}
	return security.DecryptGCM(stored, key)
}
