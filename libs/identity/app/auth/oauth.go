package auth

import (
	"os"

	"github.com/awesome-goose/goose/types"
	identityclientlog "github.com/thescaffold/gox-apps/libs/identity/app/clientlog"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/response"
	"github.com/thescaffold/gox-packages/libs/core/utils"
)

// The browser sign-in hop. A client app with no session POSTs oauth/web/initiate, is sent to the
// sign-in screens (guard) with the returned state and scope, which POST oauth/server/initiate,
// sign the person in, mint an auth code (oauth/server/auth-code) and send them back with it; the
// client app redeems the code at oauth/web/access-token. Every step is one ClientLog row, so the
// dance can be audited. Without these routes the sign-in page never finishes loading.

// OAuthInitiateDto carries the three initiate bodies.
type OAuthInitiateDto struct {
	Ctx      ntxctx.NTXContext `context:"ntx"`
	ClientId string            `json:"clientId" binding:"required"`
	State    string            `json:"state"`
	Scope    string            `json:"scope"`
}

// OAuthAuthCodeDto carries POST oauth/server/auth-code.
type OAuthAuthCodeDto struct {
	Ctx       ntxctx.NTXContext `context:"ntx"`
	SessionId string            `json:"sessionId" binding:"required"`
}

// OAuthAccessTokenDto carries the three access-token bodies. The browser start page sends the
// returned code and state swapped (it posts the URL's `code` as `state` and its `state` as
// `authCode`), so either field may hold the code.
type OAuthAccessTokenDto struct {
	Ctx       ntxctx.NTXContext `context:"ntx"`
	ClientId  string            `json:"clientId"`
	AuthCode  string            `json:"authCode"`
	State     string            `json:"state"`
	SessionId string            `json:"sessionId" binding:"required"`
}

func (c *AuthController) clientLogs() *identityclientlog.ClientLogService {
	if c.clientLog == nil && c.clientLogEntity != nil {
		c.clientLog = identityclientlog.NewService(c.clientLogEntity)
	}
	return c.clientLog
}

func (c *AuthController) OAuthWebInitiate(dto *OAuthInitiateDto) types.Output {
	return c.oauthInitiate(dto, "web")
}

func (c *AuthController) OAuthServerInitiate(dto *OAuthInitiateDto) types.Output {
	return c.oauthInitiate(dto, "server")
}

func (c *AuthController) OAuthDeviceInitiate(dto *OAuthInitiateDto) types.Output {
	return c.oauthInitiate(dto, "device")
}

// oauthInitiate records the start of a sign-in hop and answers with what the next screen needs:
// the session id, a state the caller can check on return (generated when it sent none), the
// scope, and the client's name and return address. `id` is the client id, which the sign-in
// screens key their storage on.
func (c *AuthController) oauthInitiate(dto *OAuthInitiateDto, kind string) types.Output {
	pref := dto.Ctx.Preference
	title := c.lang.Translate("apps.identity.app.title", nil, pref)
	logs := c.clientLogs()
	if logs == nil || dto.ClientId == "" {
		return response.BadRequest(title, "client required")
	}
	state, scope := dto.State, dto.Scope
	if state == "" {
		state = utils.Random(24)
	}
	if scope == "" {
		scope = "login"
	}
	row, err := logs.CreateInitiate(dto.ClientId, state, scope, kind)
	if err != nil || row == nil {
		return response.InternalServerError(title, "could not initiate session")
	}
	out := map[string]any{
		"id":          dto.ClientId,
		"clientId":    dto.ClientId,
		"sessionId":   row.Id,
		"state":       state,
		"scope":       scope,
		"kind":        kind,
		"redirectUrl": oauthRedirectURL(),
	}
	if c.clientEntity != nil {
		if cl, _ := c.clientEntity.First(`"id" = ? OR "key" = ?`, dto.ClientId, dto.ClientId); cl != nil {
			out["name"] = cl.Name
		}
	}
	return response.Success(out, title, kind+" initiate", nil)
}

// oauthRedirectURL is where the sign-in screens send the person back with their code: the
// client app's start page, unless OAUTH_REDIRECT_URL says otherwise.
func oauthRedirectURL() string {
	if v := os.Getenv("OAUTH_REDIRECT_URL"); v != "" {
		return v
	}
	return "/start"
}

// OAuthAuthCode mints the single-use code for a session, tied to the signed-in person and the
// workspace their token is scoped to.
func (c *AuthController) OAuthAuthCode(dto *OAuthAuthCodeDto) types.Output {
	pref := dto.Ctx.Preference
	title := c.lang.Translate("apps.identity.app.title", nil, pref)
	logs := c.clientLogs()
	if logs == nil || dto.Ctx.UserID == "" {
		return response.Unauthorized(title, "sign in first")
	}
	code, err := logs.SetAuthCode(dto.SessionId, dto.Ctx.UserID, dto.Ctx.WorkspaceID)
	if err != nil || code == "" {
		return response.BadRequest(title, "invalid session")
	}
	return response.Success(map[string]any{
		"sessionId": dto.SessionId,
		"authCode":  code,
		"state":     code,
		"user":      dto.Ctx.User,
	}, title, "auth-code issued", nil)
}

// OAuthAccessToken redeems a session's code, once, for the same session payload Login returns.
func (c *AuthController) OAuthAccessToken(dto *OAuthAccessTokenDto) types.Output {
	pref := dto.Ctx.Preference
	title := c.lang.Translate("apps.identity.app.title", nil, pref)
	logs := c.clientLogs()
	if logs == nil {
		return response.BadRequest(title, "auth-code required")
	}
	var redeemed *identityclientlog.RedeemedClientLog
	for _, code := range []string{dto.AuthCode, dto.State} {
		if code == "" {
			continue
		}
		if r, _ := logs.RedeemAuthCode(dto.SessionId, code); r != nil {
			redeemed = r
			break
		}
	}
	if redeemed == nil {
		return response.BadRequest(title, "invalid or expired auth-code")
	}
	u, err := c.userEntity.First(`"id" = ?`, redeemed.UserId)
	if err != nil || u == nil || !c.usable(u.Id) {
		return response.Unauthorized(title, "invalid or expired auth-code")
	}
	access, refresh, err := c.auth.IssueTokens(u, redeemed.ClientId, redeemed.WorkspaceId)
	if err != nil {
		return response.Unauthorized(title, "invalid or expired auth-code")
	}
	return response.Success(map[string]any{
		"user":         maskUser(u),
		"accessToken":  access,
		"refreshToken": refresh,
		"workspaces":   c.workspacesForUser(u.Id, redeemed.ClientId),
		"attributes":   c.attributesForUser(u.Id),
	}, title, "access-token issued", nil)
}
