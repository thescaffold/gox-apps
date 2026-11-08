package sso_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/beevik/etree"
	"github.com/crewjam/saml"

	"github.com/thescaffold/gox-apps/libs/identity/pkg/sso"
	"github.com/thescaffold/gox-apps/libs/identity/pkg/sso/samltest"
)

type memReplay struct {
	mu   sync.Mutex
	seen map[string]bool
}

func (m *memReplay) Consume(id string, _ time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.seen[id] {
		return saml.ErrAssertionReplayed
	}
	m.seen[id] = true
	return nil
}

type samlWorld struct {
	t      *testing.T
	c      *sso.Client
	idp    *samltest.IdP
	meta   *sso.SAMLIdP
	sp     sso.SAMLServiceProvider
	replay *memReplay
}

func newSAMLWorld(t *testing.T) *samlWorld {
	t.Helper()
	c := sso.New(true)
	idp := samltest.New(t)
	raw, err := c.FetchSAMLMetadata(context.Background(), idp.Server.URL+"/metadata", false)
	if err != nil {
		t.Fatal(err)
	}
	meta, err := c.ParseSAMLMetadata(raw)
	if err != nil {
		t.Fatal(err)
	}
	if meta.EntityID != idp.EntityID || meta.SSOURL != idp.Server.URL+"/sso" {
		t.Fatalf("metadata read as %+v", meta)
	}
	return &samlWorld{t: t, c: c, idp: idp, meta: meta, replay: &memReplay{seen: map[string]bool{}},
		sp: sso.SAMLServiceProvider{EntityID: "https://app.test/sso/saml/ws-1", ACSURL: "https://app.test/sso/saml/ws-1/acs"}}
}

// start makes an AuthnRequest and returns what the IdP reads from it.
func (w *samlWorld) start() samltest.Request {
	w.t.Helper()
	redirect, reqID, err := w.c.SAMLAuthnRequest(w.sp, w.meta, "relay-1")
	if err != nil {
		w.t.Fatal(err)
	}
	if !strings.HasPrefix(redirect, w.meta.SSOURL+"?") {
		w.t.Fatalf("redirect = %s", redirect)
	}
	req, relay, err := samltest.ParseRedirect(redirect)
	if err != nil {
		w.t.Fatal(err)
	}
	if req.ID != reqID || req.ACS != w.sp.ACSURL || req.Issuer != w.sp.EntityID || relay != "relay-1" {
		w.t.Fatalf("request = %+v relay %q, id %q", req, relay, reqID)
	}
	return req
}

func (w *samlWorld) verify(req samltest.Request, o samltest.Options) (*sso.Claims, error) {
	w.t.Helper()
	if o.Email == "" {
		o.Email = "Dana@Acme.test"
	}
	body, err := w.idp.Response(req, o)
	if err != nil {
		w.t.Fatal(err)
	}
	return w.c.SAMLVerify(w.sp, w.meta, body, req.ID, w.replay)
}

func TestSAML_AcceptsASignedAssertionOrResponse(t *testing.T) {
	w := newSAMLWorld(t)
	for _, mode := range []string{samltest.SignAssertion, samltest.SignResponse, samltest.SignBoth} {
		cl, err := w.verify(w.start(), samltest.Options{Sign: mode, Name: "Dana Ito"})
		if err != nil {
			t.Fatalf("%s: %v", mode, err)
		}
		if cl.Email != "dana@acme.test" || !cl.EmailVerified || cl.Name != "Dana Ito" || cl.Subject != "Dana@Acme.test" {
			t.Fatalf("%s: claims %+v", mode, cl)
		}
	}
	// Within the skew: expired 30 s ago still passes.
	if _, err := w.verify(w.start(), samltest.Options{NotOnOrAfter: time.Now().Add(-30 * time.Second)}); err != nil {
		t.Fatalf("inside the skew: %v", err)
	}
	// The NameID is the address when no attribute carries one.
	cl, err := w.verify(w.start(), samltest.Options{Email: "-", NameID: "Erin@acme.test"})
	if err != nil || cl.Email != "erin@acme.test" {
		t.Fatalf("NameID only: %+v %v", cl, err)
	}
}

func TestSAML_RefusesWhatIsNotSignedForThisRequestAndThisServiceProvider(t *testing.T) {
	w := newSAMLWorld(t)
	other, _ := samltest.KeyPair(t)
	setText := func(path, v string) func(*etree.Element) {
		return func(r *etree.Element) {
			el := r.FindElement(path)
			if el == nil {
				t.Fatalf("no %s", path)
			}
			el.SetText(v)
		}
	}
	cases := []struct {
		name string
		o    samltest.Options
		want string // in the reason
	}{
		{"unsigned", samltest.Options{Sign: samltest.SignNone}, "signature element not present"},
		{"signed by another key", samltest.Options{Key: other}, "cannot validate signature on Assertion"},
		{"response signed by another key", samltest.Options{Sign: samltest.SignResponse, Key: other}, "cannot validate signature on Response"},
		{"address changed after signing", samltest.Options{Tamper: setText("//Attribute[@Name='email']/AttributeValue", "mallory@acme.test")}, "cannot validate signature on Assertion"},
		{"subject changed after a response signature", samltest.Options{Sign: samltest.SignResponse, Tamper: setText("//NameID", "mallory@acme.test")}, "cannot validate signature on Response"},
		{"another audience", samltest.Options{Audience: "https://other.example/sp"}, "AudienceRestriction"},
		{"no audience", samltest.Options{Audience: "-"}, "AudienceRestriction"},
		{"another recipient", samltest.Options{Recipient: "https://other.example/acs"}, "Recipient"},
		{"another destination", samltest.Options{Sign: samltest.SignResponse, Destination: "https://other.example/acs"}, "Destination"},
		{"another request", samltest.Options{InResponseTo: "id-not-ours"}, "InResponseTo"},
		{"unsolicited", samltest.Options{InResponseTo: "-"}, "InResponseTo"},
		{"another issuer", samltest.Options{Issuer: "https://evil.example/idp"}, "Issuer"},
		{"expired beyond the skew", samltest.Options{NotOnOrAfter: time.Now().Add(-2 * time.Minute)}, "expired"},
		{"not yet valid", samltest.Options{NotBefore: time.Now().Add(2 * time.Minute)}, "not yet valid"},
		{"issued in the future", samltest.Options{IssueInstant: time.Now().Add(2 * time.Minute)}, "in the future"},
		{"no address", samltest.Options{Email: "-", NameID: "opaque-1234"}, "no email address"},
	}
	for _, c := range cases {
		cl, err := w.verify(w.start(), c.o)
		if err == nil || !errors.Is(err, sso.ErrSAMLResponse) || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: accepted (%+v) or refused for another reason: %v", c.name, cl, err)
		}
	}
	if _, err := w.c.SAMLVerify(w.sp, w.meta, "not base64!", w.start().ID, w.replay); !errors.Is(err, sso.ErrSAMLResponse) {
		t.Errorf("garbage: %v", err)
	}
	if _, err := w.c.SAMLVerify(w.sp, w.meta, "", "", w.replay); !errors.Is(err, sso.ErrSAMLResponse) {
		t.Errorf("no request: %v", err)
	}
	// The IdP answered but did not sign the person in.
	if _, err := w.verify(w.start(), samltest.Options{Status: "urn:oasis:names:tc:SAML:2.0:status:Responder"}); !errors.Is(err, sso.ErrSAMLDenied) {
		t.Errorf("failure status: %v", err)
	}
}

func TestSAML_AnAssertionIsAcceptedOnce(t *testing.T) {
	w := newSAMLWorld(t)
	if _, err := w.verify(w.start(), samltest.Options{AssertionID: "_replayed"}); err != nil {
		t.Fatal(err)
	}
	// The same assertion, even answering a new request, is refused.
	_, err := w.verify(w.start(), samltest.Options{AssertionID: "_replayed"})
	if !errors.Is(err, sso.ErrSAMLResponse) || !strings.Contains(err.Error(), "already been used") {
		t.Fatalf("replay: %v", err)
	}
}

func TestSAML_MetadataMustBeAnIdPWithASigningKeyAndARedirectSignOn(t *testing.T) {
	w := newSAMLWorld(t)
	good := w.idp.Metadata()
	for name, raw := range map[string]string{
		"garbage":         "<nope",
		"empty":           "",
		"no signing cert": strings.Replace(good, `use="signing"`, `use="encryption"`, 1),
		"no redirect":     strings.Replace(good, "HTTP-Redirect", "HTTP-POST", 1),
		"not an IdP":      strings.ReplaceAll(good, "IDPSSODescriptor", "SPSSODescriptor"),
	} {
		if _, err := w.c.ParseSAMLMetadata([]byte(raw)); !errors.Is(err, sso.ErrSAMLMetadata) {
			t.Errorf("%s: %v", name, err)
		}
	}
	// Without AllowPrivate the sign-on URL must be https.
	if _, err := sso.New(false).ParseSAMLMetadata([]byte(good)); !errors.Is(err, sso.ErrSAMLMetadata) {
		t.Errorf("plain-http sign-on accepted: %v", err)
	}
	if _, err := sso.New(false).FetchSAMLMetadata(context.Background(), w.idp.Server.URL+"/metadata", false); err == nil {
		t.Error("fetching metadata from a loopback http URL must be refused")
	}
	md, err := w.c.SAMLMetadata(w.sp)
	if err != nil || !strings.Contains(string(md), `entityID="`+w.sp.EntityID+`"`) || !strings.Contains(string(md), `Location="`+w.sp.ACSURL+`"`) {
		t.Fatalf("SP metadata: %v\n%s", err, md)
	}
}
