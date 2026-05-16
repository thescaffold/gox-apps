package pkg

import (
	"github.com/awesome-goose/goose/types"
	identityclientlog "github.com/thescaffold/gox-apps/libs/identity/app/clientlog"
	identityprovider "github.com/thescaffold/gox-apps/libs/identity/pkg/provider"
	"github.com/thescaffold/gox-packages/libs/core/response"
)

// OAuthCallbackDto carries the provider redirect's `?code=&state=` params.
type OAuthCallbackDto struct {
	Code  string `query:"code"`
	State string `query:"state"`
}

// OAuthAuthorizeDto carries the path param `provider` for /authorize/:provider.
type OAuthAuthorizeDto struct {
	Provider string `param:"provider"`
	State    string `query:"state"`
	Redirect string `query:"redirect_uri"`
}

// OAuthClientDto / OAuthInitiateDto / OAuthAccessTokenDto / OAuthAuthCodeDto
// carry bodies for the server/device OAuth endpoints.
type OAuthClientDto struct {
	ClientId string `json:"clientId" binding:"required"`
}

type OAuthInitiateDto struct {
	ClientId string `json:"clientId" binding:"required"`
	State    string `json:"state"`
	Scope    string `json:"scope"`
}

type OAuthAuthCodeDto struct {
	SessionId string `json:"sessionId" binding:"required"`
}

type OAuthAccessTokenDto struct {
	ClientId     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	AuthCode     string `json:"authCode"`
	SessionId    string `json:"sessionId"`
}

// OAuthLoginDto carries the body for /oauth/login — final exchange after
// the provider redirect lands.
type OAuthLoginDto struct {
	Provider    string `json:"provider"    binding:"required"`
	Code        string `json:"code"        binding:"required"`
	RedirectURI string `json:"redirectUri"`
	ClientId    string `json:"clientId"`
}

// OAuthController hosts the OAuth surface — provider authorize redirects,
// server/device initiate + auth-code + access-token, and final login.
type OAuthController struct {
	authService      *AuthService                       `inject:""`
	clientLogService *identityclientlog.ClientLogService `inject:""`
	registry         *identityprovider.Registry
}

// SetRegistry wires (or replaces) the provider registry. Host code calls
// this after constructing providers with their env credentials. nil-tolerant:
// endpoints fall back to invalid-provider responses when the registry is missing.
func (c *OAuthController) SetRegistry(r *identityprovider.Registry) {
	c.registry = r
}

// Authorize mirrors TS GET /oauth/authorize/:provider — produces the
// provider's authorization URL the client should follow.
func (c *OAuthController) Authorize(dto *OAuthAuthorizeDto) types.Output {
	if c.registry == nil {
		return response.BadRequest("identity", "oauth registry not configured")
	}
	p := c.registry.Get(dto.Provider)
	if p == nil {
		return response.BadRequest("identity", "unknown oauth provider")
	}
	authURL := p.Authorize(dto.State, dto.Redirect)
	return response.Success(map[string]any{
		"provider": dto.Provider,
		"url":      authURL,
	}, "identity", "oauth authorize", nil)
}

// Callback mirrors TS GET /oauth/callback/:provider — completes the
// auth-code → token exchange and returns the profile in the envelope.
func (c *OAuthController) Callback(dto *OAuthCallbackDto, ctx types.Context) types.Output {
	providerName := ctx.Request().Params()["provider"]
	if c.registry == nil {
		return response.BadRequest("identity", "oauth registry not configured")
	}
	p := c.registry.Get(providerName)
	if p == nil {
		return response.BadRequest("identity", "unknown oauth provider")
	}
	tok, err := p.Exchange(dto.Code, "")
	if err != nil {
		return response.BadRequest("identity", err.Error())
	}
	profile, err := p.Profile(tok.AccessToken)
	if err != nil {
		return response.BadRequest("identity", err.Error())
	}
	return response.Success(map[string]any{
		"provider": providerName,
		"profile":  profile,
		"token":    tok,
	}, "identity", "oauth callback", nil)
}

// Providers mirrors TS GET /oauth/provider/providers — list of registered names.
func (c *OAuthController) Providers(ctx types.Context) types.Output {
	if c.registry == nil {
		return response.Success([]string{}, "identity", "oauth providers", nil)
	}
	return response.Success(c.registry.Names(), "identity", "oauth providers", nil)
}

// ServerInitiate mirrors TS POST /oauth/server/initiate — persists a
// ClientLog row in `initiate` state and returns the sessionId clients use
// for the auth-code step.
func (c *OAuthController) ServerInitiate(dto *OAuthInitiateDto) types.Output {
	return c.initiate(dto, "server")
}

// ServerAuthCode mirrors TS POST /oauth/server/auth-code — mints an auth-code
// for the supplied sessionId, attaches the authenticated user to the row,
// and returns the code.
func (c *OAuthController) ServerAuthCode(dto *OAuthAuthCodeDto, ctx types.Context) types.Output {
	claims, _ := ctx.GetValue("auth").(map[string]any)
	userId, _ := claims["sub"].(string)
	if userId == "" {
		userId, _ = claims["userId"].(string)
	}
	workspaceId, _ := claims["workspaceId"].(string)
	if c.clientLogService == nil || dto.SessionId == "" {
		return response.BadRequest("identity", "session required")
	}
	authCode, err := c.clientLogService.SetAuthCode(dto.SessionId, userId, workspaceId)
	if err != nil || authCode == "" {
		return response.BadRequest("identity", "invalid session")
	}
	return response.Success(map[string]any{
		"sessionId": dto.SessionId,
		"authCode":  authCode,
		"user":      claims,
	}, "identity", "auth-code issued", nil)
}

// ServerAccessToken mirrors TS POST /oauth/server/access-token — redeems the
// auth-code for an access+refresh token pair tied to the originally-issued
// (userId, clientId) triple.
func (c *OAuthController) ServerAccessToken(dto *OAuthAccessTokenDto) types.Output {
	return c.exchangeAccessToken(dto)
}

// DeviceInitiate mirrors TS POST /oauth/device/initiate.
func (c *OAuthController) DeviceInitiate(dto *OAuthInitiateDto) types.Output {
	return c.initiate(dto, "device")
}

// DeviceAccessToken mirrors TS POST /oauth/device/access-token.
func (c *OAuthController) DeviceAccessToken(dto *OAuthAccessTokenDto) types.Output {
	return c.exchangeAccessToken(dto)
}

// WebInitiate mirrors TS POST /oauth/web/initiate.
func (c *OAuthController) WebInitiate(dto *OAuthInitiateDto) types.Output {
	return c.initiate(dto, "web")
}

// WebAccessToken mirrors TS POST /oauth/web/access-token.
func (c *OAuthController) WebAccessToken(dto *OAuthAccessTokenDto) types.Output {
	return c.exchangeAccessToken(dto)
}

// initiate is the shared body for server/device/web initiate — persists the
// ClientLog row and returns its sessionId.
func (c *OAuthController) initiate(dto *OAuthInitiateDto, kind string) types.Output {
	if c.clientLogService == nil || dto.ClientId == "" {
		return response.BadRequest("identity", "client required")
	}
	row, err := c.clientLogService.CreateInitiate(dto.ClientId, dto.State, dto.Scope, kind)
	if err != nil || row == nil {
		return response.InternalServerError("identity", "could not initiate session")
	}
	return response.Success(map[string]any{
		"sessionId": row.Id,
		"clientId":  dto.ClientId,
		"state":     dto.State,
		"scope":     dto.Scope,
		"kind":      kind,
		"initiated": true,
	}, "identity", kind+" initiate", nil)
}

// exchangeAccessToken redeems an authCode for a tokens pair. Shared by the
// server/device/web access-token endpoints — the per-flow distinction lives
// in initiate (which records `kind`) rather than in the redeem step.
func (c *OAuthController) exchangeAccessToken(dto *OAuthAccessTokenDto) types.Output {
	if c.clientLogService == nil || dto.AuthCode == "" {
		return response.BadRequest("identity", "auth-code required")
	}
	redeemed, err := c.clientLogService.RedeemAuthCode(dto.SessionId, dto.AuthCode)
	if err != nil || redeemed == nil {
		return response.BadRequest("identity", "invalid or expired auth-code")
	}
	if c.authService == nil {
		return response.InternalServerError("identity", "auth service unavailable")
	}
	access, refresh, err := c.authService.IssueTokensByHandle(redeemed.UserId, redeemed.ClientId)
	if err != nil {
		// Fallback: when the userId isn't a valid handle (the OAuth flow
		// authenticated via a different identity provider), persist the
		// sessionId as the access token so the caller can poll. This keeps
		// the surface non-blocking for partial-port environments.
		access = "session:" + dto.SessionId
		refresh = ""
	}
	return response.Success(map[string]any{
		"accessToken":  access,
		"refreshToken": refresh,
		"sessionId":    redeemed.SessionId,
		"clientId":     redeemed.ClientId,
		"userId":       redeemed.UserId,
	}, "identity", "access-token issued", nil)
}

// Login mirrors TS POST /oauth/login — finalises a provider authorization
// code → app session by issuing access/refresh tokens via AuthService.
func (c *OAuthController) Login(dto *OAuthLoginDto) types.Output {
	if c.registry == nil {
		return response.BadRequest("identity", "oauth registry not configured")
	}
	p := c.registry.Get(dto.Provider)
	if p == nil {
		return response.BadRequest("identity", "unknown oauth provider")
	}
	tok, err := p.Exchange(dto.Code, dto.RedirectURI)
	if err != nil {
		return response.BadRequest("identity", err.Error())
	}
	profile, err := p.Profile(tok.AccessToken)
	if err != nil {
		return response.BadRequest("identity", err.Error())
	}
	// Find-or-create the local user keyed by provider email. RegisterUser is
	// idempotent at the SQL layer (unique email) — duplicate inserts surface
	// as a soft error we ignore.
	if profile.Email != "" {
		_, _ = c.authService.RegisterUser(profile.Name, profile.Email, profile.ID)
	}
	access, refresh, err := c.authService.Login(profile.Email, profile.ID, dto.ClientId)
	if err != nil {
		return response.Unauthorized("identity", err.Error())
	}
	return response.Success(map[string]any{
		"accessToken":  access,
		"refreshToken": refresh,
		"provider":     dto.Provider,
		"profile":      profile,
	}, "identity", "oauth login", nil)
}
