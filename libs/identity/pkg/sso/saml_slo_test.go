package sso_test

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/beevik/etree"

	"github.com/thescaffold/gox-apps/libs/identity/pkg/sso"
	"github.com/thescaffold/gox-apps/libs/identity/pkg/sso/samltest"
)

// V4-03: signed AuthnRequests with the service provider's own key (per connection, rotatable),
// and SAML single logout, started by either side, against samltest's IdP. Encrypted assertions
// stay refused (D-15) even now that the SP holds a key; IdP-initiated sign-in stays refused (D-27).

// keyed gives the world's service provider a signing key and an SLO URL, and tells the IdP.
func (w *samlWorld) keyed(sign bool) *sso.SPKeyPair {
	w.t.Helper()
	kp, err := sso.NewSPKeyPair(w.sp.EntityID)
	if err != nil {
		w.t.Fatal(err)
	}
	w.sp.SLOURL = w.sp.EntityID + "/slo"
	w.sp.Key, w.sp.Cert, w.sp.SignRequests = kp.Key, kp.Cert, sign
	w.idp.TrustSP(w.sp.SLOURL, kp.Cert)
	return kp
}

func TestSAML_SignedAuthnRequestTheIdPCanVerify(t *testing.T) {
	w := newSAMLWorld(t)
	// Unsigned unless asked: no key, or a key with signing off.
	redirect, _, err := w.c.SAMLAuthnRequest(w.sp, w.meta, "r")
	if err != nil || strings.Contains(redirect, "Signature=") {
		t.Fatalf("unsigned request: %v %s", err, redirect)
	}
	w.keyed(false)
	if redirect, _, _ := w.c.SAMLAuthnRequest(w.sp, w.meta, "r"); strings.Contains(redirect, "Signature=") {
		t.Fatal("signed with signing off")
	}

	w.sp.SignRequests = true
	redirect, _, err = w.c.SAMLAuthnRequest(w.sp, w.meta, "relay-1")
	if err != nil || !strings.Contains(redirect, "SigAlg=") || !strings.Contains(redirect, "Signature=") {
		t.Fatalf("signed request: %v %s", err, redirect)
	}
	u, _ := url.Parse(redirect)
	if err := samltest.VerifyRedirect(u.RawQuery, w.sp.Cert); err != nil {
		t.Fatalf("the IdP cannot verify it: %v", err)
	}
	other, _ := sso.NewSPKeyPair("https://other.example")
	if err := samltest.VerifyRedirect(u.RawQuery, other.Cert); err == nil {
		t.Fatal("another key verifies it")
	}
	// An IdP that wants signed requests takes this one through its own sign-on page.
	w.idp.WantSigned = true
	res, err := http.Get(redirect)
	if err != nil || res.StatusCode != 200 {
		t.Fatalf("the IdP refused the signed request: %v %v", err, res.Status)
	}
	reqs := w.idp.Requests()
	if !reqs[len(reqs)-1].Signed {
		t.Fatal("the IdP did not see a signature")
	}
	// The IdP's metadata says it wants them, and the parsed IdP says so.
	raw, _ := w.c.FetchSAMLMetadata(context.Background(), w.idp.Server.URL+"/metadata", true)
	meta, _ := w.c.ParseSAMLMetadata(raw)
	if !meta.WantsSignedRequests || meta.SLOURL != w.idp.Server.URL+"/slo" {
		t.Fatalf("parsed IdP: %+v", meta)
	}
}

func TestSAML_SPMetadataPublishesItsSigningKeysAndSLONeverAnEncryptionKey(t *testing.T) {
	w := newSAMLWorld(t)
	kp := w.keyed(true)
	next, _ := sso.NewSPKeyPair(w.sp.EntityID)
	w.sp.NextCert = next.Cert
	md, err := w.c.SAMLMetadata(w.sp)
	if err != nil {
		t.Fatal(err)
	}
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(md); err != nil {
		t.Fatal(err)
	}
	var signing []string
	for _, kd := range doc.FindElements("//KeyDescriptor") {
		if kd.SelectAttrValue("use", "") != "signing" {
			t.Fatalf("a KeyDescriptor of use %q is published (D-15: no encryption key)", kd.SelectAttrValue("use", ""))
		}
		signing = append(signing, strings.Join(strings.Fields(kd.FindElement(".//X509Certificate").Text()), ""))
	}
	want := []string{base64.StdEncoding.EncodeToString(kp.Cert.Raw), base64.StdEncoding.EncodeToString(next.Cert.Raw)}
	if len(signing) != 2 || signing[0] != want[0] || signing[1] != want[1] {
		t.Fatalf("signing keys %d, want the current and the next", len(signing))
	}
	sp := doc.FindElement("//SPSSODescriptor")
	if sp.SelectAttrValue("AuthnRequestsSigned", "") != "true" {
		t.Fatalf("AuthnRequestsSigned = %q", sp.SelectAttrValue("AuthnRequestsSigned", ""))
	}
	slo := doc.FindElement("//SingleLogoutService")
	if slo == nil || slo.SelectAttrValue("Location", "") != w.sp.SLOURL || !strings.HasSuffix(slo.SelectAttrValue("Binding", ""), "HTTP-Redirect") {
		t.Fatalf("SingleLogoutService: %v", slo)
	}
}

func TestSAML_KeyPairRoundTripsThroughPEM(t *testing.T) {
	kp, err := sso.NewSPKeyPair("https://app.test/sp")
	if err != nil {
		t.Fatal(err)
	}
	back, err := sso.ParseSPKeyPair(kp.KeyPEM(), kp.CertB64())
	if err != nil || !back.Cert.Equal(kp.Cert) || back.Key.N.Cmp(kp.Key.N) != 0 {
		t.Fatalf("round trip: %v", err)
	}
	if kp.Fingerprint() == "" || len(kp.Fingerprint()) != 95 {
		t.Fatalf("fingerprint %q", kp.Fingerprint())
	}
}

// D-15 holds with a key held: an encrypted assertion is refused, never decrypted.
func TestSAML_EncryptedAssertionStillRefused(t *testing.T) {
	w := newSAMLWorld(t)
	w.keyed(true)
	req := w.start()
	_, err := w.verify(req, samltest.Options{Sign: samltest.SignResponse, Tamper: func(r *etree.Element) {
		a := r.FindElement("//Assertion")
		r.RemoveChild(a)
		r.CreateElement("saml:EncryptedAssertion").CreateElement("xenc:EncryptedData")
	}})
	if !errors.Is(err, sso.ErrSAMLResponse) || !strings.Contains(err.Error(), "encrypted") {
		t.Fatalf("encrypted assertion: %v", err)
	}
}

// The assertion's SessionIndex and NameID come back with the claims: single logout names the
// session by them.
func TestSAML_ClaimsCarryTheSessionIndex(t *testing.T) {
	w := newSAMLWorld(t)
	cl, err := w.verify(w.start(), samltest.Options{SessionIndex: "_s-42"})
	if err != nil || cl.SessionIndex != "_s-42" || cl.Subject != "Dana@Acme.test" {
		t.Fatalf("claims %+v %v", cl, err)
	}
}

func TestSAML_SPInitiatedLogout(t *testing.T) {
	w := newSAMLWorld(t)
	w.keyed(false)

	redirect, reqID, err := w.c.SAMLLogoutRequest(w.sp, w.meta, "dana@acme.test", "_s-1", "relay-9")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(redirect, w.meta.SLOURL+"?") || !strings.Contains(redirect, "Signature=") {
		t.Fatalf("logout redirect %s", redirect)
	}
	// The IdP verifies it, ends its session and redirects back with a signed LogoutResponse.
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	res, err := client.Get(redirect)
	if err != nil || res.StatusCode != http.StatusFound {
		t.Fatalf("IdP: %v %v", err, res.Status)
	}
	lo := w.idp.Logouts()
	if len(lo) != 1 || lo[0].Kind != "request" || lo[0].ID != reqID || lo[0].NameID != "dana@acme.test" || lo[0].SessionIndex != "_s-1" || lo[0].Issuer != w.sp.EntityID {
		t.Fatalf("the IdP received %+v", lo)
	}
	back, _ := url.Parse(res.Header.Get("Location"))
	if back.Scheme+"://"+back.Host+back.Path != w.sp.SLOURL || back.Query().Get("RelayState") != "relay-9" {
		t.Fatalf("back to %s", back)
	}
	if err := w.c.SAMLLogoutResponseVerify(w.sp, w.meta, back.RawQuery, []string{reqID}); err != nil {
		t.Fatalf("the IdP's answer: %v", err)
	}
	// It answers that request only, it must be signed by the IdP, and say Success.
	if err := w.c.SAMLLogoutResponseVerify(w.sp, w.meta, back.RawQuery, []string{"id-other"}); err == nil {
		t.Fatal("an answer to another request accepted")
	}
	other, _ := samltest.KeyPair(t)
	for name, o := range map[string]samltest.LogoutOptions{
		"unsigned":        {Unsigned: true},
		"another key":     {Key: other},
		"another issuer":  {Issuer: "https://evil.example/idp"},
		"a failure":       {Status: "urn:oasis:names:tc:SAML:2.0:status:Responder"},
		"changed message": {Tamper: func(q url.Values) { q.Set("RelayState", "elsewhere") }},
	} {
		u, _ := url.Parse(w.idp.LogoutResponseURL(w.sp.SLOURL, reqID, "relay-9", o))
		if err := w.c.SAMLLogoutResponseVerify(w.sp, w.meta, u.RawQuery, []string{reqID}); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
	// Without an SLO URL in the IdP's metadata there is nothing to send.
	noSLO := *w.meta
	noSLO.SLOURL = ""
	if _, _, err := w.c.SAMLLogoutRequest(w.sp, &noSLO, "dana@acme.test", "", ""); !errors.Is(err, sso.ErrSAMLNoLogout) {
		t.Fatalf("no IdP SLO: %v", err)
	}
	// Nor without a key: the profile requires logout messages to be signed.
	unkeyed := w.sp
	unkeyed.Key, unkeyed.Cert = nil, nil
	if _, _, err := w.c.SAMLLogoutRequest(unkeyed, w.meta, "dana@acme.test", "", ""); !errors.Is(err, sso.ErrSAMLNoLogout) {
		t.Fatalf("no SP key: %v", err)
	}
}

func TestSAML_IdPInitiatedLogout(t *testing.T) {
	w := newSAMLWorld(t)
	w.keyed(false)
	query := func(o samltest.LogoutOptions, dest string) string {
		u, _ := url.Parse(w.idp.LogoutRequestURL(dest, "dana@acme.test", "_s-7", "rs", o))
		return u.RawQuery
	}

	lr, err := w.c.SAMLLogoutRequestVerify(w.sp, w.meta, query(samltest.LogoutOptions{ID: "_lr-1"}, w.sp.SLOURL), w.replay)
	if err != nil {
		t.Fatalf("a signed LogoutRequest: %v", err)
	}
	if lr.ID != "_lr-1" || lr.NameID != "dana@acme.test" || len(lr.SessionIndexes) != 1 || lr.SessionIndexes[0] != "_s-7" || lr.RelayState != "rs" {
		t.Fatalf("read as %+v", lr)
	}
	// Once only.
	if _, err := w.c.SAMLLogoutRequestVerify(w.sp, w.meta, query(samltest.LogoutOptions{ID: "_lr-1"}, w.sp.SLOURL), w.replay); err == nil {
		t.Fatal("a replayed LogoutRequest accepted")
	}
	other, _ := samltest.KeyPair(t)
	for name, c := range map[string]struct {
		o    samltest.LogoutOptions
		dest string
	}{
		"unsigned":        {samltest.LogoutOptions{Unsigned: true}, w.sp.SLOURL},
		"another key":     {samltest.LogoutOptions{Key: other}, w.sp.SLOURL},
		"another issuer":  {samltest.LogoutOptions{Issuer: "https://evil.example/idp"}, w.sp.SLOURL},
		"for another SP":  {samltest.LogoutOptions{}, "https://other.example/slo"},
		"changed after":   {samltest.LogoutOptions{Tamper: func(q url.Values) { q.Set("RelayState", "x") }}, w.sp.SLOURL},
		"issued long ago": {samltest.LogoutOptions{IssueInstant: timeAgo(10)}, w.sp.SLOURL},
	} {
		if _, err := w.c.SAMLLogoutRequestVerify(w.sp, w.meta, query(c.o, c.dest), w.replay); !errors.Is(err, sso.ErrSAMLLogout) {
			t.Errorf("%s: %v", name, err)
		}
	}

	// The answer goes to the IdP's SLO URL, signed, InResponseTo the request.
	redirect, err := w.c.SAMLLogoutResponse(w.sp, w.meta, lr.ID, lr.RelayState, true)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := http.Get(redirect); err != nil {
		t.Fatal(err)
	}
	lo := w.idp.Logouts()
	if len(lo) != 1 || lo[0].Kind != "response" || lo[0].InResponseTo != "_lr-1" || !strings.HasSuffix(lo[0].Status, ":Success") || lo[0].RelayState != "rs" {
		t.Fatalf("the IdP received %+v", lo)
	}
	u, _ := url.Parse(redirect)
	if err := samltest.VerifyRedirect(u.RawQuery, w.sp.Cert); err != nil {
		t.Fatalf("the response is not signed by the SP: %v", err)
	}
}

func timeAgo(minutes int) time.Time { return time.Now().Add(-time.Duration(minutes) * time.Minute) }
