package tests

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/beevik/etree"
	"github.com/google/uuid"
	"github.com/thescaffold/gox-apps/libs/identity/app/auth"
	"github.com/thescaffold/gox-apps/libs/identity/pkg/sso/samltest"
)

// SAML 2.0 connections (U-S21): the same flow as OpenID Connect — start,
// the IdP, back to identity, exchange — against samltest's signing IdP.

func (w *ssoWorld) connectSAML(idp *samltest.IdP, enforce bool) reply {
	w.r.t.Helper()
	return decode(w.r.t, w.r.ctl.SsoPutConnection(&auth.SsoConnectionDto{
		Ctx: w.ctx(w.admin), Protocol: "saml", MetadataURL: idp.Server.URL + "/metadata",
		Domains: []string{w.domain}, Enforce: enforce,
	}))
}

// samlStart begins a sign-in and returns what the IdP receives.
func (w *ssoWorld) samlStart(email string) (samltest.Request, string) {
	w.r.t.Helper()
	start := decode(w.r.t, w.r.ctl.SsoStart(&auth.SsoStartDto{Email: email, ClientId: w.r.client, ReturnTo: "/sso/complete"}))
	if start.code != 200 {
		w.r.t.Fatalf("start: %d %v", start.code, start.body)
	}
	req, relay, err := samltest.ParseRedirect(start.data()["url"].(string))
	if err != nil || relay == "" {
		w.r.t.Fatalf("start url %v: %v", start.data()["url"], err)
	}
	return req, relay
}

// samlPost is the browser posting the IdP's form to the ACS.
func (w *ssoWorld) samlPost(workspace, body, relay string) string {
	w.r.t.Helper()
	return w.r.ctl.SsoSamlAcs(&auth.SsoSamlAcsDto{Workspace: workspace, SAMLResponse: body, RelayState: relay}).Headers()["Location"]
}

func (w *ssoWorld) samlSignIn(idp *samltest.IdP, email string, o samltest.Options) string {
	w.r.t.Helper()
	req, relay := w.samlStart(email)
	if o.Email == "" {
		o.Email = email
	}
	body, err := idp.Response(req, o)
	if err != nil {
		w.r.t.Fatal(err)
	}
	return w.samlPost(w.ws, body, relay)
}

const completeAt = "https://app.test/sso/complete?code="

func TestSSO_SAML_InvitedPersonSignsInAndJoins(t *testing.T) {
	w := ssoRig(t)
	idp := samltest.New(t)
	sp := "https://app.test/api/apps/identity/sso/saml/" + w.ws

	// Before the connection exists the admin already sees what to register.
	before := decode(t, w.r.ctl.SsoGetConnection(&auth.SsoWorkspaceDto{Ctx: w.ctx(w.admin)}))
	if before.data()["configured"] != false || before.data()["spEntityId"] != sp || before.data()["acsUrl"] != sp+"/acs" {
		t.Fatalf("before: %v", before.body)
	}
	rep := w.connectSAML(idp, false)
	if rep.code != 200 || rep.data()["protocol"] != "saml" || rep.data()["issuer"] != idp.EntityID || rep.data()["metadataUrl"] != idp.Server.URL+"/metadata" || rep.data()["hasSecret"] != false {
		t.Fatalf("connect: %d %v", rep.code, rep.body)
	}
	if d := decode(t, w.r.ctl.SsoDiscover(&auth.SsoDiscoverDto{Email: "x@" + w.domain})); d.data()["protocol"] != "saml" || d.data()["sso"] != true {
		t.Fatalf("discover: %v", d.body)
	}

	email := "dana@" + w.domain
	w.dir.invited[w.ws+"|"+email] = true
	req, relay := w.samlStart(email)
	if req.ACS != sp+"/acs" || req.Issuer != sp {
		t.Fatalf("AuthnRequest: %+v", req)
	}
	body, _ := idp.Response(req, samltest.Options{Email: "Dana@" + strings.ToUpper(w.domain), Name: "Dana Saml", AssertionID: "_a-" + uuid.NewString()})
	loc := w.samlPost(w.ws, body, relay)
	if !strings.HasPrefix(loc, completeAt) {
		t.Fatalf("ACS went to %q", loc)
	}
	ex := decode(t, w.r.ctl.SsoExchange(&auth.SsoExchangeDto{Code: strings.TrimPrefix(loc, completeAt)}))
	if ex.code != 200 || ex.data()["accessToken"] == "" || ex.data()["isNew"] != true {
		t.Fatalf("exchange: %d %v", ex.code, ex.body)
	}
	u := ex.data()["user"].(map[string]any)
	if u["email"] != email || len(w.dir.joined) != 1 || !strings.HasSuffix(w.dir.joined[0], "|"+email) {
		t.Fatalf("user %v, joined %v", u, w.dir.joined)
	}

	// The same POST again: the state is gone.
	if again := w.samlPost(w.ws, body, relay); !strings.Contains(again, "error=expired") {
		t.Fatalf("replayed post: %q", again)
	}
	// The same assertion inside an answer to a new request: refused by its id.
	req2, relay2 := w.samlStart(email)
	reused, _ := idp.Response(req2, samltest.Options{Email: email, AssertionID: assertionID(t, body)})
	if l := w.samlPost(w.ws, reused, relay2); !strings.Contains(l, "error=invalid_token") {
		t.Fatalf("replayed assertion: %q", l)
	}
	// A new sign-in (the IdP signing the whole response this time) is the same account.
	loc = w.samlSignIn(idp, email, samltest.Options{Sign: samltest.SignResponse})
	second := decode(t, w.r.ctl.SsoExchange(&auth.SsoExchangeDto{Code: strings.TrimPrefix(loc, completeAt)}))
	if second.data()["isNew"] != false || second.data()["user"].(map[string]any)["id"] != u["id"] {
		t.Fatalf("second sign-in: %q %v", loc, second.body)
	}
}

func assertionID(t *testing.T, b64 string) string {
	t.Helper()
	doc := etree.NewDocument()
	raw, err := base64.StdEncoding.DecodeString(b64)
	if err != nil || doc.ReadFromBytes(raw) != nil {
		t.Fatalf("response: %v", err)
	}
	return doc.FindElement("//Assertion").SelectAttrValue("ID", "")
}

func TestSSO_SAML_RefusesWhatIsNotAnInvitedSignedAnswerToThisCompanysRequest(t *testing.T) {
	w := ssoRig(t)
	idp := samltest.New(t)
	w.connectSAML(idp, false)
	email := "eve@" + w.domain
	other, _ := samltest.KeyPair(t)

	if l := w.samlSignIn(idp, email, samltest.Options{}); !strings.Contains(l, "error=not_invited") {
		t.Fatalf("not invited: %q", l)
	}
	w.dir.invited[w.ws+"|"+email] = true
	for name, c := range map[string]struct {
		o    samltest.Options
		want string
	}{
		"another company's address": {samltest.Options{Email: "eve@other.example"}, "wrong_domain"},
		"signed by another key":     {samltest.Options{Key: other}, "invalid_token"},
		"unsigned":                  {samltest.Options{Sign: samltest.SignNone}, "invalid_token"},
		"for another service":       {samltest.Options{Audience: "https://elsewhere.example/sp"}, "invalid_token"},
		"the IdP said no":           {samltest.Options{Status: "urn:oasis:names:tc:SAML:2.0:status:AuthnFailed"}, "denied"},
	} {
		if l := w.samlSignIn(idp, email, c.o); !strings.Contains(l, "error="+c.want) {
			t.Errorf("%s: redirected to %q, want error=%s", name, l, c.want)
		}
	}
	// Unsolicited (IdP-initiated) responses are not taken.
	req, _ := w.samlStart(email)
	body, _ := idp.Response(req, samltest.Options{Email: email})
	if l := w.samlPost(w.ws, body, ""); !strings.Contains(l, "error=invalid_request") {
		t.Errorf("no RelayState: %q", l)
	}
	// A sign-in started for this company cannot finish at another workspace's ACS.
	req, relay := w.samlStart(email)
	body, _ = idp.Response(req, samltest.Options{Email: email})
	if l := w.samlPost(uuid.NewString(), body, relay); !strings.Contains(l, "error=invalid_request") {
		t.Errorf("another workspace's ACS: %q", l)
	}
	var n int64
	w.r.db.Raw(`SELECT COUNT(*) FROM "IdentityUsers" WHERE email = ?`, email).Scan(&n)
	if n != 0 || len(w.dir.joined) != 0 {
		t.Fatalf("nothing may be created for a refused sign-in: users=%d joined=%v", n, w.dir.joined)
	}
}

func TestSSO_SAML_Configuration(t *testing.T) {
	w := ssoRig(t)
	idp := samltest.New(t)
	put := func(dto auth.SsoConnectionDto) reply {
		dto.Ctx = w.ctx(w.admin)
		if dto.Domains == nil {
			dto.Domains = []string{w.domain}
		}
		return decode(t, w.r.ctl.SsoPutConnection(&dto))
	}
	for name, dto := range map[string]auth.SsoConnectionDto{
		"no metadata":      {Protocol: "saml"},
		"bad metadata":     {Protocol: "saml", MetadataXML: "<EntityDescriptor/>"},
		"unreachable URL":  {Protocol: "saml", MetadataURL: "http://127.0.0.1:1/metadata"},
		"unknown protocol": {Protocol: "ws-fed", MetadataXML: idp.Metadata()},
		"unowned domain":   {Protocol: "saml", MetadataXML: idp.Metadata(), Domains: []string{"gmail.com"}},
	} {
		if rep := put(dto); rep.code != 400 {
			t.Errorf("%s: %d %v", name, rep.code, rep.body)
		}
	}
	// Pasted metadata; then an update that sends none keeps it.
	if rep := put(auth.SsoConnectionDto{Protocol: "saml", MetadataXML: idp.Metadata()}); rep.code != 200 || rep.data()["metadataUrl"] != "" || rep.data()["issuer"] != idp.EntityID {
		t.Fatalf("pasted: %d %v", rep.code, rep.body)
	}
	if rep := put(auth.SsoConnectionDto{Protocol: "saml", Enforce: true}); rep.code != 200 || rep.data()["enforce"] != true {
		t.Fatalf("update: %d %v", rep.code, rep.body)
	}
	// Enforcement is the same for either protocol.
	email := "gus-" + uuid.NewString()[:6] + "@" + w.domain
	personAt(w.r, email)
	if rep := decode(t, w.r.ctl.Login(&auth.LoginDto{Ref: email, Secret: testPassword, ClientId: w.r.client})); rep.code != 403 {
		t.Fatalf("password under SAML enforcement: %d", rep.code)
	}
	// Back to OpenID Connect: the SAML connection had no secret to keep.
	if rep := put(auth.SsoConnectionDto{Protocol: "oidc", Issuer: w.idp.srv.URL, OIDCClientID: "oidc-client"}); rep.code != 400 {
		t.Fatalf("oidc without a secret: %d", rep.code)
	}

	// The service provider's metadata, for the IdP's admin.
	out := w.r.ctl.SsoSamlMetadata(&auth.SsoSamlMetadataDto{Workspace: w.ws})
	md, _ := out.Data().([]byte)
	sp := "https://app.test/api/apps/identity/sso/saml/" + w.ws
	if out.Code() != 200 || out.ContentType() != "application/samlmetadata+xml" || !strings.Contains(string(md), `entityID="`+sp+`"`) || !strings.Contains(string(md), `Location="`+sp+`/acs"`) {
		t.Fatalf("SP metadata: %d %s %s", out.Code(), out.ContentType(), md)
	}
	if out := w.r.ctl.SsoSamlMetadata(&auth.SsoSamlMetadataDto{Workspace: uuid.NewString()}); out.Code() != 404 {
		t.Fatalf("an unknown workspace: %d", out.Code())
	}
}
