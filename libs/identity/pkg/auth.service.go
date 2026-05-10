package pkg

import (
	"encoding/base64"
	"errors"
	"os"
	"strings"
	"time"

	identityattribute "github.com/thescaffold/gox-apps/libs/identity/app/attribute"
	identitypermission "github.com/thescaffold/gox-apps/libs/identity/app/permission"
	identityrole "github.com/thescaffold/gox-apps/libs/identity/app/role"
	"github.com/thescaffold/gox-apps/libs/identity/app/token"
	"github.com/thescaffold/gox-apps/libs/identity/app/user"
	"github.com/thescaffold/gox-packages/libs/core/auth"
	"github.com/thescaffold/gox-packages/libs/core/security"
	"github.com/thescaffold/gox-packages/libs/core/utils"
)

const (
	tokenTypeRefresh = "refresh"
	tokenTypeAPI     = "api"
	accessExpiry     = 15 * time.Minute
	refreshExpiry    = 30 * 24 * time.Hour
)

type AuthService struct {
	userEntity       *user.UserEntity                     `inject:""`
	tokenEntity      *token.TokenEntity                   `inject:""`
	roleEntity       *identityrole.RoleEntity             `inject:""`
	permissionEntity *identitypermission.PermissionEntity `inject:""`
	attributeEntity  *identityattribute.AttributeEntity   `inject:""`
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

// Login validates credentials and returns signed access + refresh tokens.
func (s *AuthService) Login(email, password, clientId string) (accessToken, refreshToken string, err error) {
	u, err := s.userEntity.First(`"email" = ?`, email)
	if err != nil || u == nil {
		return "", "", errors.New("invalid credentials")
	}
	if u.Secret == nil || !security.Compare(password, *u.Secret) {
		return "", "", errors.New("invalid credentials")
	}

	claims := s.buildClaims(u, clientId, "")
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
		Token:     refreshToken,
		ExpiresAt: &expiresAt,
	})

	return accessToken, refreshToken, nil
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
	email, password := parts[0], parts[1]

	u, err := s.userEntity.First(`"email" = ?`, email)
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

	rec, err := s.tokenEntity.First(`"token" = ? AND "type" = ?`, refreshToken, tokenTypeRefresh)
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

// RegisterUser creates a new user with a bcrypt-hashed password.
func (s *AuthService) RegisterUser(name, email, password string) (*user.User, error) {
	hashed, err := security.Hash(password)
	if err != nil {
		return nil, err
	}
	u := &user.User{
		Name:   name,
		Email:  email,
		Secret: &hashed,
	}
	u.Id = utils.UUID()
	if err := s.userEntity.Insert(u); err != nil {
		return nil, err
	}
	return u, nil
}

// --- Role / Permission resolution ---

// ResolveRoles returns all role records for a client (optionally scoped to workspace).
func (s *AuthService) ResolveRoles(_, clientId, workspaceId string) ([]identityrole.Role, error) {
	query := `"client_id" = ?`
	args := []any{clientId}
	if workspaceId != "" {
		query += ` AND ("workspace_id" = ? OR "workspace_id" IS NULL)`
		args = append(args, workspaceId)
	}
	return s.roleEntity.Find(0, 0, query, args...)
}

// ResolvePermissions returns all permission records for a user in a client workspace.
func (s *AuthService) ResolvePermissions(userId, clientId, workspaceId string) ([]identitypermission.Permission, error) {
	query := `"user_id" = ? AND "client_id" = ?`
	args := []any{userId, clientId}
	if workspaceId != "" {
		query += ` AND ("workspace_id" = ? OR "workspace_id" IS NULL)`
		args = append(args, workspaceId)
	}
	return s.permissionEntity.Find(0, 0, query, args...)
}

// ResolvePreference returns the user's preference attributes for a client workspace.
func (s *AuthService) ResolvePreference(userId, clientId, workspaceId string) (map[string]any, error) {
	query := `"user_id" = ? AND "client_id" = ?`
	args := []any{userId, clientId}
	if workspaceId != "" {
		query += ` AND ("workspace_id" = ? OR "workspace_id" IS NULL)`
		args = append(args, workspaceId)
	}
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
