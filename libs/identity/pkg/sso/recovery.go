package sso

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Recovery is the way back in for a company that requires single sign-on and
// has lost its identity provider. Whoever controls the company's DNS is, by
// definition, the authority over its domain, so they prove it by publishing a
// TXT record we name. That lifts enforcement for their address for a short
// window; they must still sign in the ordinary way (an emailed code or a
// password), so DNS control alone does not hand over an account.

const (
	// RecoveryPrefix is the label published under the domain.
	RecoveryPrefix = "_envelope-sso-recovery"
	recoveryValue  = "envelope-sso-recovery="
	proofTTL       = time.Hour
	grantTTL       = 30 * time.Minute
	maxStartsPerH  = 5
)

// Errors a caller can match.
var (
	ErrRecoveryNotNeeded = errors.New("sso: this address does not require single sign-on")
	ErrRecoveryThrottled = errors.New("sso: too many recovery requests; try again later")
	ErrRecoveryPending   = errors.New("sso: no recovery in progress for this address")
	ErrRecoveryNotFound  = errors.New("sso: the DNS record was not found yet")
)

// RecoveryChallenge is what the person must publish.
type RecoveryChallenge struct {
	Host      string    `json:"host"`
	Type      string    `json:"type"`
	Value     string    `json:"value"`
	ExpiresAt time.Time `json:"expiresAt"`
}

func (s *Store) now() time.Time {
	if s.Now != nil {
		return s.Now().UTC()
	}
	return time.Now().UTC()
}

func (s *Store) lookupTXT(ctx context.Context, name string) ([]string, error) {
	if s.LookupTXT != nil {
		return s.LookupTXT(ctx, name)
	}
	return net.DefaultResolver.LookupTXT(ctx, name)
}

func hashProof(t string) string {
	h := sha256.Sum256([]byte(t))
	return hex.EncodeToString(h[:])
}

// StartRecovery issues a challenge for an address whose company requires SSO.
func (s *Store) StartRecovery(ctx context.Context, email string) (*RecoveryChallenge, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	conn, err := s.ForEmail(ctx, email)
	if err != nil || !conn.Enforce {
		return nil, ErrRecoveryNotNeeded
	}
	domain := EmailDomain(email)
	now := s.now()

	var recent int64
	if err := s.DB.WithContext(ctx).Table("IdentitySsoRecoveries").
		Where("domain = ? AND created_at > ?", domain, now.Add(-time.Hour)).Count(&recent).Error; err != nil {
		return nil, err
	}
	if recent >= maxStartsPerH {
		return nil, ErrRecoveryThrottled
	}
	token, err := Random(24)
	if err != nil {
		return nil, err
	}
	exp := now.Add(proofTTL)
	err = s.DB.WithContext(ctx).Table("IdentitySsoRecoveries").Create(map[string]any{
		"id": uuid.NewString(), "email": email, "domain": domain, "token_hash": hashProof(token),
		"expires_at": exp, "created_at": now,
	}).Error
	if err != nil {
		return nil, err
	}
	return &RecoveryChallenge{Host: RecoveryPrefix + "." + domain, Type: "TXT", Value: recoveryValue + token, ExpiresAt: exp}, nil
}

// VerifyRecovery checks the DNS record and, when it carries the challenge
// issued for this address, lifts enforcement for it for a short window.
func (s *Store) VerifyRecovery(ctx context.Context, email string) (time.Time, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	now := s.now()
	var row struct {
		ID        string
		Domain    string
		TokenHash string
	}
	err := s.DB.WithContext(ctx).Table("IdentitySsoRecoveries").
		Select("id, domain, token_hash").
		Where("email = ? AND expires_at > ? AND granted_until IS NULL", email, now).
		Order("created_at DESC").Limit(1).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return time.Time{}, ErrRecoveryPending
	}
	if err != nil {
		return time.Time{}, err
	}
	txts, err := s.lookupTXT(ctx, RecoveryPrefix+"."+row.Domain)
	if err != nil {
		return time.Time{}, ErrRecoveryNotFound
	}
	for _, t := range txts {
		t = strings.TrimSpace(strings.Trim(t, `"`))
		if !strings.HasPrefix(t, recoveryValue) {
			continue
		}
		got := hashProof(strings.TrimPrefix(t, recoveryValue))
		if subtle.ConstantTimeCompare([]byte(got), []byte(row.TokenHash)) == 1 {
			until := now.Add(grantTTL)
			if err := s.DB.WithContext(ctx).Table("IdentitySsoRecoveries").
				Where("id = ?", row.ID).Update("granted_until", until).Error; err != nil {
				return time.Time{}, err
			}
			return until, nil
		}
	}
	return time.Time{}, ErrRecoveryNotFound
}

// recoveryGranted reports whether the address holds a live recovery grant.
func (s *Store) recoveryGranted(ctx context.Context, email string) bool {
	var n int64
	err := s.DB.WithContext(ctx).Table("IdentitySsoRecoveries").
		Where("email = ? AND granted_until > ?", strings.ToLower(strings.TrimSpace(email)), s.now()).Count(&n).Error
	return err == nil && n > 0
}
