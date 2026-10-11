// Package revocation ends sessions that a stateless access token would
// otherwise outlive. Each person has at most one row: tokens issued before
// its instant are refused. "Sign out everywhere" and closing an account set it.
//
// Tokens carry their issue time in milliseconds (`iatMs`); a token without one
// predates this package and is treated as issued at time zero, so a
// revocation also ends every token minted before the feature existed.
package revocation

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// Store records revocations.
type Store struct{ DB *gorm.DB }

var store *Store

// SetStore registers the store used by Check.
func SetStore(s *Store) { store = s }

// RevokeUser invalidates every access token issued to the person before now.
func (s *Store) RevokeUser(ctx context.Context, userID, reason string) error {
	now := time.Now().UTC()
	return s.DB.WithContext(ctx).Exec(`
		INSERT INTO "IdentityUserRevocations" ("user_id","valid_after_ms","reason","created_at")
		VALUES (?, ?, ?, ?)
		ON CONFLICT ("user_id") DO UPDATE SET "valid_after_ms" = EXCLUDED."valid_after_ms",
			"reason" = EXCLUDED."reason", "created_at" = EXCLUDED."created_at"`,
		userID, now.UnixMilli(), reason, now).Error
}

// Revoked reports whether a token issued at iatMs for userID is refused.
func (s *Store) Revoked(ctx context.Context, userID string, iatMs int64) bool {
	var after int64
	err := s.DB.WithContext(ctx).Raw(`SELECT "valid_after_ms" FROM "IdentityUserRevocations" WHERE "user_id" = ?`, userID).Scan(&after).Error
	return err == nil && after > 0 && iatMs < after
}

// RevokeUser uses the registered store; it reports whether one is configured.
func RevokeUser(ctx context.Context, userID, reason string) (bool, error) {
	if store == nil {
		return false, nil
	}
	return true, store.RevokeUser(ctx, userID, reason)
}

// Check is a core/auth Revoked hook over verified claims. With no store it
// refuses nothing, so it is safe to wire unconditionally.
func Check(claims map[string]any) bool {
	if store == nil {
		return false
	}
	sub, _ := claims["sub"].(string)
	if sub == "" {
		return false
	}
	var iat int64
	switch v := claims["iatMs"].(type) {
	case float64:
		iat = int64(v)
	case int64:
		iat = v
	}
	return store.Revoked(context.Background(), sub, iat)
}
