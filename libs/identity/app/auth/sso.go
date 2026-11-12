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
	"github.com/thescaffold/gox-apps/libs/identity/pkg/org"
	identityotp "github.com/thescaffold/gox-apps/libs/identity/pkg/otp"
	"github.com/thescaffold/gox-apps/libs/identity/pkg/sso"
	"github.com/thescaffold/gox-apps/libs/identity/pkg/tokenstore"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/response"
)

// Single sign-on (OpenID Connect or SAML 2.0). A workspace registers its
// identity provider and the email domains it owns; people with those emails
// sign in through it. Membership is by invitation only: SSO never creates
// access for someone the company has not invited (see sso.Directory).
//
// The browser flow is:
//
//	POST /sso/start      {email}          → {url}   (the IdP's authorize URL)
//	GET  /sso/callback   ?code&state      → 302 to the app with a one-time ?code
//	POST /sso/exchange   {code}           → the same body as /otp/verify
//
// A SAML connection has the same start and exchange; /start's url carries an
// AuthnRequest (HTTP-Redirect), and the IdP posts its Response back to the
// workspace's ACS instead of the callback:
//
//	POST /sso/saml/:workspace/acs   SAMLResponse&RelayState → 302 as the callback
//	GET  /sso/saml/:workspace       this service provider's metadata
//	GET  /sso/saml/:workspace/slo   single logout, either side starting (V4-03)
//
// The service provider has its own signing key (made on first use, rotated in
// two steps at POST /sso/connection/sp-key): it signs AuthnRequests when the
// connection says so, and every logout message. Signing out of a browser whose
// session came through SAML ends it here and returns the IdP's logout URL
// (logoutUrl) for the browser to go on to; a LogoutRequest from the IdP ends
// the sessions it names. IdP-initiated sign-in stays refused (D-27).
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
	// RequestID is the SAML AuthnRequest the response must answer.
	RequestID string `json:"q,omitempty"`
}

type ssoResult struct {
	UserID   string `json:"u"`
	ClientID string `json:"c"`
	IsNew    bool   `json:"n"`
	// A SAML sign-in: what single logout names the session by.
	SAML *ssoSession `json:"s,omitempty"`
}

// ssoSession is a SAML session as the IdP knows it.
type ssoSession struct {
	ConnectionID string `json:"c"`
	NameID       string `json:"n"`
	SessionIndex string `json:"i"`
}

const (
	ssoMfaPrefix = "identity:sso:mfa:" // a SAML session waiting on the second factor
	ssoSloPrefix = "identity:sso:slo:" // a LogoutRequest this side sent
)

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
	Ctx ntxctx.NTXContext `context:"ntx"`
	// Protocol is "oidc" (the default) or "saml".
	Protocol string `json:"protocol"`
	// OpenID Connect: the issuer and client.
	Issuer       string `json:"issuer"`
	OIDCClientID string `json:"clientId"`
	// ClientSecret is write-only; leave empty on update to keep the stored one.
	ClientSecret string `json:"clientSecret"`
	// SAML: the IdP's metadata, by URL (fetched now and refreshed hourly) or
	// pasted. Neither on update keeps the stored metadata.
	MetadataURL string   `json:"metadataUrl"`
	MetadataXML string   `json:"metadataXml"`
	Domains     []string `json:"domains"`
	Enforce     bool     `json:"enforce"`
	// SignRequests (SAML) signs the AuthnRequests with the service provider's key; nil keeps
	// what was set (a new connection follows the IdP's WantAuthnRequestsSigned).
	SignRequests *bool `json:"signRequests"`
}

// SsoSpKeyDto carries POST /sso/connection/sp-key {action}: "rotate" stages a next key
// (published beside the current one), "activate" makes it the one that signs, "cancel" drops it.
type SsoSpKeyDto struct {
	Ctx    ntxctx.NTXContext `context:"ntx"`
	Action string            `json:"action" binding:"required"`
}

// SsoSamlSloDto carries GET /sso/saml/:workspace/slo (HTTP-Redirect binding): the query exactly
// as sent, which the signature covers.
type SsoSamlSloDto struct {
	Ctx       ntxctx.NTXContext `context:"ntx"`
	Workspace string            `param:"workspace"`
	Query     string            `raw:"query"`
}

// SsoSamlMetadataDto carries GET /sso/saml/:workspace.
type SsoSamlMetadataDto struct {
	Ctx       ntxctx.NTXContext `context:"ntx"`
	Workspace string            `param:"workspace"`
}

// SsoSamlAcsDto carries POST /sso/saml/:workspace/acs (HTTP-POST binding).
type SsoSamlAcsDto struct {
	Ctx          ntxctx.NTXContext `context:"ntx"`
	Workspace    string            `param:"workspace"`
	SAMLResponse string            `form:"SAMLResponse"`
	RelayState   string            `form:"RelayState"`
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

// ssoSAMLBase is where the workspaces' SAML service providers live: each is
// <base>/<workspace id>, its ACS <base>/<workspace id>/acs.
func ssoSAMLBase() string {
	if v := os.Getenv("SSO_SAML_BASE_URL"); v != "" {
		return strings.TrimRight(v, "/")
	}
	return ssoBaseURL() + "/api/apps/identity/sso/saml"
}

// samlSP is a workspace's service provider. One per workspace (as is the
// connection), so its URLs are known before the connection is saved: the
// company's admin registers them with the IdP first.
func samlSP(workspaceID string) sso.SAMLServiceProvider {
	e := ssoSAMLBase() + "/" + url.PathEscape(workspaceID)
	return sso.SAMLServiceProvider{EntityID: e, ACSURL: e + "/acs", SLOURL: e + "/slo"}
}

// samlSPFor is the workspace's service provider with its signing key (made now when create and
// there is none) and conn's signing choice. Without a usable key (no IDENTITY_TOKEN_KEY to seal
// it), nothing is signed and there is no single logout.
func (c *AuthController) samlSPFor(ctx context.Context, workspaceID string, conn *sso.Connection, create bool) (sso.SAMLServiceProvider, *sso.SPKeys) {
	sp := samlSP(workspaceID)
	if conn != nil {
		sp.SignRequests = conn.SignRequests
	}
	if c.ssoStore == nil {
		return sp, nil
	}
	seal := tokenstore.Key()
	keys, err := c.ssoStore.SPKeysFor(ctx, workspaceID)
	if errors.Is(err, sso.ErrNotFound) && create && len(seal) == 32 {
		kp, kerr := sso.NewSPKeyPair(sp.EntityID)
		if kerr != nil {
			return sp, nil
		}
		sealed, kerr := tokenstore.SealProviderToken(kp.KeyPEM(), seal)
		if kerr != nil {
			return sp, nil
		}
		keys, err = c.ssoStore.CreateSPKeys(ctx, &sso.SPKeys{WorkspaceID: workspaceID, KeySealed: sealed, Cert: kp.CertB64()})
	}
	if err != nil || keys == nil {
		return sp, nil
	}
	if pemKey, err := tokenstore.OpenProviderToken(keys.KeySealed, seal); err == nil {
		if kp, err := sso.ParseSPKeyPair(pemKey, keys.Cert); err == nil {
			sp.Key, sp.Cert = kp.Key, kp.Cert
		}
	}
	if keys.NextCert != "" {
		if kp, err := sso.ParseSPCert(keys.NextCert); err == nil {
			sp.NextCert = kp.Cert
		}
	}
	return sp, keys
}

// spKeyView is what the console shows of the service provider's key: the certificate(s) to give
// the IdP, their fingerprints, and when it last rotated.
func spKeyView(sp sso.SAMLServiceProvider, keys *sso.SPKeys) map[string]any {
	v := map[string]any{"spSloUrl": sp.SLOURL, "spCert": "", "spFingerprint": "", "spNextCert": "", "spNextFingerprint": "", "spRotatedAt": nil}
	if sp.Cert != nil {
		kp := &sso.SPKeyPair{Cert: sp.Cert}
		v["spCert"], v["spFingerprint"], v["spCertExpiresAt"] = kp.CertPEM(), kp.Fingerprint(), sp.Cert.NotAfter
	}
	if sp.NextCert != nil {
		kp := &sso.SPKeyPair{Cert: sp.NextCert}
		v["spNextCert"], v["spNextFingerprint"] = kp.CertPEM(), kp.Fingerprint()
	}
	if keys != nil {
		v["spRotatedAt"] = keys.RotatedAt
	}
	return v
}

// samlIdP loads a SAML connection's IdP: from its metadata URL (cached for an
// hour, so key rotations are picked up), else, or when that fails, from the
// copy stored when it was saved.
func (c *AuthController) samlIdP(ctx context.Context, conn *sso.Connection) (*sso.SAMLIdP, error) {
	if conn.SAMLMetadataURL != "" {
		if raw, err := c.ssoClient.FetchSAMLMetadata(ctx, conn.SAMLMetadataURL, false); err == nil {
			if idp, err := c.ssoClient.ParseSAMLMetadata(raw); err == nil && idp.EntityID == conn.Issuer {
				return idp, nil
			}
		}
	}
	return c.ssoClient.ParseSAMLMetadata([]byte(conn.SAMLMetadata))
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
	return response.Success(map[string]any{"sso": true, "enforced": conn.Enforce, "protocol": conn.Protocol}, title, "ok", nil)
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
	if conn.SAML() {
		return c.samlStart(ctx, conn, dto, title)
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

// samlStart is SsoStart for a SAML connection: the url is the IdP's sign-on
// service with an AuthnRequest; RelayState is the state key.
func (c *AuthController) samlStart(ctx context.Context, conn *sso.Connection, dto *SsoStartDto, title string) types.Output {
	idp, err := c.samlIdP(ctx, conn)
	if err != nil {
		return response.Error(title, "the company's identity provider is not reachable", 502)
	}
	state, err := sso.Random(24)
	if err != nil {
		return response.InternalServerError(title, "could not start sign-in")
	}
	sp, _ := c.samlSPFor(ctx, conn.WorkspaceID, conn, false)
	redirect, requestID, err := c.ssoClient.SAMLAuthnRequest(sp, idp, state)
	if err != nil {
		return response.InternalServerError(title, "could not start sign-in")
	}
	st := ssoState{ConnectionID: conn.ID, RequestID: requestID, ClientID: dto.ClientId,
		Email: strings.ToLower(strings.TrimSpace(dto.Email)), ReturnTo: safeReturnTo(dto.ReturnTo)}
	b, _ := json.Marshal(st)
	c.cache.Set(ssoStatePrefix+state, string(b), ssoStateTTL)
	return response.Success(map[string]any{"url": redirect}, title, "ok", nil)
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
	return c.ssoAdmit(ctx, conn, st, claims)
}

// SsoSamlAcs is where a SAML IdP posts its Response (the assertion consumer
// service). Only answers to a sign-in this server started are taken: the
// RelayState names the state SsoStart kept, which holds the AuthnRequest id
// the response must answer. A response with no state (IdP-initiated) is
// refused.
func (c *AuthController) SsoSamlAcs(dto *SsoSamlAcsDto) types.Output {
	if !c.ssoReady() || dto.RelayState == "" {
		return ssoFail("", "invalid_request")
	}
	raw, ok := c.cache.Get(ssoStatePrefix + dto.RelayState)
	c.cache.Del(ssoStatePrefix + dto.RelayState) // single use, even on failure
	if !ok {
		return ssoFail("", "expired")
	}
	var st ssoState
	if json.Unmarshal([]byte(raw), &st) != nil || st.RequestID == "" {
		return ssoFail("", "invalid_request")
	}
	ctx := context.Background()
	conn, err := c.ssoStore.ByID(ctx, st.ConnectionID)
	if err != nil || !conn.SAML() || conn.WorkspaceID != dto.Workspace {
		return ssoFail(st.ReturnTo, "invalid_request")
	}
	if dto.SAMLResponse == "" {
		return ssoFail(st.ReturnTo, "denied")
	}
	idp, err := c.samlIdP(ctx, conn)
	if err != nil {
		return ssoFail(st.ReturnTo, "misconfigured")
	}
	claims, err := c.ssoClient.SAMLVerify(samlSP(conn.WorkspaceID), idp, dto.SAMLResponse, st.RequestID, c.ssoStore.ReplayCache(conn.ID))
	if errors.Is(err, sso.ErrSAMLDenied) {
		return ssoFail(st.ReturnTo, "denied")
	}
	if err != nil {
		// The reason is for the company's admin, in the logs.
		slog.Warn("sso: SAML response refused", "workspace", conn.WorkspaceID, "reason", err.Error())
		return ssoFail(st.ReturnTo, "invalid_token")
	}
	return c.ssoAdmit(ctx, conn, st, claims)
}

// SsoSamlMetadata is a workspace's service-provider metadata, for the
// company's IdP admin (it holds the entity id, ACS, single logout service and
// signing certificate; nothing secret).
func (c *AuthController) SsoSamlMetadata(dto *SsoSamlMetadataDto) types.Output {
	title := c.lang.Translate("apps.identity.app.title", nil, dto.Ctx.Preference)
	if !c.ssoReady() || dto.Workspace == "" || len(dto.Workspace) > 64 {
		return response.NotFound(title, "no such workspace")
	}
	if ws, _ := c.workspaceEntity.First(`"id" = ?`, dto.Workspace); ws == nil {
		return response.NotFound(title, "no such workspace")
	}
	conn, _ := c.ssoStore.ForWorkspace(context.Background(), dto.Workspace)
	sp, _ := c.samlSPFor(context.Background(), dto.Workspace, conn, false)
	md, err := c.ssoClient.SAMLMetadata(sp)
	if err != nil {
		return response.InternalServerError(title, "could not build the metadata")
	}
	return xmlOutput(md)
}

// xmlOutput serves XML as it is.
type xmlOutput []byte

func (o xmlOutput) Data() any                  { return []byte(o) }
func (o xmlOutput) Code() int                  { return 200 }
func (o xmlOutput) Headers() map[string]string { return map[string]string{"Cache-Control": "no-cache"} }
func (o xmlOutput) ContentType() string        { return "application/samlmetadata+xml" }

// ssoAdmit is the end of every protocol's sign-in: the verified claims must be
// an address of a domain the company claimed, invited to it; the person is
// found or made, joins, and the browser goes back to the app with a one-time
// code.
func (c *AuthController) ssoAdmit(ctx context.Context, conn *sso.Connection, st ssoState, claims *sso.Claims) types.Output {
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
		ws := c.bootstrapWorkspace(u, st.ClientID)
		c.upsertAttributes(u.Id, "user", map[string]string{t: claims.Email})
		c.seedDefaultRoles()
		identityuser.Created(c.tracker, u, st.ClientID, wsID(ws), "sso")
	} else {
		if closed, _ := c.attributeEntity.First(`"user_id" = ? AND "key" = ?`, u.Id, "closed_at"); closed != nil {
			return ssoFail(st.ReturnTo, "denied")
		}
		if suspended, _ := c.attributeEntity.First(`"user_id" = ? AND "key" = ?`, u.Id, "suspended_at"); suspended != nil {
			return ssoFail(st.ReturnTo, "suspended")
		}
	}
	if err := dir.Joined(sso.WithClientID(ctx, st.ClientID), conn.WorkspaceID, u.Id, claims.Email); err != nil {
		return ssoFail(st.ReturnTo, "error")
	}

	code, err := sso.Random(24)
	if err != nil {
		return ssoFail(st.ReturnTo, "error")
	}
	res := ssoResult{UserID: u.Id, ClientID: st.ClientID, IsNew: isNew}
	if conn.SAML() && claims.Subject != "" {
		res.SAML = &ssoSession{ConnectionID: conn.ID, NameID: claims.Subject, SessionIndex: claims.SessionIndex}
	}
	rb, _ := json.Marshal(res)
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
	if out, stop := c.secondStep(u, res.ClientID, dto.UserAgent, pref, func(challenge string) {
		// The device session is made after the second factor: keep the SAML session for it.
		if res.SAML != nil {
			b, _ := json.Marshal(res.SAML)
			c.cache.Set(ssoMfaPrefix+challenge, string(b), 10*time.Minute)
		}
	}); stop {
		return out
	}
	workspaces := c.workspacesForUser(u.Id, res.ClientID)
	workspaceID := ""
	if len(workspaces) > 0 {
		workspaceID = workspaces[0].Id
	}
	deviceID, sessionID := c.bootstrapDevice(u.Id, res.ClientID, dto.UserAgent, cookieValue(dto.Cookie, DeviceCookie))
	c.linkSSOSession(sessionID, res.SAML)
	access, refresh, err := c.auth.IssueTokens(u, res.ClientID, workspaceID, sessionID)
	if err != nil {
		return response.Unauthorized(title, err.Error())
	}
	// A device session for everyone signed in, with or without a workspace: it is
	// what the browser re-mints its token from (GET sessions), and it hands out the
	// workspace's token once the person creates or joins one.
	c.logDeviceLogin(deviceID, sessionID, dto.Ctx)
	return withDeviceCookie(response.Success(c.sessionPayload(u, res.ClientID, sessionID, map[string]any{
		"user":         maskUser(u),
		"accessToken":  access,
		"refreshToken": refresh,
		"isNew":        res.IsNew,
	}), title, "signed in", nil), deviceID)
}

// ssoAdmin authorizes the caller to manage the token's workspace's SSO: they
// must be an owner or admin of it, unless the host's authorizer decides
// (org.ManageSecurity).
func (c *AuthController) ssoAdmin(ntx ntxctx.NTXContext) bool {
	if ntx.UserID == "" || ntx.WorkspaceID == "" {
		return false
	}
	ok, err := org.Authorize(context.Background(), ntx.WorkspaceID, ntx.UserID, org.ManageSecurity, func() (bool, error) {
		link, _ := c.ucwEntity.First(`"user_id" = ? AND "workspace_id" = ?`, ntx.UserID, ntx.WorkspaceID)
		return link != nil && link.Status != nil && (*link.Status == "owner" || *link.Status == "admin"), nil
	})
	return err == nil && ok
}

// spView is what the console shows an admin to register with an IdP: the
// OpenID redirect URI and the workspace's SAML service provider.
// linkSSOSession records on a device session the SAML session it came from (or clears it: a
// password sign-in on the same device is not the IdP's to end).
func (c *AuthController) linkSSOSession(sessionID string, s *ssoSession) {
	if sessionID == "" || c.db == nil || c.db.DB == nil {
		return
	}
	if s == nil {
		s = &ssoSession{}
	}
	_ = c.db.DB.Exec(`UPDATE "IdentityDeviceSessions" SET "sso_connection_id" = ?, "sso_name_id" = ?, "sso_session_index" = ? WHERE "id" = ?`,
		s.ConnectionID, s.NameID, s.SessionIndex, sessionID).Error
}

func spView(workspaceID string) map[string]any {
	sp := samlSP(workspaceID)
	return map[string]any{"redirectUri": ssoRedirectURI(), "spEntityId": sp.EntityID, "acsUrl": sp.ACSURL, "spMetadataUrl": sp.EntityID}
}

func connectionView(conn *sso.Connection) map[string]any {
	v := spView(conn.WorkspaceID)
	for k, x := range map[string]any{
		"protocol": conn.Protocol, "issuer": conn.Issuer, "clientId": conn.OIDCClientID, "domains": conn.Domains(),
		"enforce": conn.Enforce, "hasSecret": conn.SecretSealed != "", "updatedAt": conn.UpdatedAt,
		"metadataUrl": conn.SAMLMetadataURL,
	} {
		v[k] = x
	}
	return v
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
	ctx := context.Background()
	conn, err := c.ssoStore.ForWorkspace(ctx, dto.Ctx.WorkspaceID)
	if err != nil {
		// Not configured is a normal state; still tell the console the callback
		// URL to register with the provider, and the SAML signing certificate.
		v := spView(dto.Ctx.WorkspaceID)
		for k, x := range c.keyView(ctx, dto.Ctx.WorkspaceID, nil) {
			v[k] = x
		}
		v["configured"] = false
		return response.Success(v, title, "ok", nil)
	}
	v := connectionView(conn)
	for k, x := range c.keyView(ctx, conn.WorkspaceID, conn) {
		v[k] = x
	}
	v["configured"] = true
	return response.Success(v, title, "ok", nil)
}

// keyView is the service provider's key (made now if there is none) and, for a SAML connection,
// whether its IdP offers single logout and wants signed requests.
func (c *AuthController) keyView(ctx context.Context, workspaceID string, conn *sso.Connection) map[string]any {
	sp, keys := c.samlSPFor(ctx, workspaceID, conn, true)
	v := spKeyView(sp, keys)
	v["signRequests"], v["idpSlo"], v["idpWantsSigned"] = false, false, false
	if conn != nil && conn.SAML() {
		v["signRequests"] = conn.SignRequests
		if idp, err := c.ssoClient.ParseSAMLMetadata([]byte(conn.SAMLMetadata)); err == nil {
			v["idpSlo"], v["idpWantsSigned"] = idp.SLOURL != "", idp.WantsSignedRequests
		}
	}
	return v
}

// SsoSpKey is POST /sso/connection/sp-key: rotate the SAML service provider's signing key in two
// steps, so the IdP can trust the next certificate before it is used. "rotate" stages a next key,
// published in the metadata beside the current one; "activate" makes it the one that signs (the
// old one is dropped); "cancel" drops a staged key.
func (c *AuthController) SsoSpKey(dto *SsoSpKeyDto) types.Output {
	title := c.lang.Translate("apps.identity.app.title", nil, dto.Ctx.Preference)
	if !c.ssoReady() {
		return response.NotFound(title, "single sign-on is not available")
	}
	if !c.ssoAdmin(dto.Ctx) {
		return response.Forbidden(title, "only company admins can manage single sign-on")
	}
	ctx := context.Background()
	ws := dto.Ctx.WorkspaceID
	seal := tokenstore.Key()
	if len(seal) != 32 {
		return response.Error(title, "single sign-on is not configured on this server", 503)
	}
	sp, _ := c.samlSPFor(ctx, ws, nil, true)
	switch dto.Action {
	case "rotate":
		kp, err := sso.NewSPKeyPair(sp.EntityID)
		if err != nil {
			return response.InternalServerError(title, "could not make a key")
		}
		sealed, err := tokenstore.SealProviderToken(kp.KeyPEM(), seal)
		if err != nil {
			return response.InternalServerError(title, "could not store the key")
		}
		if err := c.ssoStore.StageSPKey(ctx, ws, sealed, kp.CertB64()); err != nil {
			return response.InternalServerError(title, "could not store the key")
		}
	case "activate":
		if err := c.ssoStore.ActivateSPKey(ctx, ws); errors.Is(err, sso.ErrNotFound) {
			return response.Conflict(title, "there is no new key to activate; rotate first")
		} else if err != nil {
			return response.InternalServerError(title, "could not activate the key")
		}
	case "cancel":
		if err := c.ssoStore.DropNextSPKey(ctx, ws); err != nil {
			return response.InternalServerError(title, "could not drop the key")
		}
	default:
		return response.BadRequest(title, "action must be rotate, activate or cancel")
	}
	conn, _ := c.ssoStore.ForWorkspace(ctx, ws)
	return response.Success(c.keyView(ctx, ws, conn), title, "saved", nil)
}

// SsoPutConnection creates or replaces the workspace's SSO settings. Every
// claimed domain must be one the workspace has verified; the provider must be
// reachable and speak OpenID Connect, or its SAML metadata must be usable, so
// a typo is caught here rather than at the first sign-in.
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

	protocol := strings.ToLower(strings.TrimSpace(dto.Protocol))
	if protocol == "" {
		protocol = sso.ProtocolOIDC
	}
	prev, _ := c.ssoStore.ForWorkspace(ctx, dto.Ctx.WorkspaceID)
	conn := &sso.Connection{WorkspaceID: dto.Ctx.WorkspaceID, Protocol: protocol, Enforce: dto.Enforce}
	switch protocol {
	case sso.ProtocolOIDC:
		issuer := strings.TrimRight(strings.TrimSpace(dto.Issuer), "/")
		conn.Issuer, conn.OIDCClientID = issuer, strings.TrimSpace(dto.OIDCClientID)
		if issuer == "" || conn.OIDCClientID == "" {
			return response.BadRequest(title, "the issuer URL and client ID are required")
		}
		if _, err := c.ssoClient.Discover(ctx, issuer); err != nil {
			return response.BadRequest(title, "the identity provider could not be reached, or is not an OpenID Connect provider")
		}
	case sso.ProtocolSAML:
		var raw []byte
		switch {
		case strings.TrimSpace(dto.MetadataURL) != "":
			conn.SAMLMetadataURL = strings.TrimSpace(dto.MetadataURL)
			b, err := c.ssoClient.FetchSAMLMetadata(ctx, conn.SAMLMetadataURL, true)
			if err != nil {
				return response.BadRequest(title, "the metadata URL could not be read (it must be https and public)")
			}
			raw = b
		case strings.TrimSpace(dto.MetadataXML) != "":
			raw = []byte(dto.MetadataXML)
		case prev != nil && prev.SAML():
			conn.SAMLMetadataURL, raw = prev.SAMLMetadataURL, []byte(prev.SAMLMetadata)
		default:
			return response.BadRequest(title, "give the identity provider's SAML metadata URL or XML")
		}
		idp, err := c.ssoClient.ParseSAMLMetadata(raw)
		if err != nil {
			return response.BadRequest(title, "the SAML metadata is not usable: it needs the identity provider's entity ID, a signing certificate and an HTTP-Redirect sign-on URL")
		}
		conn.Issuer, conn.SAMLMetadata = idp.EntityID, string(raw)
		switch {
		case dto.SignRequests != nil:
			conn.SignRequests = *dto.SignRequests
		case prev != nil && prev.SAML():
			conn.SignRequests = prev.SignRequests
		default:
			conn.SignRequests = idp.WantsSignedRequests
		}
	default:
		return response.BadRequest(title, "protocol must be oidc or saml")
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

	db, _ := json.Marshal(domains)
	conn.DomainsJSON = db
	if protocol == sso.ProtocolOIDC {
		if dto.ClientSecret != "" {
			sealed, err := tokenstore.SealProviderToken(dto.ClientSecret, key)
			if err != nil {
				return response.InternalServerError(title, "could not store the secret")
			}
			conn.SecretSealed = sealed
		} else if prev != nil && !prev.SAML() {
			conn.SecretSealed = prev.SecretSealed
		}
		if conn.SecretSealed == "" {
			return response.BadRequest(title, "a client secret is required")
		}
	}
	if err := c.ssoStore.Save(ctx, conn); err != nil {
		return response.InternalServerError(title, "could not save")
	}
	v := connectionView(conn)
	for k, x := range c.keyView(ctx, conn.WorkspaceID, conn) {
		v[k] = x
	}
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
