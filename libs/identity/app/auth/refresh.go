package auth

import (
	"github.com/awesome-goose/goose/types"
	"github.com/thescaffold/gox-apps/libs/identity/pkg/revocation"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/response"
)

// RefreshDto carries POST /refresh.
type RefreshDto struct {
	Ctx ntxctx.NTXContext `context:"ntx"`
	// Token is the refresh token Login, OtpVerify or SwitchWorkspace returned.
	Token string `json:"token" binding:"required"`
}

// Refresh exchanges a refresh token for a new access token, so a native app
// can stay signed in after the short-lived access token expires. The browser
// shell does not need it: it re-mints through its device cookie (Sessions).
//
// Every way a session can end is checked here, not just the token's signature:
// the refresh token must still be on record (sign-out and account closing
// delete it), must not predate a "sign out everywhere", and the account must
// not be closed or suspended. Claims are rebuilt from the database, so a role
// or permission removed since sign-in is not carried over, and a workspace the
// person has left is refused. Only an access token is issued and nothing is
// written: the refresh token stays the credential and is not rotated.
//
// Every refusal is the same 401, so the response says nothing about why.
func (c *AuthController) Refresh(dto *RefreshDto) types.Output {
	pref := dto.Ctx.Preference
	title := c.lang.Translate("apps.identity.app.title", nil, pref)
	refuse := func() types.Output {
		return response.Unauthorized(title, c.lang.Translate("apps.identity.app.post.login.error.invalid-request", nil, pref))
	}

	claims, err := c.auth.ValidateToken(dto.Token)
	if err != nil || revocation.Check(claims) {
		return refuse()
	}
	// On record, unexpired, and of the refresh kind (an access token is not).
	if _, err := c.auth.RefreshToken(dto.Token); err != nil {
		return refuse()
	}
	sub, _ := claims["sub"].(string)
	clientID, _ := claims["clientId"].(string)
	workspaceID, _ := claims["workspaceId"].(string)
	if sub == "" || clientID == "" {
		return refuse()
	}
	u, _ := c.userEntity.First(`"id" = ?`, sub)
	if u == nil || !c.usable(u.Id) {
		return refuse()
	}
	if workspaceID != "" {
		if link, _ := c.ucwEntity.First(`"user_id" = ? AND "client_id" = ? AND "workspace_id" = ?`, u.Id, clientID, workspaceID); link == nil {
			return refuse()
		}
	}

	access, err := c.auth.IssueAccessToken(u, clientID, workspaceID)
	if err != nil {
		return response.InternalServerError(title, "could not sign you in")
	}
	return response.Success(map[string]any{
		"accessToken": access,
		"user":        maskUser(u),
	}, title, c.lang.Translate("apps.identity.app.post.login.success", nil, pref), nil)
}
