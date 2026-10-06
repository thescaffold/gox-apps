// Package apikey issues and checks machine credentials for a workspace.
//
// A key looks like  <brand>_<env>_<kind>_<random>  (for example
// lgr_test_sk_3f9c…): the environment (test or live) and kind (sk = secret,
// pk = publishable) are readable from the prefix, never from a URL. Only a
// SHA-256 of the key is stored, so a database leak does not leak keys, and a key
// is shown once, when it is issued.
package apikey

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base32"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Kinds and environments.
const (
	Secret      = "secret"
	Publishable = "publishable"
	Test        = "test"
	Live        = "live"
)

// Errors a caller can match.
var (
	ErrInvalid       = errors.New("invalid api key")
	ErrRevoked       = errors.New("api key revoked")
	ErrExpired       = errors.New("api key expired")
	ErrNotFound      = errors.New("api key not found")
	ErrLiveLocked    = errors.New("live keys are not available until this workspace has been approved for live")
	ErrScopeRequired = errors.New("a client and workspace are required")
)

// Claims are what a valid key authenticates as.
type Claims struct {
	KeyID       string   `json:"key_id"`
	UserID      string   `json:"user_id"`
	ClientID    string   `json:"client_id"`
	WorkspaceID string   `json:"workspace_id"`
	Kind        string   `json:"kind"`
	Environment string   `json:"environment"`
	Scopes      []string `json:"scopes"`
}

// Key is a stored key as listed (never includes the secret value).
type Key struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Kind        string     `json:"kind"`
	Environment string     `json:"environment"`
	Prefix      string     `json:"prefix"` // the first characters, safe to display
	Scopes      []string   `json:"scopes"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	RevokedAt   *time.Time `json:"revoked_at,omitempty"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	// Secret is set only on the Issue / Rotate result: the one time the key is shown.
	Secret string `json:"secret,omitempty"`
}

// Service issues and verifies keys.
type Service struct {
	DB *gorm.DB
	// Brand prefixes keys ("lgr"); default "key".
	Brand string
	// LiveAllowed says whether a workspace may hold live keys (the go-live
	// review); nil allows them.
	LiveAllowed func(ctx context.Context, clientID, workspaceID string) bool
	// TouchEvery bounds how often last_used_at is written per key (default 1m).
	TouchEvery time.Duration
	Now        func() time.Time
}

func (s *Service) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return time.Now()
}

func (s *Service) brand() string {
	if s.Brand != "" {
		return s.Brand
	}
	return "key"
}

var b32 = base32.NewEncoding("abcdefghijklmnopqrstuvwxyz234567").WithPadding(base32.NoPadding)

func hashOf(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}

func kindCode(kind string) string {
	if kind == Secret {
		return "sk"
	}
	return "pk"
}

// Issue creates a key and returns it with its one-time Secret.
func (s *Service) Issue(ctx context.Context, userID, clientID, workspaceID, name, kind, environment string, scopes []string, ttl time.Duration) (Key, error) {
	if clientID == "" || workspaceID == "" {
		return Key{}, ErrScopeRequired
	}
	if kind != Secret && kind != Publishable {
		return Key{}, fmt.Errorf("%w: kind must be secret or publishable", ErrInvalid)
	}
	if environment != Test && environment != Live {
		return Key{}, fmt.Errorf("%w: environment must be test or live", ErrInvalid)
	}
	if strings.TrimSpace(name) == "" {
		return Key{}, fmt.Errorf("%w: a key needs a name", ErrInvalid)
	}
	if environment == Live && s.LiveAllowed != nil && !s.LiveAllowed(ctx, clientID, workspaceID) {
		return Key{}, ErrLiveLocked
	}
	if len(scopes) == 0 {
		scopes = []string{"*"}
	}
	raw := make([]byte, 20)
	if _, err := rand.Read(raw); err != nil {
		return Key{}, err
	}
	secret := fmt.Sprintf("%s_%s_%s_%s", s.brand(), environment, kindCode(kind), b32.EncodeToString(raw))
	prefix := secret[:len(s.brand())+len(environment)+len(kindCode(kind))+3+4] // brand_env_kind_ + 4 random chars
	var exp *time.Time
	if ttl > 0 {
		t := s.now().Add(ttl)
		exp = &t
	}
	sc, _ := json.Marshal(scopes)
	k := Key{ID: uuid.NewString(), Name: name, Kind: kind, Environment: environment, Prefix: prefix, Scopes: scopes, CreatedAt: s.now(), ExpiresAt: exp, Secret: secret}
	if err := s.DB.WithContext(ctx).Exec(`INSERT INTO "IdentityApiKeys" (id, user_id, client_id, workspace_id, name, kind, environment, prefix, hash, scopes, expires_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?::jsonb, ?)`,
		k.ID, userID, clientID, workspaceID, name, kind, environment, prefix, hashOf(secret), string(sc), exp).Error; err != nil {
		return Key{}, err
	}
	return k, nil
}

var keyShape = regexp.MustCompile(`^[a-z0-9]{2,12}_(test|live)_(sk|pk)_[a-z2-7]{20,}$`)

// Authenticate checks a raw key and returns what it stands for. The lookup is
// by hash; revoked and expired keys are refused with distinct errors.
func (s *Service) Authenticate(ctx context.Context, raw string) (Claims, error) {
	if !keyShape.MatchString(raw) || !strings.HasPrefix(raw, s.brand()+"_") {
		return Claims{}, ErrInvalid
	}
	var r struct {
		ID, UserID, ClientID, WorkspaceID, Kind, Environment, Hash string
		Scopes                                                     []byte
		ExpiresAt, RevokedAt, LastUsedAt                           *time.Time
	}
	res := s.DB.WithContext(ctx).Raw(`SELECT id, user_id, client_id, workspace_id, kind, environment, hash, scopes, expires_at, revoked_at, last_used_at
		FROM "IdentityApiKeys" WHERE hash = ?`, hashOf(raw)).Scan(&r)
	if res.Error != nil {
		return Claims{}, res.Error
	}
	if res.RowsAffected == 0 || subtle.ConstantTimeCompare([]byte(r.Hash), []byte(hashOf(raw))) != 1 {
		return Claims{}, ErrInvalid
	}
	if r.RevokedAt != nil {
		return Claims{}, ErrRevoked
	}
	if r.ExpiresAt != nil && !s.now().Before(*r.ExpiresAt) {
		return Claims{}, ErrExpired
	}
	// a key issued for live must still be allowed live (a workspace can be suspended)
	if r.Environment == Live && s.LiveAllowed != nil && !s.LiveAllowed(ctx, r.ClientID, r.WorkspaceID) {
		return Claims{}, ErrLiveLocked
	}
	every := s.TouchEvery
	if every == 0 {
		every = time.Minute
	}
	if r.LastUsedAt == nil || s.now().Sub(*r.LastUsedAt) > every {
		_ = s.DB.WithContext(ctx).Exec(`UPDATE "IdentityApiKeys" SET last_used_at = ? WHERE id = ?`, s.now(), r.ID).Error
	}
	var scopes []string
	_ = json.Unmarshal(r.Scopes, &scopes)
	return Claims{KeyID: r.ID, UserID: r.UserID, ClientID: r.ClientID, WorkspaceID: r.WorkspaceID, Kind: r.Kind, Environment: r.Environment, Scopes: scopes}, nil
}

// Allows reports whether the claims' scopes permit a scope ("*", "wallets:*" or exact).
func (c Claims) Allows(scope string) bool {
	for _, have := range c.Scopes {
		if have == "*" || have == scope {
			return true
		}
		if strings.HasSuffix(have, ":*") && strings.HasPrefix(scope, strings.TrimSuffix(have, "*")) {
			return true
		}
	}
	return false
}

// List returns a workspace's keys (never their values).
func (s *Service) List(ctx context.Context, clientID, workspaceID string) ([]Key, error) {
	if clientID == "" {
		return nil, ErrScopeRequired
	}
	q := `SELECT id, name, kind, environment, prefix, scopes, created_at, expires_at, revoked_at, last_used_at FROM "IdentityApiKeys" WHERE client_id = ?`
	args := []any{clientID}
	if workspaceID != "" {
		q += " AND workspace_id = ?"
		args = append(args, workspaceID)
	}
	q += " ORDER BY created_at DESC"
	var rows []struct {
		ID, Name, Kind, Environment, Prefix string
		Scopes                              []byte
		CreatedAt                           time.Time
		ExpiresAt, RevokedAt, LastUsedAt    *time.Time
	}
	if err := s.DB.WithContext(ctx).Raw(q, args...).Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]Key, len(rows))
	for i, r := range rows {
		var sc []string
		_ = json.Unmarshal(r.Scopes, &sc)
		out[i] = Key{ID: r.ID, Name: r.Name, Kind: r.Kind, Environment: r.Environment, Prefix: r.Prefix, Scopes: sc, CreatedAt: r.CreatedAt, ExpiresAt: r.ExpiresAt, RevokedAt: r.RevokedAt, LastUsedAt: r.LastUsedAt}
	}
	return out, nil
}

// Revoke disables a key; another client's id is simply not found.
func (s *Service) Revoke(ctx context.Context, clientID, id string) error {
	if clientID == "" {
		return ErrScopeRequired
	}
	res := s.DB.WithContext(ctx).Exec(`UPDATE "IdentityApiKeys" SET revoked_at = ? WHERE id = ? AND client_id = ? AND revoked_at IS NULL`, s.now(), id, clientID)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// Rotate issues a replacement with the same name, kind, environment and scopes
// and revokes the old key after grace (0 revokes it at once), so a deployment
// can switch without downtime.
func (s *Service) Rotate(ctx context.Context, userID, clientID, id string, grace time.Duration) (Key, error) {
	if clientID == "" {
		return Key{}, ErrScopeRequired
	}
	var r struct {
		WorkspaceID, Name, Kind, Environment string
		Scopes                               []byte
		RevokedAt                            *time.Time
	}
	res := s.DB.WithContext(ctx).Raw(`SELECT workspace_id, name, kind, environment, scopes, revoked_at FROM "IdentityApiKeys" WHERE id = ? AND client_id = ?`, id, clientID).Scan(&r)
	if res.Error != nil {
		return Key{}, res.Error
	}
	if res.RowsAffected == 0 || r.RevokedAt != nil {
		return Key{}, ErrNotFound
	}
	var sc []string
	_ = json.Unmarshal(r.Scopes, &sc)
	nk, err := s.Issue(ctx, userID, clientID, r.WorkspaceID, r.Name, r.Kind, r.Environment, sc, 0)
	if err != nil {
		return Key{}, err
	}
	if grace <= 0 {
		return nk, s.Revoke(ctx, clientID, id)
	}
	// the old key keeps working until the grace period ends
	if err := s.DB.WithContext(ctx).Exec(`UPDATE "IdentityApiKeys" SET expires_at = ? WHERE id = ? AND client_id = ?`, s.now().Add(grace), id, clientID).Error; err != nil {
		return Key{}, err
	}
	return nk, nil
}
