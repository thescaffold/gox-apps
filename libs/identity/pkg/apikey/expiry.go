package apikey

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Expiry (V0-14, PLT-07). A host that sets Service.MaxTTL requires every key it issues to expire,
// no later than the workspace's maximum: a key with no expiry is refused with ErrExpiryRequired,
// one past the maximum with ErrExpiryTooLong. Without MaxTTL a key may still have no expiry.
//
// The maximum is a workspace setting (IdentityKeyPolicies): DefaultMaxDays unless the workspace
// chose a shorter one with SetMaxDays.

// DefaultMaxDays is the longest a key may live in a workspace that has set no maximum.
const DefaultMaxDays = 365

var (
	// ErrExpiryRequired: the key has no expiry and its host requires one.
	ErrExpiryRequired = errors.New("a key needs an expiry")
	// ErrExpiryTooLong: the key would outlive the workspace's maximum.
	ErrExpiryTooLong = errors.New("the expiry is past the organization's maximum")
)

// MaxTTL answers the longest a new key may live in a workspace. Zero or less means no rule.
type MaxTTL func(ctx context.Context, clientID, workspaceID string) (time.Duration, error)

// WorkspaceMaxTTL is the MaxTTL of the workspace's own maximum (MaxDays).
func WorkspaceMaxTTL(db *gorm.DB) MaxTTL {
	return func(ctx context.Context, _, workspaceID string) (time.Duration, error) {
		d, err := MaxDays(ctx, db, workspaceID)
		if err != nil {
			return 0, err
		}
		return time.Duration(d) * 24 * time.Hour, nil
	}
}

// MaxDays is the workspace's maximum key life in days: its setting, or DefaultMaxDays.
func MaxDays(ctx context.Context, db *gorm.DB, workspaceID string) (int, error) {
	var days []int
	if err := db.WithContext(ctx).Raw(`SELECT "max_days" FROM "IdentityKeyPolicies" WHERE "workspace_id" = ?`, workspaceID).Scan(&days).Error; err != nil {
		return 0, err
	}
	if len(days) == 0 || days[0] <= 0 {
		return DefaultMaxDays, nil
	}
	return days[0], nil
}

// SetMaxDays sets the workspace's maximum key life: 1 to DefaultMaxDays days. Keys already issued
// keep their expiry.
func SetMaxDays(ctx context.Context, db *gorm.DB, workspaceID, by string, days int) error {
	if workspaceID == "" {
		return ErrScopeRequired
	}
	if days < 1 || days > DefaultMaxDays {
		return fmt.Errorf("%w: the maximum is 1 to %d days", ErrInvalid, DefaultMaxDays)
	}
	return db.WithContext(ctx).Exec(`INSERT INTO "IdentityKeyPolicies" ("workspace_id", "max_days", "updated_by", "updated_at")
		VALUES (?, ?, ?, now())
		ON CONFLICT ("workspace_id") DO UPDATE SET "max_days" = EXCLUDED."max_days", "updated_by" = EXCLUDED."updated_by", "updated_at" = EXCLUDED."updated_at"`,
		workspaceID, days, by).Error
}

// checkExpiry applies MaxTTL to a new key's ttl.
func (s *Service) checkExpiry(ctx context.Context, clientID, workspaceID string, ttl time.Duration) error {
	if s.MaxTTL == nil {
		return nil
	}
	max, err := s.MaxTTL(ctx, clientID, workspaceID)
	if err != nil {
		return err
	}
	if max <= 0 {
		return nil
	}
	if ttl <= 0 {
		return ErrExpiryRequired
	}
	if ttl > max {
		return fmt.Errorf("%w (%d days)", ErrExpiryTooLong, int(max/(24*time.Hour)))
	}
	return nil
}
