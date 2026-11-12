package tests

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/thescaffold/gox-apps/libs/identity/app/auth"
	"github.com/thescaffold/gox-apps/libs/identity/pkg/revocation"
	"github.com/thescaffold/gox-apps/libs/identity/pkg/sso"
	"github.com/thescaffold/gox-apps/libs/identity/pkg/sso/samltest"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
)

// V4-03: a SAML connection's service provider holds its own signing key (made on first use,
// rotated in two steps), signs AuthnRequests when the connection asks, and does single logout
// both ways. IdP-initiated sign-in stays refused (D-27); encrypted assertions too (D-15).

var noFollow = &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}

// spCert reads the service provider's current certificate from the console's view.
func (w *ssoWorld) spKey(t *testing.T) (fingerprint string, view map[string]any) {
	t.Helper()
	rep := decode(t, w.r.ctl.SsoGetConnection(&auth.SsoWorkspaceDto{Ctx: w.ctx(w.admin)}))
	if rep.code != 200 {
		t.Fatalf("get connection: %d %v", rep.code, rep.body)
	}
	fp, _ := rep.data()["spFingerprint"].(string)
	return fp, rep.data()
}

func certOf(t *testing.T, view map[string]any, field string) *sso.SPKeyPair {
	t.Helper()
	pem, _ := view[field].(string)
	if pem == "" {
		t.Fatalf("no %s in %v", field, view)
	}
	kp, err := sso.ParseSPCert(pem)
	if err != nil {
		t.Fatal(err)
	}
	return kp
}

func TestSSO_SAML_SignedRequestsWithARotatableKey(t *testing.T) {
	w := ssoRig(t)
	idp := samltest.New(t)
	sp := "https://app.test/api/apps/identity/sso/saml/" + w.ws

	// The key exists before the connection does: the admin registers its certificate first.
	fp, view := w.spKey(t)
	if fp == "" || view["spSloUrl"] != sp+"/slo" || view["signRequests"] != false {
		t.Fatalf("before: %v", view)
	}
	cur := certOf(t, view, "spCert")
	if again, _ := w.spKey(t); again != fp {
		t.Fatal("the key changed between two reads")
	}
	md := string(w.r.ctl.SsoSamlMetadata(&auth.SsoSamlMetadataDto{Workspace: w.ws}).Data().([]byte))
	if !strings.Contains(md, cur.CertB64()) || !strings.Contains(md, `use="signing"`) || strings.Contains(md, `use="encryption"`) || !strings.Contains(md, sp+"/slo") {
		t.Fatalf("SP metadata lacks the signing key or SLO, or offers encryption:\n%s", md)
	}

	// Signing is off until the admin turns it on (or the IdP's metadata asks for it).
	w.connectSAML(idp, false)
	if _, view := w.spKey(t); view["signRequests"] != false || view["idpSlo"] != true {
		t.Fatalf("after connect: %v", view)
	}
	start := func() *url.URL {
		t.Helper()
		rep := decode(t, w.r.ctl.SsoStart(&auth.SsoStartDto{Email: "x@" + w.domain, ClientId: w.r.client}))
		u, _ := url.Parse(rep.data()["url"].(string))
		return u
	}
	if start().Query().Get("Signature") != "" {
		t.Fatal("signed with signing off")
	}
	on := true
	rep := decode(t, w.r.ctl.SsoPutConnection(&auth.SsoConnectionDto{Ctx: w.ctx(w.admin), Protocol: "saml", Domains: []string{w.domain}, SignRequests: &on}))
	if rep.code != 200 || rep.data()["signRequests"] != true {
		t.Fatalf("turn signing on: %d %v", rep.code, rep.body)
	}
	u := start()
	if err := samltest.VerifyRedirect(u.RawQuery, cur.Cert); err != nil {
		t.Fatalf("the request is not signed by the published key: %v", err)
	}
	// An update that does not mention signing keeps it.
	decode(t, w.r.ctl.SsoPutConnection(&auth.SsoConnectionDto{Ctx: w.ctx(w.admin), Protocol: "saml", Domains: []string{w.domain}}))
	if _, view := w.spKey(t); view["signRequests"] != true {
		t.Fatal("an update dropped signRequests")
	}

	// Rotation, step one: a next key is published beside the current one, which still signs.
	if rep := decode(t, w.r.ctl.SsoSpKey(&auth.SsoSpKeyDto{Ctx: w.ctx(w.admin), Action: "rotate"})); rep.code != 200 || rep.data()["spNextFingerprint"] == "" {
		t.Fatalf("rotate: %d %v", rep.code, rep.body)
	}
	_, view = w.spKey(t)
	next := certOf(t, view, "spNextCert")
	md = string(w.r.ctl.SsoSamlMetadata(&auth.SsoSamlMetadataDto{Workspace: w.ws}).Data().([]byte))
	if !strings.Contains(md, cur.CertB64()) || !strings.Contains(md, next.CertB64()) {
		t.Fatal("the metadata does not publish both keys during a rotation")
	}
	if err := samltest.VerifyRedirect(start().RawQuery, cur.Cert); err != nil {
		t.Fatalf("a staged key signs before it is activated: %v", err)
	}
	// Step two: the next key signs, and the old one is gone from the metadata.
	if rep := decode(t, w.r.ctl.SsoSpKey(&auth.SsoSpKeyDto{Ctx: w.ctx(w.admin), Action: "activate"})); rep.code != 200 || rep.data()["spFingerprint"] != next.Fingerprint() {
		t.Fatalf("activate: %d %v", rep.code, rep.body)
	}
	if err := samltest.VerifyRedirect(start().RawQuery, next.Cert); err != nil {
		t.Fatalf("the activated key does not sign: %v", err)
	}
	md = string(w.r.ctl.SsoSamlMetadata(&auth.SsoSamlMetadataDto{Workspace: w.ws}).Data().([]byte))
	if strings.Contains(md, cur.CertB64()) {
		t.Fatal("the retired key is still published")
	}
	if rep := decode(t, w.r.ctl.SsoSpKey(&auth.SsoSpKeyDto{Ctx: w.ctx(w.admin), Action: "activate"})); rep.code != 409 {
		t.Fatalf("activate with nothing staged: %d", rep.code)
	}
	// Only the company's admins rotate it.
	member, _ := w.r.person("member")
	if rep := decode(t, w.r.ctl.SsoSpKey(&auth.SsoSpKeyDto{Ctx: w.ctx(member.Id), Action: "rotate"})); rep.code != 403 {
		t.Fatalf("a non-admin rotated the key: %d", rep.code)
	}
}

// samlSession signs email in through the IdP on a new browser and returns its token, its device
// and the SessionIndex the IdP gave it.
func (w *ssoWorld) samlSession(t *testing.T, idp *samltest.IdP, email, sessionIndex string) (token, device string) {
	t.Helper()
	loc := w.samlSignIn(idp, email, samltest.Options{SessionIndex: sessionIndex})
	if !strings.HasPrefix(loc, completeAt) {
		t.Fatalf("ACS went to %q", loc)
	}
	ex := decode(t, w.r.ctl.SsoExchange(&auth.SsoExchangeDto{Code: strings.TrimPrefix(loc, completeAt), UserAgent: chromeMac}))
	if ex.code != 200 {
		t.Fatalf("exchange: %d %v", ex.code, ex.body)
	}
	return ex.data()["accessToken"].(string), deviceFrom(t, ex)
}

func (w *ssoWorld) slo(t *testing.T, rawQuery string) (int, string) {
	t.Helper()
	out := w.r.ctl.SsoSamlSlo(&auth.SsoSamlSloDto{Workspace: w.ws, Query: rawQuery})
	return out.Code(), out.Headers()["Location"]
}

func refused(token string) bool {
	claims, _ := coreauth.Verify(token, "changeme")
	return revocation.Check(claims)
}

func TestSSO_SAML_SingleLogoutStartedHere(t *testing.T) {
	w := ssoRig(t)
	revocation.SetStore(&revocation.Store{DB: w.r.db})
	t.Cleanup(func() { revocation.SetStore(nil) })
	idp := samltest.New(t)
	sp := "https://app.test/api/apps/identity/sso/saml/" + w.ws
	_, view := w.spKey(t)
	idp.TrustSP(sp+"/slo", certOf(t, view, "spCert").Cert)
	w.connectSAML(idp, false)
	email := "sol@" + w.domain
	w.dir.invited[w.ws+"|"+email] = true
	token, device := w.samlSession(t, idp, email, "_idx-here")

	// Signing out of this browser ends the session here and sends it on to the IdP.
	rep := decode(t, w.r.ctl.Logout(&auth.LogoutDto{Device: "yes", Cookie: cookieFor(device), Ctx: as(t, token)}))
	if rep.code != 200 {
		t.Fatalf("logout: %d %v", rep.code, rep.body)
	}
	if !refused(token) {
		t.Fatal("the session's token is still accepted after sign-out")
	}
	logoutURL, _ := rep.data()["logoutUrl"].(string)
	if !strings.HasPrefix(logoutURL, idp.Server.URL+"/slo?") {
		t.Fatalf("logoutUrl = %q", logoutURL)
	}
	res, err := noFollow.Get(logoutURL)
	if err != nil || res.StatusCode != http.StatusFound {
		t.Fatalf("the IdP refused the LogoutRequest: %v %v", err, res.Status)
	}
	lo := idp.Logouts()
	if len(lo) != 1 || lo[0].NameID != email || lo[0].SessionIndex != "_idx-here" {
		t.Fatalf("the IdP received %+v", lo)
	}
	back, _ := url.Parse(res.Header.Get("Location"))
	if back.Scheme+"://"+back.Host+back.Path != sp+"/slo" {
		t.Fatalf("the IdP sent the browser to %s", back)
	}
	code, loc := w.slo(t, back.RawQuery)
	if code != http.StatusFound || loc != "https://app.test/guard/login?signedout=sso" {
		t.Fatalf("the LogoutResponse: %d %q", code, loc)
	}
	// The same answer again is not taken (its request is answered).
	if _, loc := w.slo(t, back.RawQuery); !strings.Contains(loc, "error=") {
		t.Fatalf("a replayed LogoutResponse: %q", loc)
	}

	// A password sign-in's logout has no IdP to tell.
	ada, _ := w.r.person("ada")
	in := w.r.loginFrom(ada, chromeMac, "203.0.113.9")
	rep = decode(t, w.r.ctl.Logout(&auth.LogoutDto{Device: "yes", Cookie: cookieFor(deviceFrom(t, in)), Ctx: as(t, in.data()["accessToken"].(string))}))
	if rep.code != 200 || rep.data()["logoutUrl"] != nil {
		t.Fatalf("password logout: %d %v", rep.code, rep.body)
	}
}

func TestSSO_SAML_SingleLogoutStartedAtTheIdP(t *testing.T) {
	w := ssoRig(t)
	revocation.SetStore(&revocation.Store{DB: w.r.db})
	t.Cleanup(func() { revocation.SetStore(nil) })
	idp := samltest.New(t)
	sp := "https://app.test/api/apps/identity/sso/saml/" + w.ws
	_, view := w.spKey(t)
	spCert := certOf(t, view, "spCert").Cert
	idp.TrustSP(sp+"/slo", spCert)
	w.connectSAML(idp, false)
	email := "ida@" + w.domain
	w.dir.invited[w.ws+"|"+email] = true
	laptop, _ := w.samlSession(t, idp, email, "_idx-laptop")
	phone, _ := w.samlSession(t, idp, email, "_idx-phone")
	request := func(o samltest.LogoutOptions, nameID, index string) string {
		u, _ := url.Parse(idp.LogoutRequestURL(sp+"/slo", nameID, index, "idp-relay", o))
		return u.RawQuery
	}

	// Refused, and nothing ends: unsigned, another key, another person's.
	other, _ := samltest.KeyPair(t)
	for name, q := range map[string]string{
		"unsigned":    request(samltest.LogoutOptions{Unsigned: true}, email, "_idx-laptop"),
		"another key": request(samltest.LogoutOptions{Key: other}, email, "_idx-laptop"),
	} {
		if code, _ := w.slo(t, q); code != http.StatusBadRequest {
			t.Errorf("%s: %d", name, code)
		}
	}
	if refused(laptop) || refused(phone) {
		t.Fatal("a refused LogoutRequest ended a session")
	}

	// The laptop's session (by its SessionIndex) ends; the phone's stays.
	code, loc := w.slo(t, request(samltest.LogoutOptions{ID: "_lr-laptop"}, email, "_idx-laptop"))
	if code != http.StatusFound || !strings.HasPrefix(loc, idp.Server.URL+"/slo?") {
		t.Fatalf("IdP-initiated logout: %d %q", code, loc)
	}
	if !refused(laptop) || refused(phone) {
		t.Fatalf("after logout: laptop refused %v, phone refused %v", refused(laptop), refused(phone))
	}
	// The answer is signed by this side, answers the request and says Success.
	if res, err := http.Get(loc); err != nil || res.StatusCode != 200 {
		t.Fatalf("the IdP refused the LogoutResponse: %v %v", err, res)
	}
	lo := idp.Logouts()
	if len(lo) != 1 || lo[0].InResponseTo != "_lr-laptop" || !strings.HasSuffix(lo[0].Status, ":Success") || lo[0].RelayState != "idp-relay" {
		t.Fatalf("the IdP received %+v", lo)
	}
	// The same request again is a replay.
	if code, _ := w.slo(t, request(samltest.LogoutOptions{ID: "_lr-laptop"}, email, "_idx-laptop")); code != http.StatusBadRequest {
		t.Fatalf("replayed LogoutRequest: %d", code)
	}
	// With no SessionIndex, every session of that person through this connection ends.
	w.slo(t, request(samltest.LogoutOptions{}, email, ""))
	if !refused(phone) {
		t.Fatal("the phone's session survived a logout of every session")
	}
}
