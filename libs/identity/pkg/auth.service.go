package pkg

import (
	"errors"
	"os"
	"time"

	"github.com/thescaffold/gox-apps-identity/app/token"
	"github.com/thescaffold/gox-apps-identity/app/user"
	"github.com/thescaffold/gox-packages-core/auth"
	"github.com/thescaffold/gox-packages-core/security"
	"github.com/thescaffold/gox-packages-core/utils"
)

const (
	tokenTypeRefresh = "refresh"
	accessExpiry     = 15 * time.Minute
	refreshExpiry    = 30 * 24 * time.Hour
)

type AuthService struct {
	userEntity  *user.UserEntity   `inject:""`
	tokenEntity *token.TokenEntity `inject:""`
}

func (s *AuthService) jwtSecret() string {
	if v := os.Getenv("JWT_SECRET"); v != "" {
		return v
	}
	return "changeme"
}

// Login validates credentials and returns signed access + refresh tokens.
func (s *AuthService) Login(email, password, clientId string) (accessToken, refreshToken string, err error) {
	u, err := s.userEntity.First("email = ?", email)
	if err != nil || u == nil {
		return "", "", errors.New("invalid credentials")
	}
	if u.Secret == nil || !security.Compare(password, *u.Secret) {
		return "", "", errors.New("invalid credentials")
	}

	claims := map[string]any{
		"sub":      u.Id,
		"email":    u.Email,
		"clientId": clientId,
	}

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
	_, err := s.tokenEntity.Delete("id = ?", tokenId)
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
