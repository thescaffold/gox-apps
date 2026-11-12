// Package samltest is a SAML 2.0 identity provider for tests. It publishes
// metadata, reads the service provider's AuthnRequest from the HTTP-Redirect
// binding, and answers with a Response signed (XML-DSig, RSA-SHA256,
// exclusive canonicalization) the way the test asks: the assertion, the
// response, both or neither, by its own key or another, with any field
// changed before or after signing.
//
// It signs with goxmldsig directly, not with crewjam/saml's identity
// provider, so it is a second implementation the service provider is checked
// against.
package samltest

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"html"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/beevik/etree"
	dsig "github.com/russellhaering/goxmldsig"

	"github.com/thescaffold/gox-apps/libs/identity/pkg/sso"
)

// Signing modes.
const (
	SignAssertion = "assertion"
	SignResponse  = "response"
	SignBoth      = "both"
	SignNone      = "none"
)

// IdP is a running test identity provider.
type IdP struct {
	Key      *rsa.PrivateKey
	Cert     *x509.Certificate
	Server   *httptest.Server
	EntityID string

	// WantSigned makes /sso refuse an AuthnRequest that carries no valid
	// signature by the trusted service provider's key (TrustSP); the metadata
	// says WantAuthnRequestsSigned.
	WantSigned bool

	mu      sync.Mutex
	user    Options        // who /sso signs in
	next    func(*Options) // a change to the next response /sso makes
	seen    []Request      // requests /sso received
	spCerts []*x509.Certificate
	spSLO   string
	logouts []Logout // logout messages /slo received
}

// New starts an IdP: metadata at /metadata, sign-on at /sso.
func New(t testing.TB) *IdP {
	t.Helper()
	key, cert := KeyPair(t)
	p := &IdP{Key: key, Cert: cert}
	mux := http.NewServeMux()
	mux.HandleFunc("/metadata", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/samlmetadata+xml")
		_, _ = w.Write([]byte(p.Metadata()))
	})
	mux.HandleFunc("/sso", p.sso)
	mux.HandleFunc("/slo", p.slo)
	p.Server = httptest.NewServer(mux)
	p.EntityID = p.Server.URL + "/metadata"
	t.Cleanup(p.Server.Close)
	return p
}

// KeyPair makes an RSA key and a self-signed certificate for it.
func KeyPair(t testing.TB) (*rsa.PrivateKey, *x509.Certificate) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	tpl := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject:      pkix.Name{CommonName: "samltest"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
	der, err := x509.CreateCertificate(rand.Reader, tpl, tpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	return key, cert
}

// Metadata is the IdP's metadata XML.
func (p *IdP) Metadata() string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<md:EntityDescriptor xmlns:md="urn:oasis:names:tc:SAML:2.0:metadata" entityID="%s">
  <md:IDPSSODescriptor WantAuthnRequestsSigned="%t" protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">
    <md:KeyDescriptor use="signing">
      <ds:KeyInfo xmlns:ds="http://www.w3.org/2000/09/xmldsig#"><ds:X509Data><ds:X509Certificate>%s</ds:X509Certificate></ds:X509Data></ds:KeyInfo>
    </md:KeyDescriptor>
    <md:NameIDFormat>urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress</md:NameIDFormat>
    <md:SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect" Location="%s/sso"/>
    <md:SingleLogoutService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect" Location="%s/slo"/>
  </md:IDPSSODescriptor>
</md:EntityDescriptor>`, html.EscapeString(p.EntityID), p.WantSigned, base64.StdEncoding.EncodeToString(p.Cert.Raw), p.Server.URL, p.Server.URL)
}

// SignIn sets who /sso signs in from now on.
func (p *IdP) SignIn(email, name string) {
	p.mu.Lock()
	p.user = Options{Email: email, Name: name}
	p.mu.Unlock()
}

// Next changes the next response /sso makes (one response only).
func (p *IdP) Next(f func(*Options)) {
	p.mu.Lock()
	p.next = f
	p.mu.Unlock()
}

// Requests are the AuthnRequests /sso has received.
func (p *IdP) Requests() []Request {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]Request(nil), p.seen...)
}

// sso answers an AuthnRequest with an auto-posting form, as an IdP does after
// its own sign-in.
func (p *IdP) sso(w http.ResponseWriter, r *http.Request) {
	req, err := ParseRequest(r.URL.Query().Get("SAMLRequest"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if r.URL.Query().Get("Signature") != "" {
		if err := VerifyRedirect(r.URL.RawQuery, p.trusted()...); err != nil {
			http.Error(w, "bad signature: "+err.Error(), http.StatusBadRequest)
			return
		}
		req.Signed = true
	} else if p.WantSigned {
		http.Error(w, "the AuthnRequest must be signed", http.StatusBadRequest)
		return
	}
	p.mu.Lock()
	o, next := p.user, p.next
	p.next = nil
	p.seen = append(p.seen, req)
	p.mu.Unlock()
	if next != nil {
		next(&o)
	}
	body, err := p.Response(req, o)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = fmt.Fprint(w, Form(req.ACS, body, r.URL.Query().Get("RelayState")))
}

// Form is the HTML-POST binding's auto-submitting form.
func Form(action, samlResponse, relayState string) string {
	return `<!doctype html><html><body onload="document.forms[0].submit()"><form method="post" action="` + html.EscapeString(action) + `">` +
		`<input type="hidden" name="SAMLResponse" value="` + html.EscapeString(samlResponse) + `">` +
		`<input type="hidden" name="RelayState" value="` + html.EscapeString(relayState) + `">` +
		`<noscript><button type="submit">Continue</button></noscript></form></body></html>`
}

var (
	formAction = regexp.MustCompile(`<form[^>]*action="([^"]*)"`)
	formInput  = regexp.MustCompile(`<input type="hidden" name="([^"]+)" value="([^"]*)">`)
)

// ParseForm reads Form's action and fields back.
func ParseForm(page string) (action string, fields url.Values) {
	fields = url.Values{}
	if m := formAction.FindStringSubmatch(page); m != nil {
		action = html.UnescapeString(m[1])
	}
	for _, m := range formInput.FindAllStringSubmatch(page, -1) {
		fields.Set(m[1], html.UnescapeString(m[2]))
	}
	return action, fields
}

// Request is what the service provider asked for.
type Request struct {
	ID     string
	ACS    string
	Issuer string // the SP's entity id
	// Signed: the HTTP-Redirect query carried a valid signature (set by /sso).
	Signed bool
}

// ParseRequest decodes an HTTP-Redirect SAMLRequest.
func ParseRequest(v string) (Request, error) {
	raw, err := sso.DecodeSAMLRequest(v)
	if err != nil {
		return Request{}, err
	}
	var ar struct {
		ID     string `xml:"ID,attr"`
		ACS    string `xml:"AssertionConsumerServiceURL,attr"`
		Issuer string `xml:"urn:oasis:names:tc:SAML:2.0:assertion Issuer"`
	}
	if err := xml.Unmarshal(raw, &ar); err != nil {
		return Request{}, err
	}
	return Request{ID: ar.ID, ACS: ar.ACS, Issuer: ar.Issuer}, nil
}

// ParseRedirect reads the request and RelayState from the URL the service
// provider sends the browser to.
func ParseRedirect(raw string) (Request, string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return Request{}, "", err
	}
	req, err := ParseRequest(u.Query().Get("SAMLRequest"))
	return req, u.Query().Get("RelayState"), err
}

// Options shape a response. Zero values answer the request correctly.
type Options struct {
	Email, Name string
	// NameID defaults to Email.
	NameID string
	// Each defaults from the request: Audience to its issuer, Recipient and
	// Destination to its ACS, InResponseTo to its id. "-" leaves it out.
	Audience, Recipient, Destination, InResponseTo string
	// Issuer defaults to the IdP's entity id.
	Issuer      string
	AssertionID string
	// SessionIndex defaults to a fresh one.
	SessionIndex string
	// Times default to now, now-1m and now+5m.
	IssueInstant, NotBefore, NotOnOrAfter time.Time
	// Sign defaults to SignAssertion.
	Sign string
	// Key signs instead of the IdP's own (its certificate goes in KeyInfo).
	Key *rsa.PrivateKey
	// Status defaults to Success.
	Status string
	// Tamper changes the document after it is signed.
	Tamper func(response *etree.Element)
}

const (
	nsP = "urn:oasis:names:tc:SAML:2.0:protocol"
	nsA = "urn:oasis:names:tc:SAML:2.0:assertion"
)

func or(v, d string) string {
	switch v {
	case "":
		return d
	case "-":
		return ""
	}
	return v
}

func id() string {
	b := make([]byte, 20)
	_, _ = rand.Read(b)
	return fmt.Sprintf("_%x", b)
}

func ts(t time.Time) string { return t.UTC().Format("2006-01-02T15:04:05.000Z") }

// Response builds the base64 SAMLResponse for req.
func (p *IdP) Response(req Request, o Options) (string, error) {
	now := time.Now()
	if o.IssueInstant.IsZero() {
		o.IssueInstant = now
	}
	if o.NotBefore.IsZero() {
		o.NotBefore = now.Add(-time.Minute)
	}
	if o.NotOnOrAfter.IsZero() {
		o.NotOnOrAfter = now.Add(5 * time.Minute)
	}
	if o.AssertionID == "" {
		o.AssertionID = id()
	}
	issuer := or(o.Issuer, p.EntityID)
	inResponseTo := or(o.InResponseTo, req.ID)

	resp := etree.NewElement("samlp:Response")
	resp.CreateAttr("xmlns:samlp", nsP)
	resp.CreateAttr("xmlns:saml", nsA)
	resp.CreateAttr("ID", id())
	resp.CreateAttr("Version", "2.0")
	resp.CreateAttr("IssueInstant", ts(o.IssueInstant))
	if d := or(o.Destination, req.ACS); d != "" {
		resp.CreateAttr("Destination", d)
	}
	if inResponseTo != "" {
		resp.CreateAttr("InResponseTo", inResponseTo)
	}
	resp.CreateElement("saml:Issuer").SetText(issuer)
	resp.CreateElement("samlp:Status").CreateElement("samlp:StatusCode").CreateAttr("Value", or(o.Status, "urn:oasis:names:tc:SAML:2.0:status:Success"))

	if o.Status == "" {
		a := resp.CreateElement("saml:Assertion")
		a.CreateAttr("xmlns:saml", nsA)
		a.CreateAttr("ID", o.AssertionID)
		a.CreateAttr("Version", "2.0")
		a.CreateAttr("IssueInstant", ts(o.IssueInstant))
		a.CreateElement("saml:Issuer").SetText(issuer)
		subj := a.CreateElement("saml:Subject")
		nid := subj.CreateElement("saml:NameID")
		nid.CreateAttr("Format", "urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress")
		nid.SetText(or(o.NameID, o.Email))
		sc := subj.CreateElement("saml:SubjectConfirmation")
		sc.CreateAttr("Method", "urn:oasis:names:tc:SAML:2.0:cm:bearer")
		scd := sc.CreateElement("saml:SubjectConfirmationData")
		if inResponseTo != "" {
			scd.CreateAttr("InResponseTo", inResponseTo)
		}
		scd.CreateAttr("NotOnOrAfter", ts(o.NotOnOrAfter))
		if r := or(o.Recipient, req.ACS); r != "" {
			scd.CreateAttr("Recipient", r)
		}
		cond := a.CreateElement("saml:Conditions")
		cond.CreateAttr("NotBefore", ts(o.NotBefore))
		cond.CreateAttr("NotOnOrAfter", ts(o.NotOnOrAfter))
		if aud := or(o.Audience, req.Issuer); aud != "" {
			cond.CreateElement("saml:AudienceRestriction").CreateElement("saml:Audience").SetText(aud)
		}
		as := a.CreateElement("saml:AuthnStatement")
		as.CreateAttr("AuthnInstant", ts(o.IssueInstant))
		as.CreateAttr("SessionIndex", or(o.SessionIndex, id()))
		as.CreateElement("saml:AuthnContext").CreateElement("saml:AuthnContextClassRef").SetText("urn:oasis:names:tc:SAML:2.0:ac:classes:PasswordProtectedTransport")
		st := a.CreateElement("saml:AttributeStatement")
		for name, v := range map[string]string{"email": or(o.Email, ""), "displayName": or(o.Name, "")} {
			if v == "" {
				continue
			}
			at := st.CreateElement("saml:Attribute")
			at.CreateAttr("Name", name)
			at.CreateAttr("NameFormat", "urn:oasis:names:tc:SAML:2.0:attrname-format:basic")
			at.CreateElement("saml:AttributeValue").SetText(v)
		}
		if o.Sign == "" || o.Sign == SignAssertion || o.Sign == SignBoth {
			if err := p.sign(a, o.Key); err != nil {
				return "", err
			}
		}
	}
	if o.Sign == SignResponse || o.Sign == SignBoth {
		if err := p.sign(resp, o.Key); err != nil {
			return "", err
		}
	}
	if o.Tamper != nil {
		o.Tamper(resp)
	}
	doc := etree.NewDocument()
	doc.SetRoot(resp)
	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// sign puts an enveloped signature right after el's Issuer, where the SAML
// schema wants it.
func (p *IdP) sign(el *etree.Element, key *rsa.PrivateKey) error {
	if key == nil {
		key = p.Key
	}
	ctx, err := dsig.NewSigningContext(key, [][]byte{p.Cert.Raw})
	if err != nil {
		return err
	}
	ctx.Canonicalizer = dsig.MakeC14N10ExclusiveCanonicalizerWithPrefixList("")
	if err := ctx.SetSignatureMethod(dsig.RSASHA256SignatureMethod); err != nil {
		return err
	}
	sig, err := ctx.ConstructSignature(el, true)
	if err != nil {
		return err
	}
	el.InsertChildAt(1, sig)
	return nil
}
