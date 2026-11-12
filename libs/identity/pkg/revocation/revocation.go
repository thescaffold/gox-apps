// Package revocation ends sessions that a stateless access token would
// otherwise outlive. Each person has at most one row: tokens issued before
// its instant are refused. "Sign out everywhere" and closing an account set it.
//
// A token that names its device session ("sid", V0-15) is also refused once
// that session is signed out (its IdentityDeviceSessions row is gone), so
// signing one device out ends its tokens at once.
//
// Tokens carry their issue time in milliseconds (`iatMs`); a token without one
// predates this package and is treated as issued at time zero, so a
// revocation also ends every token minted before the feature existed.
package revocation

import (
	"context"
	"time"

	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
	"gorm.io/gorm"
)

// Store records revocations.
type Store struct{ DB *gorm.DB }

var store *Store

// SetStore registers the store used by Check.
func SetStore(s *Store) {
	store = s
	// Every module's default session guard asks Check, so a revoked session is refused everywhere.
	coreauth.SetSessionRevoked(Check)
}

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
	return s.RevokedIn(ctx, userID, "", iatMs)
}

// RevokedIn is Revoked for a token of device session sid ("" for none): it is
// also refused when that session has been signed out. One query either way.
func (s *Store) RevokedIn(ctx context.Context, userID, sid string, iatMs int64) bool {
	var r struct {
		After int64
		Ended bool
	}
	err := s.DB.WithContext(ctx).Raw(`SELECT
		COALESCE((SELECT "valid_after_ms" FROM "IdentityUserRevocations" WHERE "user_id" = ?), 0) AS after,
		(? <> '' AND NOT EXISTS (SELECT 1 FROM "IdentityDeviceSessions" WHERE "id" = ? AND "user_id" = ? AND "deleted_at" IS NULL)) AS ended`,
		userID, sid, sid, userID).Scan(&r).Error
	if err != nil {
		return false
	}
	return r.Ended || (r.After > 0 && iatMs < r.After)
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
	sid, _ := claims["sid"].(string)
	return store.RevokedIn(context.Background(), sub, sid, iat)
}
