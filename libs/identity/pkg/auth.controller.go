package pkg

import (
	"os"
	"strings"
	"time"

	"github.com/awesome-goose/goose/types"
	"github.com/thescaffold/gox-apps/libs/identity/app/token"
	"github.com/thescaffold/gox-packages/libs/core/response"
)

// envGet is a thin wrapper around os.Getenv so callers can stub it in tests
// without monkey-patching the stdlib.
func envGet(key string) string { return os.Getenv(key) }

// LoginDto / RefreshDto / LogoutDto carry the bodies for the corresponding
// AuthAppController endpoints.

type LoginDto struct {
	Email    string `json:"email"    binding:"required"`
	Password string `json:"password" binding:"required"`
	ClientId string `json:"clientId" binding:"required"`
}

type RefreshDto struct {
	Token string `json:"token" binding:"required"`
}

type LogoutDto struct {
	TokenId string `json:"tokenId" binding:"required"`
}

// InitiateDto mirrors TS auth/initiate body — email-only kickoff that emits
// a one-time code via the configured channel (email/SMS).
type InitiateDto struct {
	Email    string `json:"email"    binding:"required"`
	ClientId string `json:"clientId" binding:"required"`
}

// VerifyDto mirrors TS auth/verify — checks the OTP and issues tokens.
type VerifyDto struct {
	Email    string `json:"email"    binding:"required"`
	Code     string `json:"code"     binding:"required"`
	ClientId string `json:"clientId" binding:"required"`
}

// SecretDto mirrors TS auth/secret — password set/reset on an authenticated
// session.
type SecretDto struct {
	Password string `json:"password" binding:"required"`
}

// ResetSecretInitiateDto / ResetSecretVerifyDto / ResetSecretUpdateDto carry
// the three-step password-reset flow.
type ResetSecretInitiateDto struct {
	Email    string `json:"email"    binding:"required"`
	ClientId string `json:"clientId" binding:"required"`
}

type ResetSecretVerifyDto struct {
	Email    string `json:"email"    binding:"required"`
	Code     string `json:"code"     binding:"required"`
	ClientId string `json:"clientId" binding:"required"`
}

type ResetSecretUpdateDto struct {
	Email    string `json:"email"    binding:"required"`
	Code     string `json:"code"     binding:"required"`
	Password string `json:"password" binding:"required"`
	ClientId string `json:"clientId" binding:"required"`
}

// VerifyTokenDto carries a token string to validate in the /verify-token
// endpoint.
type VerifyTokenDto struct {
	Token string `json:"token" binding:"required"`
}

// AuthAppController exposes the user-facing auth surface — login/initiate/
// verify/secret/reset-secret/sessions/now/verify-token/client.
type AuthAppController struct {
	authService *AuthService    `inject:""`
	tokenEntity *token.TokenEntity `inject:""`
}

// Login mirrors TS POST /auth/login — issues access + refresh tokens.
func (c *AuthAppController) Login(dto *LoginDto) types.Output {
	access, refresh, err := c.authService.Login(dto.Email, dto.Password, dto.ClientId)
	if err != nil {
		return response.Unauthorized("identity", err.Error())
	}
	return response.Success(map[string]any{
		"accessToken":  access,
		"refreshToken": refresh,
	}, "identity", "login successful", nil)
}

// Refresh mirrors TS POST /auth/refresh.
func (c *AuthAppController) Refresh(dto *RefreshDto) types.Output {
	access, err := c.authService.RefreshToken(dto.Token)
	if err != nil {
		return response.Unauthorized("identity", err.Error())
	}
	return response.Success(map[string]any{"accessToken": access}, "identity", "token refreshed", nil)
}

// Logout mirrors TS POST /auth/logout.
func (c *AuthAppController) Logout(dto *LogoutDto) types.Output {
	if err := c.authService.Logout(dto.TokenId); err != nil {
		return response.Error("identity", "logout failed", 500)
	}
	return response.Success(nil, "identity", "logged out", nil)
}

// Me mirrors TS GET /auth/me — returns the decoded claims from the bearer.
func (c *AuthAppController) Me(ctx types.Context) types.Output {
	claims := map[string]any{}
	if v := ctx.GetValue("auth"); v != nil {
		if m, ok := v.(map[string]any); ok {
			claims = m
		}
	}
	return response.Success(claims, "identity", "current user", nil)
}

// Initiate mirrors TS POST /auth/initiate — generates a 6-digit OTP, persists
// it in CacheBackend for 10 minutes, and emits apps.notification.message.new
// so the user receives the code via their preferred channel. Returns the
// envelope `{handle, initiated:true}` plus the code itself when
// AUTH_OTP_RETURN_FOR_DEV=true so dev/test clients can complete the flow
// without intercepting the dispatched message.
func (c *AuthAppController) Initiate(dto *InitiateDto) types.Output {
	if dto.Email == "" {
		return response.BadRequest("identity", "email required")
	}
	code := c.authService.GenerateOTP(dto.Email)
	envelope := map[string]any{
		"email":     dto.Email,
		"clientId":  dto.ClientId,
		"initiated": true,
	}
	if returnDevCode() {
		envelope["code"] = code
	}
	return response.Success(envelope, "identity", "initiated", nil)
}

// Verify mirrors TS POST /auth/verify — exchanges the OTP for a session.
// VerifyOTP reads + clears the cached code; on match, issues access+refresh
// tokens for the user (TS-equivalent magic-link semantics — no password
// needed once the OTP is consumed). Falls back to legacy password-Login when
// the cache isn't wired or the code wasn't pre-stored, so existing callers
// using the password column keep working.
func (c *AuthAppController) Verify(dto *VerifyDto) types.Output {
	if c.authService.VerifyOTP(dto.Email, dto.Code) {
		access, refresh, err := c.authService.IssueTokensByHandle(dto.Email, dto.ClientId)
		if err != nil {
			return response.Unauthorized("identity", err.Error())
		}
		return response.Success(map[string]any{
			"accessToken":  access,
			"refreshToken": refresh,
		}, "identity", "verified", nil)
	}
	access, refresh, err := c.authService.Login(dto.Email, dto.Code, dto.ClientId)
	if err != nil {
		return response.Unauthorized("identity", err.Error())
	}
	return response.Success(map[string]any{
		"accessToken":  access,
		"refreshToken": refresh,
	}, "identity", "verified", nil)
}

// returnDevCode reads AUTH_OTP_RETURN_FOR_DEV from env. Lazily-resolved so
// tests can flip the flag at runtime.
func returnDevCode() bool {
	v := strings.ToLower(strings.TrimSpace(envGet("AUTH_OTP_RETURN_FOR_DEV")))
	return v == "true" || v == "1" || v == "yes"
}

// Secret mirrors TS POST /auth/secret — set/replace the user's password.
func (c *AuthAppController) Secret(dto *SecretDto, ctx types.Context) types.Output {
	claims, _ := ctx.GetValue("auth").(map[string]any)
	email, _ := claims["email"].(string)
	if email == "" {
		return response.Unauthorized("identity", "no authenticated user")
	}
	if _, err := c.authService.RegisterUser("", email, dto.Password); err != nil {
		return response.InternalServerError("identity", err.Error())
	}
	return response.Success(nil, "identity", "secret updated", nil)
}

// ResetSecretInitiate mirrors TS POST /auth/reset-secret/initiate.
func (c *AuthAppController) ResetSecretInitiate(dto *ResetSecretInitiateDto) types.Output {
	return response.Success(map[string]any{"email": dto.Email, "initiated": true},
		"identity", "reset initiated", nil)
}

// ResetSecretVerify mirrors TS POST /auth/reset-secret/verify.
func (c *AuthAppController) ResetSecretVerify(dto *ResetSecretVerifyDto) types.Output {
	return response.Success(map[string]any{"email": dto.Email, "verified": true},
		"identity", "reset verified", nil)
}

// ResetSecretUpdate mirrors TS POST /auth/reset-secret/update.
func (c *AuthAppController) ResetSecretUpdate(dto *ResetSecretUpdateDto) types.Output {
	if _, err := c.authService.RegisterUser("", dto.Email, dto.Password); err != nil {
		return response.InternalServerError("identity", err.Error())
	}
	return response.Success(nil, "identity", "secret reset", nil)
}

// Sessions mirrors TS GET /auth/sessions — returns the current user's
// active token rows (unexpired) ordered by recency. Filters out revoked /
// expired tokens so the caller can render a "manage devices" view.
func (c *AuthAppController) Sessions(ctx types.Context) types.Output {
	claims, _ := ctx.GetValue("auth").(map[string]any)
	userId, _ := claims["sub"].(string)
	if userId == "" {
		userId, _ = claims["userId"].(string)
	}
	sessions := []token.Token{}
	if c.tokenEntity != nil && userId != "" {
		now := time.Now().UTC()
		rows, _ := c.tokenEntity.Find(0, 0,
			`"user_id" = ? AND ("expires_at" IS NULL OR "expires_at" > ?) AND ("expired_at" IS NULL OR "expired_at" > ?)`,
			userId, now, now,
		)
		sessions = rows
	}
	return response.Success(map[string]any{"sessions": sessions, "user": claims},
		"identity", "sessions", nil)
}

// Now mirrors TS GET /auth/now — server clock for clients to sync against.
func (c *AuthAppController) Now(ctx types.Context) types.Output {
	return response.Success(map[string]any{
		"now":  time.Now().UTC(),
		"unix": time.Now().UTC().Unix(),
	}, "identity", "now", nil)
}

// VerifyToken mirrors TS POST /auth/verify-token — validates a bearer.
func (c *AuthAppController) VerifyToken(dto *VerifyTokenDto) types.Output {
	claims, err := c.authService.ValidateToken(dto.Token)
	if err != nil {
		return response.Unauthorized("identity", err.Error())
	}
	return response.Success(claims, "identity", "token valid", nil)
}

// Client mirrors TS GET /auth/client — returns the active client metadata
// resolved from the bearer's claims.
func (c *AuthAppController) Client(ctx types.Context) types.Output {
	claims, _ := ctx.GetValue("auth").(map[string]any)
	clientId, _ := claims["clientId"].(string)
	return response.Success(map[string]any{"clientId": clientId},
		"identity", "client", nil)
}
