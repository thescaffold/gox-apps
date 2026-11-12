package auth

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/url"
	"os"
	"strings"

	"github.com/awesome-goose/goose/io/output"
	"github.com/awesome-goose/goose/types"
	identitydevicesession "github.com/thescaffold/gox-apps/libs/identity/app/devicesession"
	"github.com/thescaffold/gox-apps/libs/identity/pkg/sso"
	"github.com/thescaffold/gox-packages/libs/core/response"
)

// SAML single logout (V4-03). See sso.go's file comment for the flow.

// sloState is what an SP-initiated logout remembers for the IdP's answer.
type sloState struct {
	ConnectionID string `json:"c"`
	RequestID    string `json:"q"`
}

// ssoSignedOutURL is where the browser lands once single logout is done.
func ssoSignedOutURL(errCode string) string {
	p := os.Getenv("SSO_SIGNED_OUT_PATH")
	if p == "" || !strings.HasPrefix(p, "/") || strings.HasPrefix(p, "//") {
		p = "/guard/login?signedout=sso"
	}
	if errCode != "" {
		sep := "?"
		if strings.Contains(p, "?") {
			sep = "&"
		}
		p += sep + "error=" + url.QueryEscape(errCode)
	}
	return ssoBaseURL() + p
}

// samlLogoutStart makes the signed LogoutRequest URL for a SAML session that is being signed out
// here, or "" when its connection offers no single logout.
func (c *AuthController) samlLogoutStart(connectionID, nameID, sessionIndex string) string {
	if !c.ssoReady() || connectionID == "" || nameID == "" {
		return ""
	}
	ctx := context.Background()
	conn, err := c.ssoStore.ByID(ctx, connectionID)
	if err != nil || !conn.SAML() {
		return ""
	}
	idp, err := c.samlIdP(ctx, conn)
	if err != nil {
		return ""
	}
	sp, _ := c.samlSPFor(ctx, conn.WorkspaceID, conn, false)
	state, err := sso.Random(24)
	if err != nil {
		return ""
	}
	redirect, requestID, err := c.ssoClient.SAMLLogoutRequest(sp, idp, nameID, sessionIndex, state)
	if err != nil {
		return ""
	}
	b, _ := json.Marshal(sloState{ConnectionID: conn.ID, RequestID: requestID})
	c.cache.Set(ssoSloPrefix+state, string(b), ssoStateTTL)
	return redirect
}

// SsoSamlSlo is a workspace's SingleLogoutService (HTTP-Redirect binding). A LogoutResponse
// answers the LogoutRequest this side sent when the person signed out here: the browser goes to
// the signed-out page. A LogoutRequest is the IdP signing the person out: the sessions it names
// (by NameID, and SessionIndex when given) end here, and the browser goes back to the IdP with
// the answer. Anything not signed by the connection's IdP is refused and ends nothing.
func (c *AuthController) SsoSamlSlo(dto *SsoSamlSloDto) types.Output {
	title := c.lang.Translate("apps.identity.app.title", nil, dto.Ctx.Preference)
	if !c.ssoReady() {
		return response.NotFound(title, "single sign-on is not available")
	}
	q, err := url.ParseQuery(dto.Query)
	if err != nil {
		return response.BadRequest(title, "not a SAML logout message")
	}
	ctx := context.Background()
	conn, err := c.ssoStore.ForWorkspace(ctx, dto.Workspace)
	if err != nil || !conn.SAML() {
		return response.NotFound(title, "no SAML connection")
	}
	idp, err := c.samlIdP(ctx, conn)
	if err != nil {
		return response.Error(title, "the company's identity provider is not usable", 502)
	}
	sp, _ := c.samlSPFor(ctx, conn.WorkspaceID, conn, false)

	switch {
	case q.Get("SAMLResponse") != "":
		key := ssoSloPrefix + q.Get("RelayState")
		raw, ok := c.cache.Get(key)
		c.cache.Del(key) // single use
		var st sloState
		if !ok || json.Unmarshal([]byte(raw), &st) != nil || st.ConnectionID != conn.ID {
			return output.Redirect(ssoSignedOutURL("slo_expired"))
		}
		if err := c.ssoClient.SAMLLogoutResponseVerify(sp, idp, dto.Query, []string{st.RequestID}); err != nil {
			slog.Warn("sso: SAML LogoutResponse refused", "workspace", conn.WorkspaceID, "reason", err.Error())
			return output.Redirect(ssoSignedOutURL("slo_refused"))
		}
		return output.Redirect(ssoSignedOutURL(""))

	case q.Get("SAMLRequest") != "":
		lr, err := c.ssoClient.SAMLLogoutRequestVerify(sp, idp, dto.Query, c.ssoStore.ReplayCache(conn.ID))
		if err != nil {
			slog.Warn("sso: SAML LogoutRequest refused", "workspace", conn.WorkspaceID, "reason", err.Error())
			return response.BadRequest(title, "the logout request could not be verified")
		}
		ended := c.endSAMLSessions(conn.ID, lr)
		back, err := c.ssoClient.SAMLLogoutResponse(sp, idp, lr.ID, lr.RelayState, ended)
		if err != nil {
			return response.InternalServerError(title, "could not answer the identity provider")
		}
		return output.Redirect(back)
	}
	return response.BadRequest(title, "not a SAML logout message")
}

// endSAMLSessions signs out the device sessions an IdP's LogoutRequest names. It reports false
// only when they could not be ended (none matching is a success: they are already gone).
func (c *AuthController) endSAMLSessions(connectionID string, lr *sso.SAMLLogout) bool {
	if c.deviceSessionEntity == nil {
		return false
	}
	where, args := `"sso_connection_id" = ? AND "sso_name_id" = ?`, []any{connectionID, lr.NameID}
	if len(lr.SessionIndexes) > 0 {
		where += ` AND "sso_session_index" IN ?`
		args = append(args, lr.SessionIndexes)
	}
	rows, err := c.deviceSessionEntity.Find(0, 0, where, args...)
	if err != nil {
		return false
	}
	byUser := map[string][]identitydevicesession.DeviceSession{}
	for _, r := range rows {
		byUser[r.UserId] = append(byUser[r.UserId], r)
	}
	for u, rs := range byUser {
		if err := c.endSessions(u, rs); err != nil {
			return false
		}
	}
	return true
}
