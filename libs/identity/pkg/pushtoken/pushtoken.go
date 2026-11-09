// Package pushtoken keeps each app install's push token (PLAN U5, TRD
// §7.5.5): a native app PUTs {platform, token, appVersion, locale} on sign-in
// and whenever the OS rotates the token, and the token goes on sign-out. The
// notification side reads For(userID) to reach a person's devices.
//
// A token identifies one install, so it is unique across people: when a
// different person signs in on the same install and registers, the token
// moves to them (the previous person stops getting that install's pushes).
package pushtoken

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Platforms and the push service each uses.
const (
	PlatformIOS     = "ios"
	PlatformAndroid = "android"
	PlatformWeb     = "web"
)

var providers = map[string]string{PlatformIOS: "apns", PlatformAndroid: "fcm", PlatformWeb: "webpush"}

// Errors a caller can tell apart.
var (
	ErrNoUser   = errors.New("pushtoken: a signed-in person is required")
	ErrPlatform = errors.New("pushtoken: platform must be ios, android or web")
	ErrToken    = errors.New("pushtoken: token is required (at most 4096 characters)")
	ErrField    = errors.New("pushtoken: appVersion is at most 64 characters and locale at most 35")
)

// Token is one install's registration. The token itself is not echoed back.
type Token struct {
	Id         string    `gorm:"column:id;primaryKey"  json:"id"`
	UserId     string    `gorm:"column:user_id"        json:"-"`
	DeviceId   string    `gorm:"column:device_id"      json:"deviceId,omitempty"`
	Platform   string    `gorm:"column:platform"       json:"platform"`
	Provider   string    `gorm:"column:provider"       json:"provider"`
	Token      string    `gorm:"column:token"          json:"-"`
	AppVersion string    `gorm:"column:app_version"    json:"appVersion"`
	Locale     string    `gorm:"column:locale"         json:"locale"`
	CreatedAt  time.Time `gorm:"column:created_at"     json:"createdAt"`
	UpdatedAt  time.Time `gorm:"column:updated_at"     json:"updatedAt"`
}

func (Token) TableName() string { return "IdentityPushTokens" }

// Registration is what an app sends.
type Registration struct {
	Platform   string `json:"platform"`
	Token      string `json:"token"`
	AppVersion string `json:"appVersion"`
	Locale     string `json:"locale"`
	// DeviceId optionally names the identity Device of this install, so that
	// signing that device out removes its token too.
	DeviceId string `json:"deviceId"`
}

// Service is the store.
type Service struct {
	DB  *gorm.DB
	Now func() time.Time
}

// New is a store over db.
func New(db *gorm.DB) *Service { return &Service{DB: db} }

func (s *Service) now() time.Time {
	if s.Now != nil {
		return s.Now().UTC()
	}
	return time.Now().UTC()
}

// Register records (or refreshes) userID's registration of r.Token.
func (s *Service) Register(ctx context.Context, userID string, r Registration) (Token, error) {
	if userID == "" {
		return Token{}, ErrNoUser
	}
	platform := strings.ToLower(strings.TrimSpace(r.Platform))
	provider, ok := providers[platform]
	if !ok {
		return Token{}, ErrPlatform
	}
	tok := strings.TrimSpace(r.Token)
	if tok == "" || len(tok) > 4096 {
		return Token{}, ErrToken
	}
	version, locale := strings.TrimSpace(r.AppVersion), strings.TrimSpace(r.Locale)
	if utf8.RuneCountInString(version) > 64 || utf8.RuneCountInString(locale) > 35 {
		return Token{}, ErrField
	}
	deviceID := strings.TrimSpace(r.DeviceId)
	if len(deviceID) > 36 {
		deviceID = ""
	}
	now := s.now()
	row := Token{
		Id: uuid.NewString(), UserId: userID, DeviceId: deviceID, Platform: platform, Provider: provider,
		Token: tok, AppVersion: version, Locale: locale, CreatedAt: now, UpdatedAt: now,
	}
	err := s.DB.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "token"}},
		DoUpdates: clause.AssignmentColumns([]string{"user_id", "device_id", "platform", "provider", "app_version", "locale", "updated_at"}),
	}).Create(&row).Error
	if err != nil {
		return Token{}, err
	}
	var saved Token
	if err := s.DB.WithContext(ctx).Where(`"token" = ?`, tok).First(&saved).Error; err != nil {
		return Token{}, err
	}
	return saved, nil
}

// Deregister removes userID's registration of token (another person's is
// left alone) and says whether there was one.
func (s *Service) Deregister(ctx context.Context, userID, token string) (bool, error) {
	if userID == "" {
		return false, ErrNoUser
	}
	res := s.DB.WithContext(ctx).Where(`"user_id" = ? AND "token" = ?`, userID, strings.TrimSpace(token)).Delete(&Token{})
	return res.RowsAffected > 0, res.Error
}

// DeregisterDevice removes the tokens registered for an identity device (its
// sign-out), for userID when given.
func (s *Service) DeregisterDevice(ctx context.Context, userID, deviceID string) (int64, error) {
	if deviceID == "" {
		return 0, nil
	}
	q := s.DB.WithContext(ctx).Where(`"device_id" = ?`, deviceID)
	if userID != "" {
		q = q.Where(`"user_id" = ?`, userID)
	}
	res := q.Delete(&Token{})
	return res.RowsAffected, res.Error
}

// DeregisterUser removes every token of userID (signing out everywhere,
// closing the account).
func (s *Service) DeregisterUser(ctx context.Context, userID string) (int64, error) {
	if userID == "" {
		return 0, nil
	}
	res := s.DB.WithContext(ctx).Where(`"user_id" = ?`, userID).Delete(&Token{})
	return res.RowsAffected, res.Error
}

// For lists userID's registrations, newest first, tokens included, for the
// push sender.
func (s *Service) For(ctx context.Context, userID string) ([]Token, error) {
	var out []Token
	err := s.DB.WithContext(ctx).Where(`"user_id" = ?`, userID).Order(`"updated_at" DESC`).Find(&out).Error
	return out, err
}
