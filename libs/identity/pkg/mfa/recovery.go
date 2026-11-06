package mfa

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"gorm.io/gorm"
)

// RecoveryCount is how many recovery codes a person holds at a time.
const RecoveryCount = 10

// The alphabet leaves out look-alikes (0/o, 1/l/i).
const recoveryAlphabet = "abcdefghjkmnpqrstuvwxyz23456789"

func newRecoveryCode() (string, error) {
	// Rejection sampling keeps every letter equally likely.
	limit := 256 - 256%len(recoveryAlphabet)
	out := make([]byte, 0, 11)
	b := make([]byte, 1)
	for n := 0; n < 10; {
		if _, err := rand.Read(b); err != nil {
			return "", err
		}
		if int(b[0]) >= limit {
			continue
		}
		if n == 5 {
			out = append(out, '-')
		}
		out = append(out, recoveryAlphabet[int(b[0])%len(recoveryAlphabet)])
		n++
	}
	return string(out), nil
}

func normalizeRecovery(code string) string {
	r := strings.NewReplacer("-", "", " ", "")
	return strings.ToLower(r.Replace(strings.TrimSpace(code)))
}

func hashRecovery(code string) string {
	sum := sha256.Sum256([]byte("identity-mfa-recovery:" + normalizeRecovery(code)))
	return hex.EncodeToString(sum[:])
}

// replaceRecovery discards the person's codes and issues RecoveryCount new
// ones. The codes are returned once and only their hashes are kept.
func (s *Service) replaceRecovery(ctx context.Context, tx *gorm.DB, userID string) ([]string, error) {
	if err := tx.WithContext(ctx).Where(`"user_id" = ?`, userID).Delete(&RecoveryCode{}).Error; err != nil {
		return nil, err
	}
	codes := make([]string, 0, RecoveryCount)
	rows := make([]RecoveryCode, 0, RecoveryCount)
	seen := map[string]bool{}
	for len(codes) < RecoveryCount {
		c, err := newRecoveryCode()
		if err != nil {
			return nil, err
		}
		h := hashRecovery(c)
		if seen[h] {
			continue
		}
		seen[h] = true
		codes = append(codes, c)
		rows = append(rows, RecoveryCode{ID: newID(), UserID: userID, CodeHash: h, CreatedAt: s.now()})
	}
	if err := tx.WithContext(ctx).Create(&rows).Error; err != nil {
		return nil, err
	}
	return codes, nil
}

// RegenerateRecovery replaces the person's recovery codes (the proof is
// checked by the caller, Prove). Only an enrolled person has codes.
func (s *Service) RegenerateRecovery(ctx context.Context, userID string) ([]string, error) {
	ok, err := s.Enrolled(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrNotEnrolled
	}
	var codes []string
	err = s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		codes, err = s.replaceRecovery(ctx, tx, userID)
		return err
	})
	return codes, err
}

// useRecovery spends one of the person's codes, exactly once.
func (s *Service) useRecovery(ctx context.Context, userID, code string) (bool, error) {
	if len(normalizeRecovery(code)) != 10 {
		return false, nil
	}
	res := s.DB.WithContext(ctx).Model(&RecoveryCode{}).
		Where(`"user_id" = ? AND "code_hash" = ? AND "used_at" IS NULL`, userID, hashRecovery(code)).
		Update("used_at", s.now())
	return res.RowsAffected == 1, res.Error
}
