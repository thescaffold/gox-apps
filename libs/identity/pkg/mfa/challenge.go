package mfa

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// Challenge purposes.
const (
	PurposeSignIn   = "signin"   // a sign-in waiting for its second factor
	PurposeStepUp   = "stepup"   // a signed-in person re-verifying before a sensitive change
	PurposeRegister = "register" // a security key being added
)

const (
	challengeTTL = 10 * time.Minute
	maxAttempts  = 5
)

// Challenge is a proof in progress.
type Challenge struct {
	ID         string          `gorm:"column:id;primaryKey"`
	TokenHash  string          `gorm:"column:token_hash"`
	Purpose    string          `gorm:"column:purpose"`
	UserID     string          `gorm:"column:user_id"`
	ClientID   string          `gorm:"column:client_id"`
	UserAgent  string          `gorm:"column:user_agent"`
	Session    json.RawMessage `gorm:"column:session;type:jsonb"`
	Attempts   int             `gorm:"column:attempts"`
	ExpiresAt  time.Time       `gorm:"column:expires_at"`
	ConsumedAt *time.Time      `gorm:"column:consumed_at"`
	CreatedAt  time.Time       `gorm:"column:created_at"`
}

func (Challenge) TableName() string { return "IdentityMfaChallenges" }

// Pending is what the first step of a sign-in answers with.
type Pending struct {
	Challenge string    `json:"challenge"`
	Methods   []string  `json:"methods"`
	ExpiresAt time.Time `json:"expiresAt"`
}

func hashToken(t string) string {
	sum := sha256.Sum256([]byte(t))
	return hex.EncodeToString(sum[:])
}

func newToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (s *Service) open(ctx context.Context, purpose, userID, clientID, userAgent string, session json.RawMessage) (string, time.Time, error) {
	tok, err := newToken()
	if err != nil {
		return "", time.Time{}, err
	}
	now := s.now()
	c := Challenge{
		ID: newID(), TokenHash: hashToken(tok), Purpose: purpose, UserID: userID, ClientID: clientID,
		UserAgent: userAgent, Session: session, ExpiresAt: now.Add(challengeTTL), CreatedAt: now,
	}
	// Old challenges are swept as new ones are made.
	_ = s.DB.WithContext(ctx).Where(`"expires_at" < ?`, now.Add(-time.Hour)).Delete(&Challenge{}).Error
	if err := s.DB.WithContext(ctx).Create(&c).Error; err != nil {
		return "", time.Time{}, err
	}
	return tok, c.ExpiresAt, nil
}

// StartSignIn opens the second step of a sign-in the first factor has passed.
// The host answers with the Pending instead of tokens.
func (s *Service) StartSignIn(ctx context.Context, userID, clientID, userAgent string) (Pending, error) {
	st, err := s.Status(ctx, userID)
	if err != nil {
		return Pending{}, err
	}
	tok, exp, err := s.open(ctx, PurposeSignIn, userID, clientID, userAgent, nil)
	if err != nil {
		return Pending{}, err
	}
	return Pending{Challenge: tok, Methods: st.Methods, ExpiresAt: exp}, nil
}

// StartStepUp opens a re-verification for a signed-in person.
func (s *Service) StartStepUp(ctx context.Context, userID string) (Pending, error) {
	st, err := s.Status(ctx, userID)
	if err != nil {
		return Pending{}, err
	}
	if !st.Enabled {
		return Pending{}, ErrNotEnrolled
	}
	tok, exp, err := s.open(ctx, PurposeStepUp, userID, "", "", nil)
	if err != nil {
		return Pending{}, err
	}
	return Pending{Challenge: tok, Methods: st.Methods, ExpiresAt: exp}, nil
}

// take finds a live challenge by token and counts the attempt. A challenge
// past its expiry, already used, or out of attempts is refused.
func (s *Service) take(ctx context.Context, token, purpose string) (Challenge, error) {
	var c Challenge
	err := s.DB.WithContext(ctx).Where(`"token_hash" = ? AND "purpose" = ?`, hashToken(token), purpose).Take(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Challenge{}, ErrChallengeExpired
	}
	if err != nil {
		return Challenge{}, err
	}
	if c.ConsumedAt != nil || !s.now().Before(c.ExpiresAt) {
		return Challenge{}, ErrChallengeExpired
	}
	res := s.DB.WithContext(ctx).Model(&Challenge{}).
		Where(`"id" = ? AND "attempts" < ? AND "consumed_at" IS NULL`, c.ID, maxAttempts).
		Update("attempts", gorm.Expr(`"attempts" + 1`))
	if res.Error != nil {
		return Challenge{}, res.Error
	}
	if res.RowsAffected != 1 {
		return Challenge{}, ErrTooManyAttempts
	}
	c.Attempts++
	return c, nil
}

// consume marks the challenge used; only one caller can.
func (s *Service) consume(ctx context.Context, id string) error {
	res := s.DB.WithContext(ctx).Model(&Challenge{}).Where(`"id" = ? AND "consumed_at" IS NULL`, id).Update("consumed_at", s.now())
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return ErrChallengeExpired
	}
	return nil
}

// Proof is a second factor: one of a code from the authenticator app, a
// recovery code, or a security key's assertion (the JSON the browser's
// navigator.credentials.get returned, for options from KeyOptions).
type Proof struct {
	Code         string          `json:"code"`
	RecoveryCode string          `json:"recoveryCode"`
	Assertion    json.RawMessage `json:"assertion"`
}

// Completed is a challenge that has been proven.
type Completed struct {
	UserID    string
	ClientID  string
	UserAgent string
	Method    string
}

// CompleteSignIn proves the second step of a sign-in. The host then issues
// tokens and creates the device session for Completed.UserID.
func (s *Service) CompleteSignIn(ctx context.Context, token string, p Proof) (Completed, error) {
	return s.complete(ctx, token, PurposeSignIn, "", p)
}

// Prove checks a re-verification for userID (from StartStepUp).
func (s *Service) Prove(ctx context.Context, userID, token string, p Proof) error {
	if token == "" {
		return ErrNoProof
	}
	_, err := s.complete(ctx, token, PurposeStepUp, userID, p)
	return err
}

func (s *Service) complete(ctx context.Context, token, purpose, userID string, p Proof) (Completed, error) {
	c, err := s.take(ctx, token, purpose)
	if err != nil {
		return Completed{}, err
	}
	if userID != "" && c.UserID != userID {
		return Completed{}, ErrChallengeExpired
	}
	method := ""
	ok := false
	switch {
	case len(p.Assertion) > 0 && string(p.Assertion) != "null":
		method = MethodWebAuthn
		ok, err = s.useKey(ctx, c, p.Assertion)
	case p.Code != "":
		method = MethodTOTP
		ok, err = s.useTOTP(ctx, c.UserID, p.Code)
	case p.RecoveryCode != "":
		method = MethodRecovery
		ok, err = s.useRecovery(ctx, c.UserID, p.RecoveryCode)
	default:
		return Completed{}, ErrNoProof
	}
	if err != nil {
		return Completed{}, err
	}
	if !ok {
		if c.Attempts >= maxAttempts {
			return Completed{}, ErrTooManyAttempts
		}
		return Completed{}, ErrInvalidCode
	}
	if err := s.consume(ctx, c.ID); err != nil {
		return Completed{}, err
	}
	return Completed{UserID: c.UserID, ClientID: c.ClientID, UserAgent: c.UserAgent, Method: method}, nil
}
