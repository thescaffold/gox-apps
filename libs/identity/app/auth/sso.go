package auth

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/awesome-goose/goose/io/output"
	"github.com/awesome-goose/goose/types"
	identityuser "github.com/thescaffold/gox-apps/libs/identity/app/user"
	identityotp "github.com/thescaffold/gox-apps/libs/identity/pkg/otp"
	"github.com/thescaffold/gox-apps/libs/identity/pkg/sso"
	"github.com/thescaffold/gox-apps/libs/identity/pkg/tokenstore"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/response"
)

// Single sign-on (OpenID Connect). A workspace registers its identity
// provider and the email domains it owns; people with those emails sign in
// through it. Membership is by invitation only: SSO never creates access for
// someone the company has not invited (see sso.Directory).
//
// The browser flow is:
//
//	POST /sso/start      {email}          → {url}   (the IdP's authorize URL)
//	GET  /sso/callback   ?code&state      → 302 to the app with a one-time ?code
//	POST /sso/exchange   {code}           → the same body as /otp/verify
//
// Tokens never appear in a URL: the callback hands the app a short-lived,
// single-use code that /sso/exchange turns into a session.

const (
	ssoStatePrefix  = "identity:sso:state:"
	ssoResultPrefix = "identity:sso:result:"
	ssoStateTTL     = 10 * time.Minute
	ssoResultTTL    = time.Minute
)

// ssoState is what /start remembers for the callback.
type ssoState struct {
	ConnectionID string `json:"c"`
	Nonce        string `json:"n"`
	Verifier     string `json:"v"`
	ClientID     string `json:"cl"`
	Email        string `json:"e"`
	ReturnTo     string `json:"r"`
}

type ssoResult struct {
	UserID   string `json:"u"`
	ClientID string `json:"c"`
	IsNew    bool   `json:"n"`
}

// SsoDiscoverDto carries POST /sso/discover.
type SsoDiscoverDto struct {
	Ctx      ntxctx.NTXContext `context:"ntx"`
	Email    string            `json:"email" binding:"required"`
	ClientId string            `json:"clientId"`
}

// SsoStartDto carries POST /sso/start.
type SsoStartDto struct {
	Ctx      ntxctx.NTXContext `context:"ntx"`
	Email    string            `json:"email"    binding:"required"`
	ClientId string            `json:"clientId" binding:"required"`
	// ReturnTo is a same-origin path the app resumes at ("/sso/complete").
	ReturnTo string `json:"returnTo"`
}

// SsoCallbackDto carries GET /sso/callback.
type SsoCallbackDto struct {
	Ctx   ntxctx.NTXContext `context:"ntx"`
	Code  string            `query:"code"`
	State string            `query:"state"`
	Error string            `query:"error"`
}

// SsoExchangeDto carries POST /sso/exchange.
type SsoExchangeDto struct {
	Ctx       ntxctx.NTXContext `context:"ntx"`
	Code      string            `json:"code"     binding:"required"`
	UserAgent string            `json:"userAgent"`
	Cookie    string            `header:"Cookie"`
}

// SsoConnectionDto carries PUT /sso/connection.
type SsoConnectionDto struct {
	Ctx          ntxctx.NTXContext `context:"ntx"`
	Issuer       string            `json:"issuer"       binding:"required"`
	OIDCClientID string            `json:"clientId"     binding:"required"`
	// ClientSecret is write-only; leave empty on update to keep the stored one.
	ClientSecret string   `json:"clientSecret"`
	Domains      []string `json:"domains"`
	Enforce      bool     `json:"enforce"`
}

// SsoWorkspaceDto carries GET/DELETE /sso/connection.
type SsoWorkspaceDto struct {
	Ctx ntxctx.NTXContext `context:"ntx"`
}

func ssoBaseURL() string { return strings.TrimRight(os.Getenv("BASE_URL"), "/") }

func ssoRedirectURI() string {
	if v := os.Getenv("SSO_REDIRECT_URL"); v != "" {
		return v
	}
	return ssoBaseURL() + "/api/apps/identity/sso/callback"
}

// safeReturnTo accepts only a same-origin absolute path.
func safeReturnTo(p string) string {
	if p == "" || !strings.HasPrefix(p, "/") || strings.HasPrefix(p, "//") || strings.ContainsAny(p, "\\\r\n") {
		return "/sso/complete"
	}
	return p
}

func (c *AuthController) ssoReady() bool { return c.ssoStore != nil && c.ssoClient != nil }

// SsoDiscover tells the sign-in screen whether an email signs in with SSO,
// and whether it must.
func (c *AuthController) SsoDiscover(dto *SsoDiscoverDto) types.Output {
	title := c.lang.Translate("apps.identity.app.title", nil, dto.Ctx.Preference)
	if !c.ssoReady() {
		return response.Success(map[string]any{"sso": false, "enforced": false}, title, "ok", nil)
	}
	conn, err := c.ssoStore.ForEmail(context.Background(), dto.Email)
	if err != nil {
		return response.Success(map[string]any{"sso": false, "enforced": false}, title, "ok", nil)
	}
	return response.Success(map[string]any{"sso": true, "enforced": conn.Enforce}, title, "ok", nil)
}

// SsoStart begins a sign-in: it returns the identity provider's authorize URL.
func (c *AuthController) SsoStart(dto *SsoStartDto) types.Output {
	title := c.lang.Translate("apps.identity.app.title", nil, dto.Ctx.Preference)
	if !c.ssoReady() {
		return response.NotFound(title, "single sign-on is not available")
	}
	ctx := context.Background()
	conn, err := c.ssoStore.ForEmail(ctx, dto.Email)
	if err != nil {
		return response.NotFound(title, "single sign-on is not set up for this email")
	}
	d, err := c.ssoClient.Discover(ctx, conn.Issuer)
	if err != nil {
		return response.Error(title, "the company's identity provider is not reachable", 502)
	}
	state, e1 := sso.Random(24)
	nonce, e2 := sso.Random(24)
	verifier, challenge, e3 := sso.PKCE()
	if e1 != nil || e2 != nil || e3 != nil {
		return response.InternalServerError(title, "could not start sign-in")
	}
	st := ssoState{ConnectionID: conn.ID, Nonce: nonce, Verifier: verifier, ClientID: dto.ClientId,
		Email: strings.ToLower(strings.TrimSpace(dto.Email)), ReturnTo: safeReturnTo(dto.ReturnTo)}
	b, _ := json.Marshal(st)
	c.cache.Set(ssoStatePrefix+state, string(b), ssoStateTTL)
	return response.Success(map[string]any{
		"url": sso.AuthURL(d, conn.OIDCClientID, ssoRedirectURI(), state, nonce, challenge, st.Email),
	}, title, "ok", nil)
}

// ssoFail sends the browser back to the app with a short error code. No
// detail about why leaves the server: the reasons are for the logs and for
// the company's admin, not for whoever is at the keyboard.
func ssoFail(returnTo, code string) types.Output {
	return output.Redirect(ssoBaseURL() + safeReturnTo(returnTo) + "?error=" + url.QueryEscape(code))
}

// SsoCallback is where the identity provider sends the browser back.
func (c *AuthController) SsoCallback(dto *SsoCallbackDto) types.Output {
	if !c.ssoReady() || dto.State == "" {
		return ssoFail("", "invalid_request")
	}
	raw, ok := c.cache.Get(ssoStatePrefix + dto.State)
	c.cache.Del(ssoStatePrefix + dto.State) // single use, even on failure
	if !ok {
		return ssoFail("", "expired")
	}
	var st ssoState
	if json.Unmarshal([]byte(raw), &st) != nil {
		return ssoFail("", "invalid_request")
	}
	if dto.Error != "" || dto.Code == "" {
		return ssoFail(st.ReturnTo, "denied")
	}
	ctx := context.Background()
	conn, err := c.ssoStore.ByID(ctx, st.ConnectionID)
	if err != nil {
		return ssoFail(st.ReturnTo, "invalid_request")
	}
	secret, err := tokenstore.OpenProviderToken(conn.SecretSealed, tokenstore.Key())
	if err != nil || secret == "" {
		return ssoFail(st.ReturnTo, "misconfigured")
	}
	d, err := c.ssoClient.Discover(ctx, conn.Issuer)
	if err != nil {
		return ssoFail(st.ReturnTo, "provider_unreachable")
	}
	idToken, err := c.ssoClient.Exchange(ctx, d, conn.OIDCClientID, secret, dto.Code, ssoRedirectURI(), st.Verifier)
	if err != nil {
		return ssoFail(st.ReturnTo, "provider_rejected")
	}
	claims, err := c.ssoClient.Verify(ctx, d, idToken, conn.OIDCClientID, st.Nonce)
	if err != nil {
		return ssoFail(st.ReturnTo, "invalid_token")
	}
	// The provider must vouch for the address, and it must be the one this
	// company owns: a token for someone@other.example is not this company's.
	if !claims.EmailVerified {
		return ssoFail(st.ReturnTo, "email_unverified")
	}
	owns := false
	for _, dom := range conn.Domains() {
		if dom == sso.EmailDomain(claims.Email) {
			owns = true
		}
	}
	if !owns {
		return ssoFail(st.ReturnTo, "wrong_domain")
	}
	dir := sso.CurrentDirectory()
	if dir == nil {
		return ssoFail(st.ReturnTo, "misconfigured")
	}
	invited, err := dir.Invited(ctx, conn.WorkspaceID, claims.Email)
	if err != nil {
		return ssoFail(st.ReturnTo, "error")
	}
	if !invited {
		return ssoFail(st.ReturnTo, "not_invited")
	}

	u, _ := c.userEntity.First(`"ref" = ? OR "email" = ?`, claims.Email, claims.Email)
	isNew := u == nil
	if isNew {
		r, t := claims.Email, identityotp.Email
		name := claims.Name
		if name == "" {
			name = claims.Email
		}
		u = &identityuser.User{Ref: &r, Type: &t, Name: name, Email: claims.Email}
		if err := c.userEntity.Insert(u); err != nil {
			return ssoFail(st.ReturnTo, "error")
		}
		c.bootstrapWorkspace(u, st.ClientID)
		c.upsertAttributes(u.Id, "user", map[string]string{t: claims.Email})
		c.seedDefaultRoles()
	} else {
		if closed, _ := c.attributeEntity.First(`"user_id" = ? AND "key" = ?`, u.Id, "closed_at"); closed != nil {
			return ssoFail(st.ReturnTo, "denied")
		}
		if suspended, _ := c.attributeEntity.First(`"user_id" = ? AND "key" = ?`, u.Id, "suspended_at"); suspended != nil {
			return ssoFail(st.ReturnTo, "suspended")
		}
	}
	if err := dir.Joined(ctx, conn.WorkspaceID, u.Id, claims.Email); err != nil {
		return ssoFail(st.ReturnTo, "error")
	}

	code, err := sso.Random(24)
	if err != nil {
		return ssoFail(st.ReturnTo, "error")
	}
	rb, _ := json.Marshal(ssoResult{UserID: u.Id, ClientID: st.ClientID, IsNew: isNew})
	c.cache.Set(ssoResultPrefix+code, string(rb), ssoResultTTL)
	return output.Redirect(ssoBaseURL() + st.ReturnTo + "?code=" + url.QueryEscape(code))
}

// SsoExchange turns the one-time code into a session, in the same shape as a
// passwordless sign-in.
func (c *AuthController) SsoExchange(dto *SsoExchangeDto) types.Output {
	pref := dto.Ctx.Preference
	title := c.lang.Translate("apps.identity.app.title", nil, pref)
	raw, ok := c.cache.Get(ssoResultPrefix + dto.Code)
	c.cache.Del(ssoResultPrefix + dto.Code)
	if !ok {
		return response.Unauthorized(title, "the sign-in link has expired")
	}
	var res ssoResult
	if json.Unmarshal([]byte(raw), &res) != nil {
		return response.Unauthorized(title, "the sign-in link has expired")
	}
	u, _ := c.userEntity.First(`"id" = ?`, res.UserID)
	if u == nil {
		return response.Unauthorized(title, "the sign-in link has expired")
	}
	workspaces := c.workspacesForUser(u.Id, res.ClientID)
	workspaceID := ""
	if len(workspaces) > 0 {
		workspaceID = workspaces[0].Id
	}
	access, refresh, err := c.auth.IssueTokens(u, res.ClientID, workspaceID)
	if err != nil {
		return response.Unauthorized(title, err.Error())
	}
	deviceID := ""
	if len(workspaces) > 0 {
		var sessionID string
		deviceID, sessionID = c.bootstrapDevice(u.Id, res.ClientID, dto.UserAgent, cookieValue(dto.Cookie, DeviceCookie))
		c.logDeviceLogin(deviceID, sessionID, dto.Ctx)
	}
	return withDeviceCookie(response.Success(c.sessionPayload(u, res.ClientID, map[string]any{
		"user":         maskUser(u),
		"accessToken":  access,
		"refreshToken": refresh,
		"isNew":        res.IsNew,
	}), title, "signed in", nil), deviceID)
}

// ssoAdmin authorizes the caller to manage the token's workspace's SSO: they
// must be an owner or admin of it.
func (c *AuthController) ssoAdmin(ntx ntxctx.NTXContext) bool {
	if ntx.UserID == "" || ntx.WorkspaceID == "" {
		return false
	}
	link, _ := c.ucwEntity.First(`"user_id" = ? AND "workspace_id" = ?`, ntx.UserID, ntx.WorkspaceID)
	if link == nil || link.Status == nil {
		return false
	}
	return *link.Status == "owner" || *link.Status == "admin"
}

func connectionView(conn *sso.Connection) map[string]any {
	return map[string]any{
		"issuer": conn.Issuer, "clientId": conn.OIDCClientID, "domains": conn.Domains(),
		"enforce": conn.Enforce, "hasSecret": conn.SecretSealed != "", "updatedAt": conn.UpdatedAt,
		"redirectUri": ssoRedirectURI(),
	}
}

// SsoGetConnection returns the workspace's SSO settings (never the secret).
func (c *AuthController) SsoGetConnection(dto *SsoWorkspaceDto) types.Output {
	title := c.lang.Translate("apps.identity.app.title", nil, dto.Ctx.Preference)
	if !c.ssoReady() {
		return response.NotFound(title, "single sign-on is not available")
	}
	if !c.ssoAdmin(dto.Ctx) {
		return response.Forbidden(title, "only company admins can manage single sign-on")
	}
	conn, err := c.ssoStore.ForWorkspace(context.Background(), dto.Ctx.WorkspaceID)
	if err != nil {
		// Not configured is a normal state; still tell the console the callback
		// URL to register with the provider.
		return response.Success(map[string]any{"configured": false, "redirectUri": ssoRedirectURI()}, title, "ok", nil)
	}
	v := connectionView(conn)
	v["configured"] = true
	return response.Success(v, title, "ok", nil)
}

// SsoPutConnection creates or replaces the workspace's SSO settings. Every
// claimed domain must be one the workspace has verified; the provider must be
// reachable and speak OpenID Connect, so a typo is caught here rather than at
// the first sign-in.
func (c *AuthController) SsoPutConnection(dto *SsoConnectionDto) types.Output {
	title := c.lang.Translate("apps.identity.app.title", nil, dto.Ctx.Preference)
	if !c.ssoReady() {
		return response.NotFound(title, "single sign-on is not available")
	}
	if !c.ssoAdmin(dto.Ctx) {
		return response.Forbidden(title, "only company admins can manage single sign-on")
	}
	ctx := context.Background()
	dir := sso.CurrentDirectory()
	if dir == nil {
		return response.Error(title, "single sign-on is not configured on this server", 503)
	}
	key := tokenstore.Key()
	if len(key) != 32 {
		return response.Error(title, "single sign-on is not configured on this server", 503)
	}

	issuer := strings.TrimRight(strings.TrimSpace(dto.Issuer), "/")
	if _, err := c.ssoClient.Discover(ctx, issuer); err != nil {
		return response.BadRequest(title, "the identity provider could not be reached, or is not an OpenID Connect provider")
	}
	var domains []string
	seen := map[string]bool{}
	for _, raw := range dto.Domains {
		d := strings.ToLower(strings.TrimSpace(raw))
		if d == "" || seen[d] {
			continue
		}
		if !strings.Contains(d, ".") || strings.ContainsAny(d, "@/ ") {
			return response.BadRequest(title, "domains must look like example.com")
		}
		ok, err := dir.OwnsDomain(ctx, dto.Ctx.WorkspaceID, d)
		if err != nil {
			return response.InternalServerError(title, "could not check the domain")
		}
		if !ok {
			return response.BadRequest(title, d+" is not a verified domain of this company")
		}
		seen[d] = true
		domains = append(domains, d)
	}
	if len(domains) == 0 {
		return response.BadRequest(title, "claim at least one of your verified domains")
	}
	// A domain belongs to one company's SSO.
	for _, d := range domains {
		if other, err := c.ssoStore.ForEmail(ctx, "x@"+d); err == nil && other.WorkspaceID != dto.Ctx.WorkspaceID {
			return response.Conflict(title, d+" is already claimed by another company")
		}
	}

	conn := &sso.Connection{WorkspaceID: dto.Ctx.WorkspaceID, Issuer: issuer, OIDCClientID: strings.TrimSpace(dto.OIDCClientID), Enforce: dto.Enforce}
	db, _ := json.Marshal(domains)
	conn.DomainsJSON = db
	if dto.ClientSecret != "" {
		sealed, err := tokenstore.SealProviderToken(dto.ClientSecret, key)
		if err != nil {
			return response.InternalServerError(title, "could not store the secret")
		}
		conn.SecretSealed = sealed
	} else if prev, err := c.ssoStore.ForWorkspace(ctx, dto.Ctx.WorkspaceID); err == nil {
		conn.SecretSealed = prev.SecretSealed
	}
	if conn.SecretSealed == "" {
		return response.BadRequest(title, "a client secret is required")
	}
	if err := c.ssoStore.Save(ctx, conn); err != nil {
		return response.InternalServerError(title, "could not save")
	}
	v := connectionView(conn)
	v["configured"] = true
	return response.Success(v, title, "saved", nil)
}

// SsoDeleteConnection removes the workspace's SSO (and its enforcement).
func (c *AuthController) SsoDeleteConnection(dto *SsoWorkspaceDto) types.Output {
	title := c.lang.Translate("apps.identity.app.title", nil, dto.Ctx.Preference)
	if !c.ssoReady() {
		return response.NotFound(title, "single sign-on is not available")
	}
	if !c.ssoAdmin(dto.Ctx) {
		return response.Forbidden(title, "only company admins can manage single sign-on")
	}
	if err := c.ssoStore.Delete(context.Background(), dto.Ctx.WorkspaceID); err != nil {
		return response.InternalServerError(title, "could not remove")
	}
	return response.Success(map[string]any{"configured": false}, title, "removed", nil)
}

// SsoRecoveryDto carries POST /sso/recovery/start and /sso/recovery/verify.
type SsoRecoveryDto struct {
	Ctx   ntxctx.NTXContext `context:"ntx"`
	Email string            `json:"email" binding:"required"`
}

// SsoRecoveryStart is the break-glass for a company locked out of its
// identity provider: it names a DNS TXT record that, once published, proves
// control of the domain.
func (c *AuthController) SsoRecoveryStart(dto *SsoRecoveryDto) types.Output {
	title := c.lang.Translate("apps.identity.app.title", nil, dto.Ctx.Preference)
	if !c.ssoReady() {
		return response.NotFound(title, "single sign-on is not available")
	}
	ch, err := c.ssoStore.StartRecovery(context.Background(), dto.Email)
	switch {
	case errors.Is(err, sso.ErrRecoveryNotNeeded):
		return response.NotFound(title, "this address does not require single sign-on")
	case errors.Is(err, sso.ErrRecoveryThrottled):
		return response.Error(title, "too many recovery requests; try again later", 429)
	case err != nil:
		return response.InternalServerError(title, "could not start recovery")
	}
	return response.Success(ch, title, "publish this record, then check it", nil)
}

// SsoRecoveryVerify checks the TXT record. On success enforcement is lifted
// for the address for 30 minutes; the person still signs in with an emailed
// code or a password.
func (c *AuthController) SsoRecoveryVerify(dto *SsoRecoveryDto) types.Output {
	title := c.lang.Translate("apps.identity.app.title", nil, dto.Ctx.Preference)
	if !c.ssoReady() {
		return response.NotFound(title, "single sign-on is not available")
	}
	until, err := c.ssoStore.VerifyRecovery(context.Background(), dto.Email)
	switch {
	case errors.Is(err, sso.ErrRecoveryPending):
		return response.NotFound(title, "start recovery first")
	case errors.Is(err, sso.ErrRecoveryNotFound):
		return response.Success(map[string]any{"granted": false}, title, "the record was not found yet; DNS can take a few minutes", nil)
	case err != nil:
		return response.InternalServerError(title, "could not check the record")
	}
	slog.Warn("sso: enforcement lifted by DNS recovery", "email_domain", sso.EmailDomain(dto.Email), "until", until)
	return response.Success(map[string]any{"granted": true, "until": until}, title, "you can now sign in with an emailed code", nil)
}

// ssoEnforcedResponse is the answer to a password or code sign-in for an
// address whose company requires SSO.
func (c *AuthController) ssoEnforcedResponse(pref map[string]any) types.Output {
	title := c.lang.Translate("apps.identity.app.title", nil, pref)
	return response.Forbidden(title, "your company requires single sign-on; use \"Sign in with SSO\"")
}
