package pkg

import (
	"github.com/thescaffold/gox-apps/libs/identity/pkg/tokenstore"
	"encoding/base64"
	"errors"
	"fmt"
	mrand "math/rand"
	"os"
	"strings"
	"time"

	identityattribute "github.com/thescaffold/gox-apps/libs/identity/app/attribute"
	identitypermission "github.com/thescaffold/gox-apps/libs/identity/app/permission"
	identityrole "github.com/thescaffold/gox-apps/libs/identity/app/role"
	"github.com/thescaffold/gox-apps/libs/identity/app/token"
	"github.com/thescaffold/gox-apps/libs/identity/app/user"
	"github.com/thescaffold/gox-packages/libs/core/auth"
	"github.com/thescaffold/gox-packages/libs/core/events"
	"github.com/thescaffold/gox-packages/libs/core/security"
	"github.com/thescaffold/gox-packages/libs/core/services"
	"github.com/thescaffold/gox-packages/libs/core/utils"
)

const (
	tokenTypeRefresh = "refresh"
	tokenTypeAPI     = "api"
	accessExpiry     = 15 * time.Minute
	refreshExpiry    = 30 * 24 * time.Hour
	otpExpiry        = 10 * time.Minute
	otpCachePrefix   = "identity:otp:"
)

type AuthService struct {
	userEntity       *user.UserEntity                     `inject:""`
	tokenEntity      *token.TokenEntity                   `inject:""`
	roleEntity       *identityrole.RoleEntity             `inject:""`
	permissionEntity *identitypermission.PermissionEntity `inject:""`
	attributeEntity  *identityattribute.AttributeEntity   `inject:""`
	cache            services.CacheBackend                `inject:""`
	tracker          *events.TrackerService               `inject:""`
}

// GenerateOTP produces a 6-digit numeric one-time code and stores it under
// `identity:otp:<handle>` in CacheBackend for otpExpiry. Returns the code so
// caller (controller) can wrap it in an event for the email/SMS dispatch.
// nil-cache tolerance: returns the code but doesn't persist it — handy for
// tests that want determinism without setting up the cache.
func (s *AuthService) GenerateOTP(handle string) string {
	code := fmt.Sprintf("%06d", mrand.Intn(1_000_000))
	if s.cache != nil && handle != "" {
		s.cache.Set(otpCachePrefix+handle, code, otpExpiry)
	}
	if s.tracker != nil && handle != "" {
		// Mirror TS trackerService.message dispatch on initiate — the
		// notification handler reads `key:'auth-initiate'` and routes the
		// code over the user's preferred channel (email/SMS).
		s.tracker.Message("apps.notification.message.new", map[string]any{
			"reference": handle,
			"key":       "auth-initiate",
			"channels":  []string{"email"},
			"data":      map[string]any{"code": code, "handle": handle},
			"subject":   "Your one-time verification code",
			"type":      "system",
			"priority":  "high",
			"scope":     "user",
		})
	}
	return code
}

// IssueTokensByHandle issues an access+refresh pair for the user identified
// by `handle` (ref OR email). Used by the OTP-verify path which has already
// validated the user out-of-band — no password check. workspaceId is the
// workspace the caller has already resolved for this session (U-S9): pass
// "" only when the user genuinely has no workspace yet (never as a
// convenience default — that's exactly the F6 bug this parameter exists to
// close, see buildClaims's own doc comment).
func (s *AuthService) IssueTokensByHandle(handle, clientId, workspaceId string) (string, string, error) {
	u, err := s.userEntity.First(`"ref" = ? OR "email" = ?`, handle, handle)
	if err != nil || u == nil {
		return "", "", errors.New("invalid credentials")
	}
	claims := s.buildClaims(u, clientId, workspaceId)
	access, err := auth.Sign(claims, s.jwtSecret(), accessExpiry)
	if err != nil {
		return "", "", err
	}
	refresh, err := auth.Sign(claims, s.jwtSecret(), refreshExpiry)
	if err != nil {
		return "", "", err
	}
	expiresAt := time.Now().Add(refreshExpiry)
	_ = s.tokenEntity.Insert(&token.Token{
		UserId: u.Id, ClientId: clientId, Type: tokenTypeRefresh,
		Token: tokenstore.HashRefresh(refresh), ExpiresAt: &expiresAt,
	})
	return access, refresh, nil
}

// VerifyOTP returns true when the stored code for `handle` matches `code`.
// On a successful verify the stored code is removed (single-use semantics).
func (s *AuthService) VerifyOTP(handle, code string) bool {
	if s.cache == nil || handle == "" || code == "" {
		return false
	}
	stored, ok := s.cache.Get(otpCachePrefix + handle)
	if !ok || stored != code {
		return false
	}
	s.cache.Del(otpCachePrefix + handle)
	return true
}

func (s *AuthService) jwtSecret() string {
	if v := os.Getenv("JWT_SECRET"); v != "" {
		return v
	}
	return "changeme"
}

func (s *AuthService) hmacKey() string {
	if v := os.Getenv("HMAC_KEY"); v != "" {
		return v
	}
	return "changeme-hmac"
}

// --- Auth methods ---

// VerifyCredentials checks handle+password (accepts either the new `ref`
// column, TS-aligned Phase H, or the legacy `email` — lookup is OR'd so
// existing accounts keep working) and returns the user on success. Split out
// of Login (PLAN M1-02) so a caller that needs to check something about the
// account — closed/suspended status, say — can do so *between* verifying
// the password and issuing tokens, without revealing that status to a
// request that never had a valid password in the first place (identity/app/
// auth's Login handler is exactly this: checking suspended/closed before
// VerifyCredentials would let an attacker learn an account's status with no
// valid credentials at all).
func (s *AuthService) VerifyCredentials(handle, password string) (*user.User, error) {
	u, err := s.userEntity.First(`"ref" = ? OR "email" = ?`, handle, handle)
	if err != nil || u == nil {
		return nil, errors.New("invalid credentials")
	}
	if u.Secret == nil || !security.Compare(password, *u.Secret) {
		return nil, errors.New("invalid credentials")
	}
	return u, nil
}

// IssueTokens signs and persists an access+refresh pair for an
// already-verified user. workspaceId (U-S9, PLAN M1-02) is the workspace
// this session's claims are scoped to — the caller must resolve it before
// calling (the user's own default membership, or one the caller
// re-verified belongs to them; see identity/app/auth's Login handler and
// SwitchWorkspace). Passing "" is only correct when the user genuinely has
// no workspace yet, never as a shortcut: an empty workspaceId here used to
// mean "every workspace's permissions, unfiltered" (TRD F6) —
// ResolveRoles/ResolvePermissions now treat it as "global-scope rows only"
// instead, but a caller that always passes "" out of convenience would
// still hand out a token good for nothing tenant-scoped, silently.
func (s *AuthService) IssueTokens(u *user.User, clientId, workspaceId string) (accessToken, refreshToken string, err error) {
	claims := s.buildClaims(u, clientId, workspaceId)
	accessToken, err = auth.Sign(claims, s.jwtSecret(), accessExpiry)
	if err != nil {
		return "", "", err
	}
	refreshToken, err = auth.Sign(claims, s.jwtSecret(), refreshExpiry)
	if err != nil {
		return "", "", err
	}

	expiresAt := time.Now().Add(refreshExpiry)
	_ = s.tokenEntity.Insert(&token.Token{
		UserId:    u.Id,
		ClientId:  clientId,
		Type:      tokenTypeRefresh,
		Token:     tokenstore.HashRefresh(refreshToken),
		ExpiresAt: &expiresAt,
	})

	return accessToken, refreshToken, nil
}

// Login is VerifyCredentials + IssueTokens combined, for a caller that has
// no reason to interleave anything between the two (see VerifyCredentials's
// own doc comment for why identity/app/auth's real Login handler doesn't
// use this).
func (s *AuthService) Login(handle, password, clientId, workspaceId string) (accessToken, refreshToken string, err error) {
	u, err := s.VerifyCredentials(handle, password)
	if err != nil {
		return "", "", err
	}
	return s.IssueTokens(u, clientId, workspaceId)
}

// BearerAuth verifies a JWT Bearer token and returns the decoded claims.
func (s *AuthService) BearerAuth(tokenStr string) (map[string]any, error) {
	claims, err := auth.Verify(tokenStr, s.jwtSecret())
	if err != nil {
		return nil, errors.New("invalid or expired token")
	}
	return claims, nil
}

// BasicAuth authenticates via base64-encoded "user:password" credentials.
func (s *AuthService) BasicAuth(encoded, clientId string) (map[string]any, error) {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, errors.New("invalid basic auth encoding")
	}
	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		return nil, errors.New("malformed basic auth: expected user:password")
	}
	handle, password := parts[0], parts[1]

	u, err := s.userEntity.First(`"ref" = ? OR "email" = ?`, handle, handle)
	if err != nil || u == nil {
		return nil, errors.New("invalid credentials")
	}
	if u.Secret == nil || !security.Compare(password, *u.Secret) {
		return nil, errors.New("invalid credentials")
	}

	return s.buildClaims(u, clientId, ""), nil
}

// TokenAuth authenticates via a stored API token value.
func (s *AuthService) TokenAuth(tokenValue, clientId string) (map[string]any, error) {
	rec, err := s.tokenEntity.First(`"token" = ? AND "type" = ?`, tokenValue, tokenTypeAPI)
	if err != nil || rec == nil {
		return nil, errors.New("token not found")
	}
	if rec.ExpiresAt != nil && rec.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("token expired")
	}

	u, err := s.userEntity.First(`"id" = ?`, rec.UserId)
	if err != nil || u == nil {
		return nil, errors.New("user not found")
	}

	return s.buildClaims(u, clientId, ""), nil
}

// HMACAuth verifies an HMAC-signed request signature.
// payload is the map that was signed (typically includes nonce + timestamp).
func (s *AuthService) HMACAuth(signature string, payload map[string]any) (bool, error) {
	expected, err := security.GenerateHmac(payload, "sha256", s.hmacKey())
	if err != nil {
		return false, err
	}
	if signature != expected {
		return false, errors.New("invalid hmac signature")
	}
	return true, nil
}

// RefreshToken verifies a refresh token record and issues a new access token.
func (s *AuthService) RefreshToken(refreshToken string) (string, error) {
	claims, err := auth.Verify(refreshToken, s.jwtSecret())
	if err != nil {
		return "", errors.New("invalid or expired token")
	}

	rec, err := s.tokenEntity.First(`"token" = ? AND "type" = ?`, tokenstore.HashRefresh(refreshToken), tokenTypeRefresh)
	if err != nil || rec == nil {
		return "", errors.New("token revoked")
	}
	if rec.ExpiresAt != nil && rec.ExpiresAt.Before(time.Now()) {
		return "", errors.New("token expired")
	}

	newAccess, err := auth.Sign(claims, s.jwtSecret(), accessExpiry)
	if err != nil {
		return "", err
	}
	return newAccess, nil
}

// Logout soft-deletes the token record by ID.
func (s *AuthService) Logout(tokenId string) error {
	_, err := s.tokenEntity.Delete(`"id" = ?`, tokenId)
	return err
}

// ValidateToken verifies a JWT access token and returns its claims.
func (s *AuthService) ValidateToken(tokenStr string) (map[string]any, error) {
	return auth.Verify(tokenStr, s.jwtSecret())
}

// RegisterUser creates a new user with a bcrypt-hashed password. Mirrors TS:
// `ref` is set to the supplied email (the canonical handle going forward),
// `type` is set to "email" so identity-level dispatchers know which channel
// to route OTPs through. Both legacy email + new ref columns are populated.
func (s *AuthService) RegisterUser(name, email, password string) (*user.User, error) {
	hashed, err := security.Hash(password)
	if err != nil {
		return nil, err
	}
	ref := email
	refType := "email"
	u := &user.User{
		Name:   name,
		Email:  email,
		Ref:    &ref,
		Type:   &refType,
		Secret: &hashed,
	}
	u.Id = utils.UUID()
	if err := s.userEntity.Insert(u); err != nil {
		return nil, err
	}
	return u, nil
}

// --- Role / Permission resolution ---

// workspaceScope builds the workspace half of a Resolve* WHERE clause. A
// real workspaceId scopes to that workspace's own rows plus global rows
// (workspace_id IS NULL); an empty one (U-S9, PLAN M1-02 / TRD F6) used to
// mean "no filter at all" — every row for every workspace this user or
// client ever touched, the exact cross-tenant leak F6 found. It now means
// "global rows only": the safe, minimal result for a session that genuinely
// has no workspace context yet, never a silent "everything."
func workspaceScope(workspaceId string) (clause string, args []any) {
	if workspaceId != "" {
		return ` AND ("workspace_id" = ? OR "workspace_id" IS NULL)`, []any{workspaceId}
	}
	return ` AND "workspace_id" IS NULL`, nil
}

// ResolveRoles returns all role records for a client (optionally scoped to workspace).
func (s *AuthService) ResolveRoles(_, clientId, workspaceId string) ([]identityrole.Role, error) {
	query := `"client_id" = ?`
	args := []any{clientId}
	clause, extra := workspaceScope(workspaceId)
	query += clause
	args = append(args, extra...)
	return s.roleEntity.Find(0, 0, query, args...)
}

// ResolvePermissions returns all permission records for a user in a client workspace.
func (s *AuthService) ResolvePermissions(userId, clientId, workspaceId string) ([]identitypermission.Permission, error) {
	query := `"user_id" = ? AND "client_id" = ?`
	args := []any{userId, clientId}
	clause, extra := workspaceScope(workspaceId)
	query += clause
	args = append(args, extra...)
	return s.permissionEntity.Find(0, 0, query, args...)
}

// ResolvePreference returns the user's preference attributes for a client workspace.
func (s *AuthService) ResolvePreference(userId, clientId, workspaceId string) (map[string]any, error) {
	query := `"user_id" = ? AND "client_id" = ?`
	args := []any{userId, clientId}
	clause, extra := workspaceScope(workspaceId)
	query += clause
	args = append(args, extra...)
	attrs, err := s.attributeEntity.Find(0, 0, query, args...)
	if err != nil {
		return nil, err
	}

	pref := map[string]any{}
	for _, a := range attrs {
		if a.Value != nil {
			pref[a.Key] = *a.Value
		}
	}
	return pref, nil
}

// --- helpers ---

func (s *AuthService) buildClaims(u *user.User, clientId, workspaceId string) map[string]any {
	claims := map[string]any{
		"sub":      u.Id,
		"email":    u.Email,
		"clientId": clientId,
	}
	if workspaceId != "" {
		claims["workspaceId"] = workspaceId
	}

	roles, _ := s.ResolveRoles(u.Id, clientId, workspaceId)
	roleNames := make([]string, 0, len(roles))
	for i := range roles {
		roleNames = append(roleNames, roles[i].Name)
	}
	claims["roles"] = roleNames

	permissions, _ := s.ResolvePermissions(u.Id, clientId, workspaceId)
	permKeys := make([]string, 0, len(permissions))
	for i := range permissions {
		permKeys = append(permKeys, permissions[i].Key)
	}
	claims["permissions"] = permKeys

	return claims
}
