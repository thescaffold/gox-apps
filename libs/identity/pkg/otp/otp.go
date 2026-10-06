// Package otp issues and checks one-time sign-in codes for passwordless login.
//
// Codes are six digits, stored only as a hash, expire after ten minutes, are
// single use, and are limited three ways: a minimum gap between sends, a cap on
// sends per hour, and a cap on wrong guesses per code. A deployment can set a
// "magic" code for sandboxes and tests; it is refused in production.
package otp

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"os"
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Errors a caller can match.
var (
	ErrThrottled = errors.New("a code was sent recently; wait before asking for another")
	ErrTooMany   = errors.New("too many codes requested; try again later")
	ErrInvalid   = errors.New("the code is not valid")
	ErrExpired   = errors.New("the code has expired")
	ErrLocked    = errors.New("too many wrong attempts; request a new code")
	ErrBadRef    = errors.New("enter a valid email address or phone number")
)

// Types of reference.
const (
	Email = "email"
	Phone = "phone"
)

var (
	emailRe = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)
	phoneRe = regexp.MustCompile(`^\+[1-9][0-9]{7,14}$`)
)

// Normalize canonicalises a sign-in reference and says what it is: emails are
// lower-cased; phones become E.164 (Nigerian national numbers are accepted).
func Normalize(ref string) (normalized, typ string, err error) {
	r := strings.TrimSpace(ref)
	if strings.Contains(r, "@") {
		r = strings.ToLower(r)
		if !emailRe.MatchString(r) {
			return "", "", ErrBadRef
		}
		return r, Email, nil
	}
	digits := strings.NewReplacer(" ", "", "-", "", "(", "", ")", "", ".", "").Replace(r)
	switch {
	case strings.HasPrefix(digits, "+"):
	case strings.HasPrefix(digits, "00"):
		digits = "+" + digits[2:]
	case strings.HasPrefix(digits, "234"):
		digits = "+" + digits
	case strings.HasPrefix(digits, "0") && len(digits) == 11:
		digits = "+234" + digits[1:]
	default:
		return "", "", ErrBadRef
	}
	if !phoneRe.MatchString(digits) {
		return "", "", ErrBadRef
	}
	return digits, Phone, nil
}

// Service issues and checks codes.
type Service struct {
	DB *gorm.DB
	// TTL is how long a code lives (default 10m).
	TTL time.Duration
	// MinGap is the least time between two sends to one reference (default 30s).
	MinGap time.Duration
	// MaxPerHour caps sends to one reference (default 5).
	MaxPerHour int
	// MaxAttempts caps wrong guesses per code (default 5).
	MaxAttempts int
	// Magic is a fixed code that always verifies (sandbox/test only; ignored in production).
	Magic string
	Now   func() time.Time
}

func (s *Service) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return time.Now()
}

func (s *Service) ttl() time.Duration {
	if s.TTL > 0 {
		return s.TTL
	}
	return 10 * time.Minute
}

func (s *Service) gap() time.Duration {
	if s.MinGap > 0 {
		return s.MinGap
	}
	return 30 * time.Second
}

func (s *Service) perHour() int {
	if s.MaxPerHour > 0 {
		return s.MaxPerHour
	}
	return 5
}

func (s *Service) attempts() int {
	if s.MaxAttempts > 0 {
		return s.MaxAttempts
	}
	return 5
}

func production() bool {
	for _, k := range []string{"NODE_ENV", "APP_ENV"} {
		if strings.EqualFold(os.Getenv(k), "production") {
			return true
		}
	}
	return false
}

func (s *Service) magic() string {
	if production() {
		return ""
	}
	if s.Magic != "" {
		return s.Magic
	}
	return os.Getenv("IDENTITY_MAGIC_OTP")
}

func hashCode(ref, clientID, code string) string {
	h := sha256.Sum256([]byte(ref + "\x00" + clientID + "\x00" + code))
	return hex.EncodeToString(h[:])
}

func newCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

// Issue creates a code for ref (already normalised) and returns it for sending.
// Older unused codes for the reference stop working.
func (s *Service) Issue(ctx context.Context, ref, clientID string) (string, error) {
	var recent struct {
		Last  *time.Time
		Count int
	}
	if err := s.DB.WithContext(ctx).Raw(`SELECT MAX(created_at) AS last, COUNT(*) AS count FROM "IdentityOtps"
		WHERE ref = ? AND client_id = ? AND created_at > ?`, ref, clientID, s.now().Add(-time.Hour)).Scan(&recent).Error; err != nil {
		return "", err
	}
	if recent.Count >= s.perHour() {
		return "", ErrTooMany
	}
	if recent.Last != nil && s.now().Sub(*recent.Last) < s.gap() {
		return "", ErrThrottled
	}
	code, err := newCode()
	if err != nil {
		return "", err
	}
	err = s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`UPDATE "IdentityOtps" SET consumed_at = ? WHERE ref = ? AND client_id = ? AND consumed_at IS NULL`, s.now(), ref, clientID).Error; err != nil {
			return err
		}
		return tx.Exec(`INSERT INTO "IdentityOtps" (ref, client_id, code_hash, created_at, expires_at) VALUES (?, ?, ?, ?, ?)`,
			ref, clientID, hashCode(ref, clientID, code), s.now(), s.now().Add(s.ttl())).Error
	})
	if err != nil {
		return "", err
	}
	return code, nil
}

// Check verifies a code. A correct code is consumed; a wrong one counts against
// the attempt cap, and past the cap the code is dead.
func (s *Service) Check(ctx context.Context, ref, clientID, code string) error {
	code = strings.TrimSpace(code)
	if m := s.magic(); m != "" && subtle.ConstantTimeCompare([]byte(code), []byte(m)) == 1 {
		return nil
	}
	var result error
	err := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row struct {
			ID        int64
			CodeHash  string
			Attempts  int
			ExpiresAt time.Time
			Found     bool
		}
		if err := tx.Raw(`SELECT id, code_hash, attempts, expires_at, true AS found FROM "IdentityOtps"
			WHERE ref = ? AND client_id = ? AND consumed_at IS NULL ORDER BY id DESC LIMIT 1 FOR UPDATE`, ref, clientID).Scan(&row).Error; err != nil {
			return err
		}
		if !row.Found {
			return ErrInvalid
		}
		if !s.now().Before(row.ExpiresAt) {
			return ErrExpired
		}
		if row.Attempts >= s.attempts() {
			return ErrLocked
		}
		if subtle.ConstantTimeCompare([]byte(row.CodeHash), []byte(hashCode(ref, clientID, code))) != 1 {
			// the failed guess must be counted, so this transaction commits
			// and the error is returned afterwards
			result = ErrInvalid
			return tx.Exec(`UPDATE "IdentityOtps" SET attempts = attempts + 1 WHERE id = ?`, row.ID).Error
		}
		return tx.Exec(`UPDATE "IdentityOtps" SET consumed_at = ? WHERE id = ?`, s.now(), row.ID).Error
	})
	if err != nil {
		return err
	}
	return result
}
