// Package mfa is two-factor sign-in (U-S20): an authenticator app (TOTP) and
// security keys (WebAuthn) as second factors, one-time recovery codes, and a
// workspace policy that requires 2FA of its members.
//
// It is off unless someone enrolls. A person with a confirmed factor finishes
// every sign-in (password, emailed code, single sign-on, a code-hosting
// provider) with a second step: the first step answers with a short-lived,
// single-use challenge instead of tokens, and no device session is created
// until the second factor is proven, so the device's account chooser cannot
// skip it.
//
// A workspace that requires 2FA does not stop its members signing in (they
// have to, to enroll); the host calls Gate on its own routes and refuses until
// they have.
package mfa

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Factor kinds.
const (
	KindTOTP     = "totp"
	KindWebAuthn = "webauthn"
)

// Methods a second step can be completed with.
const (
	MethodTOTP     = "totp"
	MethodWebAuthn = "webauthn"
	MethodRecovery = "recovery"
)

// Errors a caller can match. Each reads as a sentence a person can act on.
var (
	ErrInvalidCode         = errors.New("that code is not valid; check your authenticator app and try again")
	ErrChallengeExpired    = errors.New("this sign-in has expired; sign in again")
	ErrTooManyAttempts     = errors.New("too many attempts; sign in again")
	ErrNotEnrolled         = errors.New("two-factor authentication is not set up")
	ErrNotFound            = errors.New("not found")
	ErrRequiredByWorkspace = errors.New("an organization you belong to requires two-factor authentication, so your last factor cannot be removed")
	ErrEnrollmentRequired  = errors.New("this organization requires two-factor authentication; set it up under Account > Password and 2FA")
	ErrEnrollFirst         = errors.New("set up two-factor authentication for yourself before requiring it of others")
	ErrKeysUnavailable     = errors.New("security keys are not configured on this server")
	ErrNoProof             = errors.New("confirm with a code from your authenticator app, a security key or a recovery code")
)

// Factor is one second factor.
type Factor struct {
	ID           string          `gorm:"column:id;primaryKey"`
	UserID       string          `gorm:"column:user_id"`
	Kind         string          `gorm:"column:kind"`
	Name         string          `gorm:"column:name"`
	SecretSealed *string         `gorm:"column:secret_sealed"`
	CredentialID *string         `gorm:"column:credential_id"`
	Credential   json.RawMessage `gorm:"column:credential;type:jsonb"`
	LastStep     int64           `gorm:"column:last_step"`
	ConfirmedAt  *time.Time      `gorm:"column:confirmed_at"`
	LastUsedAt   *time.Time      `gorm:"column:last_used_at"`
	CreatedAt    time.Time       `gorm:"column:created_at"`
}

func (Factor) TableName() string { return "IdentityMfaFactors" }

// RecoveryCode is one one-time code, as its hash.
type RecoveryCode struct {
	ID        string     `gorm:"column:id;primaryKey"`
	UserID    string     `gorm:"column:user_id"`
	CodeHash  string     `gorm:"column:code_hash"`
	UsedAt    *time.Time `gorm:"column:used_at"`
	CreatedAt time.Time  `gorm:"column:created_at"`
}

func (RecoveryCode) TableName() string { return "IdentityMfaRecoveryCodes" }

// Policy is a workspace's requirement.
type Policy struct {
	WorkspaceID string    `gorm:"column:workspace_id;primaryKey"`
	Require     bool      `gorm:"column:require"`
	UpdatedBy   string    `gorm:"column:updated_by"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (Policy) TableName() string { return "IdentityMfaPolicies" }

// Service is the store and the rules. DB is required; Key seals TOTP secrets
// (IDENTITY_TOKEN_KEY by default); Now is a seam for tests.
type Service struct {
	DB  *gorm.DB
	Key func() string
	Now func() time.Time
	// Keys is the WebAuthn relying party; nil means security keys are off.
	Keys *Keys
}

func (s *Service) now() time.Time {
	if s.Now != nil {
		return s.Now().UTC()
	}
	return time.Now().UTC()
}

func newID() string { return uuid.NewString() }

// KeyView is a security key as a person sees it.
type KeyView struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	AddedAt    time.Time  `json:"addedAt"`
	LastUsedAt *time.Time `json:"lastUsedAt,omitempty"`
}

// Status is a person's two-factor setup.
type Status struct {
	Enabled       bool       `json:"enabled"`
	TOTP          bool       `json:"totp"`
	TOTPAddedAt   *time.Time `json:"totpAddedAt,omitempty"`
	TOTPFactorID  string     `json:"totpFactorId,omitempty"`
	Keys          []KeyView  `json:"keys"`
	RecoveryLeft  int        `json:"recoveryLeft"`
	RecoveryTotal int        `json:"recoveryTotal"`
	KeysAvailable bool       `json:"keysAvailable"`
	Methods       []string   `json:"methods"`
}

func (s *Service) confirmed(ctx context.Context, userID string) ([]Factor, error) {
	var fs []Factor
	err := s.DB.WithContext(ctx).Where(`"user_id" = ? AND "confirmed_at" IS NOT NULL`, userID).
		Order(`"created_at"`).Find(&fs).Error
	return fs, err
}

// Status reads a person's setup.
func (s *Service) Status(ctx context.Context, userID string) (Status, error) {
	fs, err := s.confirmed(ctx, userID)
	if err != nil {
		return Status{}, err
	}
	st := Status{Keys: []KeyView{}, KeysAvailable: s.Keys != nil, Methods: []string{}}
	for _, f := range fs {
		switch f.Kind {
		case KindTOTP:
			st.TOTP = true
			st.TOTPAddedAt = f.ConfirmedAt
			st.TOTPFactorID = f.ID
		case KindWebAuthn:
			st.Keys = append(st.Keys, KeyView{ID: f.ID, Name: f.Name, AddedAt: *f.ConfirmedAt, LastUsedAt: f.LastUsedAt})
		}
	}
	st.Enabled = len(fs) > 0
	var total, left int64
	if err := s.DB.WithContext(ctx).Model(&RecoveryCode{}).Where(`"user_id" = ?`, userID).Count(&total).Error; err != nil {
		return Status{}, err
	}
	if err := s.DB.WithContext(ctx).Model(&RecoveryCode{}).Where(`"user_id" = ? AND "used_at" IS NULL`, userID).Count(&left).Error; err != nil {
		return Status{}, err
	}
	st.RecoveryTotal, st.RecoveryLeft = int(total), int(left)
	if st.TOTP {
		st.Methods = append(st.Methods, MethodTOTP)
	}
	if len(st.Keys) > 0 && s.Keys != nil {
		st.Methods = append(st.Methods, MethodWebAuthn)
	}
	if st.Enabled && left > 0 {
		st.Methods = append(st.Methods, MethodRecovery)
	}
	return st, nil
}

// Enrolled reports whether the person has a confirmed second factor.
func (s *Service) Enrolled(ctx context.Context, userID string) (bool, error) {
	var n int64
	err := s.DB.WithContext(ctx).Model(&Factor{}).Where(`"user_id" = ? AND "confirmed_at" IS NOT NULL`, userID).Count(&n).Error
	return n > 0, err
}

// EnrolledAmong returns which of userIDs have a confirmed factor.
func (s *Service) EnrolledAmong(ctx context.Context, userIDs []string) (map[string]bool, error) {
	out := map[string]bool{}
	if len(userIDs) == 0 {
		return out, nil
	}
	var ids []string
	err := s.DB.WithContext(ctx).Model(&Factor{}).Where(`"user_id" IN ? AND "confirmed_at" IS NOT NULL`, userIDs).
		Distinct("user_id").Pluck("user_id", &ids).Error
	for _, id := range ids {
		out[id] = true
	}
	return out, err
}

// Required reports whether the workspace requires 2FA of its members.
func (s *Service) Required(ctx context.Context, workspaceID string) (bool, error) {
	if workspaceID == "" {
		return false, nil
	}
	var ps []Policy
	err := s.DB.WithContext(ctx).Where(`"workspace_id" = ?`, workspaceID).Limit(1).Find(&ps).Error
	return len(ps) == 1 && ps[0].Require, err
}

// SetRequired turns the workspace's requirement on or off. Whoever turns it on
// must be enrolled themselves (callers check that they may manage the
// workspace).
func (s *Service) SetRequired(ctx context.Context, workspaceID, by string, require bool) error {
	if require {
		ok, err := s.Enrolled(ctx, by)
		if err != nil {
			return err
		}
		if !ok {
			return ErrEnrollFirst
		}
	}
	return s.DB.WithContext(ctx).Exec(`INSERT INTO "IdentityMfaPolicies" ("workspace_id","require","updated_by","updated_at")
		VALUES (?,?,?,?) ON CONFLICT ("workspace_id") DO UPDATE SET "require" = EXCLUDED."require",
		"updated_by" = EXCLUDED."updated_by", "updated_at" = EXCLUDED."updated_at"`,
		workspaceID, require, by, s.now()).Error
}

// Gate refuses (ErrEnrollmentRequired) a person who has not set up 2FA in a
// workspace that requires it. Hosts call it on their own routes.
func (s *Service) Gate(ctx context.Context, userID, workspaceID string) error {
	req, err := s.Required(ctx, workspaceID)
	if err != nil || !req {
		return err
	}
	ok, err := s.Enrolled(ctx, userID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrEnrollmentRequired
	}
	return nil
}

// requiredSomewhere reports whether any workspace the person belongs to
// requires 2FA.
func (s *Service) requiredSomewhere(ctx context.Context, userID string) (bool, error) {
	var n int64
	err := s.DB.WithContext(ctx).Raw(`SELECT COUNT(*) FROM "IdentityMfaPolicies" p
		JOIN "IdentityUserClientWorkspaces" u ON u."workspace_id" = p."workspace_id"
		WHERE p."require" AND u."user_id" = ? AND u."deleted_at" IS NULL`, userID).Scan(&n).Error
	return n > 0, err
}

// RemoveFactor deletes one of the person's factors. The proof is checked by
// the caller (Prove). The last factor cannot go while a workspace requires
// 2FA; when it goes, so do the recovery codes.
func (s *Service) RemoveFactor(ctx context.Context, userID, factorID string) error {
	fs, err := s.confirmed(ctx, userID)
	if err != nil {
		return err
	}
	found := false
	for _, f := range fs {
		if f.ID == factorID {
			found = true
		}
	}
	if !found {
		return ErrNotFound
	}
	if len(fs) == 1 {
		req, err := s.requiredSomewhere(ctx, userID)
		if err != nil {
			return err
		}
		if req {
			return ErrRequiredByWorkspace
		}
	}
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where(`"id" = ? AND "user_id" = ?`, factorID, userID).Delete(&Factor{}).Error; err != nil {
			return err
		}
		if len(fs) == 1 {
			return tx.Where(`"user_id" = ?`, userID).Delete(&RecoveryCode{}).Error
		}
		return nil
	})
}

// afterFirstFactor gives the person recovery codes when they have none (their
// first factor was just confirmed). It returns the new codes, or nil.
func (s *Service) afterFirstFactor(ctx context.Context, tx *gorm.DB, userID string) ([]string, error) {
	var n int64
	if err := tx.WithContext(ctx).Model(&RecoveryCode{}).Where(`"user_id" = ?`, userID).Count(&n).Error; err != nil {
		return nil, err
	}
	if n > 0 {
		return nil, nil
	}
	return s.replaceRecovery(ctx, tx, userID)
}
