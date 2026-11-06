package mfa

import (
	"bytes"
	"context"
	"crypto/subtle"
	"encoding/base64"
	"image/png"
	"strings"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"gorm.io/gorm"

	"github.com/thescaffold/gox-apps/libs/identity/pkg/tokenstore"
)

// TOTP parameters: RFC 6238 defaults every authenticator app understands.
const (
	totpPeriod = 30
	totpDigits = otp.DigitsSix
	totpSkew   = 1 // one step either side, for clock drift
)

// TOTPSetup is what a person needs to add the account to their app. The
// secret is shown once; it is stored sealed.
type TOTPSetup struct {
	FactorID string `json:"factorId"`
	Secret   string `json:"secret"`
	URI      string `json:"uri"`
	QR       string `json:"qr"` // data:image/png;base64,…
}

func (s *Service) key() string {
	if s.Key != nil {
		return s.Key()
	}
	return tokenstore.Key()
}

// StartTOTP makes a new, unconfirmed authenticator-app factor. Any earlier
// unconfirmed one is discarded; a confirmed one stays until the new one is
// confirmed (reconfiguring never leaves the person without a factor).
func (s *Service) StartTOTP(ctx context.Context, userID, issuer, account string) (TOTPSetup, error) {
	k, err := totp.Generate(totp.GenerateOpts{Issuer: issuer, AccountName: account, Period: totpPeriod, Digits: totpDigits, Algorithm: otp.AlgorithmSHA1})
	if err != nil {
		return TOTPSetup{}, err
	}
	sealed, err := tokenstore.SealProviderToken(k.Secret(), s.key())
	if err != nil {
		return TOTPSetup{}, err
	}
	f := Factor{ID: newID(), UserID: userID, Kind: KindTOTP, Name: "Authenticator app", SecretSealed: &sealed, CreatedAt: s.now()}
	err = s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where(`"user_id" = ? AND "kind" = ? AND "confirmed_at" IS NULL`, userID, KindTOTP).Delete(&Factor{}).Error; err != nil {
			return err
		}
		return tx.Create(&f).Error
	})
	if err != nil {
		return TOTPSetup{}, err
	}
	img, err := k.Image(200, 200)
	if err != nil {
		return TOTPSetup{}, err
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return TOTPSetup{}, err
	}
	return TOTPSetup{
		FactorID: f.ID,
		Secret:   k.Secret(),
		URI:      k.URL(),
		QR:       "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()),
	}, nil
}

// ConfirmTOTP checks the first code from the new factor and makes it the
// person's authenticator app (replacing an earlier one). On a first factor it
// returns the person's new recovery codes.
func (s *Service) ConfirmTOTP(ctx context.Context, userID, factorID, code string) ([]string, error) {
	var codes []string
	err := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var f Factor
		if err := tx.Where(`"id" = ? AND "user_id" = ? AND "kind" = ? AND "confirmed_at" IS NULL`, factorID, userID, KindTOTP).
			Take(&f).Error; err != nil {
			return ErrNotFound
		}
		step, ok, err := s.match(f, code)
		if err != nil {
			return err
		}
		if !ok {
			return ErrInvalidCode
		}
		now := s.now()
		if err := tx.Model(&Factor{}).Where(`"id" = ?`, f.ID).
			Updates(map[string]any{"confirmed_at": now, "last_step": step, "last_used_at": now}).Error; err != nil {
			return err
		}
		if err := tx.Where(`"user_id" = ? AND "kind" = ? AND "id" <> ?`, userID, KindTOTP, f.ID).Delete(&Factor{}).Error; err != nil {
			return err
		}
		codes, err = s.afterFirstFactor(ctx, tx, userID)
		return err
	})
	return codes, err
}

// match checks code against the factor's secret within the skew and returns
// the time step it matched.
func (s *Service) match(f Factor, code string) (int64, bool, error) {
	code = strings.ReplaceAll(strings.TrimSpace(code), " ", "")
	if len(code) != 6 || f.SecretSealed == nil {
		return 0, false, nil
	}
	secret, err := tokenstore.OpenProviderToken(*f.SecretSealed, s.key())
	if err != nil {
		return 0, false, err
	}
	now := s.now()
	cur := now.Unix() / totpPeriod
	for d := int64(-totpSkew); d <= totpSkew; d++ {
		step := cur + d
		want, err := totp.GenerateCodeCustom(secret, time.Unix(step*totpPeriod, 0), totp.ValidateOpts{Period: totpPeriod, Digits: totpDigits, Algorithm: otp.AlgorithmSHA1})
		if err != nil {
			return 0, false, err
		}
		if subtle.ConstantTimeCompare([]byte(want), []byte(code)) == 1 {
			return step, true, nil
		}
	}
	return 0, false, nil
}

// useTOTP accepts a code from the person's confirmed authenticator app,
// exactly once: the step it matched must be later than the last one used.
func (s *Service) useTOTP(ctx context.Context, userID, code string) (bool, error) {
	var f Factor
	err := s.DB.WithContext(ctx).Where(`"user_id" = ? AND "kind" = ? AND "confirmed_at" IS NOT NULL`, userID, KindTOTP).Take(&f).Error
	if err != nil {
		return false, nil
	}
	step, ok, err := s.match(f, code)
	if err != nil || !ok {
		return false, err
	}
	res := s.DB.WithContext(ctx).Model(&Factor{}).Where(`"id" = ? AND "last_step" < ?`, f.ID, step).
		Updates(map[string]any{"last_step": step, "last_used_at": s.now()})
	return res.RowsAffected == 1, res.Error
}
