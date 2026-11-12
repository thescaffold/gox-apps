package samltest

import (
	"bytes"
	"compress/flate"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Single logout (HTTP-Redirect binding, saml-bindings-2.0 §3.4) and signed
// AuthnRequests. Every message either side sends here is signed over its
// query string with RSA-SHA256, as the profile requires.

const sigAlgRSASHA256 = "http://www.w3.org/2001/04/xmldsig-more#rsa-sha256"

// TrustSP registers the service provider's signing certificate(s), which the
// IdP checks AuthnRequests and logout messages against, and its single
// logout URL, where the IdP's own logout messages go.
func (p *IdP) TrustSP(slo string, certs ...*x509.Certificate) {
	p.mu.Lock()
	p.spSLO, p.spCerts = slo, certs
	p.mu.Unlock()
}

func (p *IdP) trusted() []*x509.Certificate {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]*x509.Certificate(nil), p.spCerts...)
}

// Logout is a logout message /slo received from the service provider.
type Logout struct {
	Kind         string // "request" or "response"
	ID           string
	InResponseTo string
	NameID       string
	SessionIndex string
	Status       string
	Destination  string
	Issuer       string
	RelayState   string
}

// Logouts are the logout messages /slo has received.
func (p *IdP) Logouts() []Logout {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]Logout(nil), p.logouts...)
}

// slo is the IdP's SingleLogoutService. A LogoutRequest (the SP signing a
// person out) is answered with a signed LogoutResponse sent to the SP's SLO
// URL; a LogoutResponse (the SP answering the IdP's own request) is recorded.
// Unsigned or wrongly signed messages are refused.
func (p *IdP) slo(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if err := VerifyRedirect(r.URL.RawQuery, p.trusted()...); err != nil {
		http.Error(w, "bad signature: "+err.Error(), http.StatusBadRequest)
		return
	}
	if v := q.Get("SAMLResponse"); v != "" {
		l, err := parseLogout(v, "response")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		l.RelayState = q.Get("RelayState")
		p.mu.Lock()
		p.logouts = append(p.logouts, l)
		p.mu.Unlock()
		_, _ = fmt.Fprint(w, "signed out")
		return
	}
	l, err := parseLogout(q.Get("SAMLRequest"), "request")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	l.RelayState = q.Get("RelayState")
	p.mu.Lock()
	p.logouts = append(p.logouts, l)
	slo := p.spSLO
	p.mu.Unlock()
	if slo == "" {
		http.Error(w, "no service provider SLO URL (TrustSP)", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, p.LogoutResponseURL(slo, l.ID, l.RelayState, LogoutOptions{}), http.StatusFound)
}

// LogoutOptions shape a logout message the IdP makes.
type LogoutOptions struct {
	// Key signs instead of the IdP's own; Unsigned sends no signature.
	Key      *rsa.PrivateKey
	Unsigned bool
	// Issuer defaults to the IdP's entity id; ID to a fresh one.
	Issuer, ID string
	// IssueInstant defaults to now.
	IssueInstant time.Time
	// Status (responses) defaults to Success.
	Status string
	// Tamper changes the encoded message after it is signed.
	Tamper func(query url.Values)
}

// LogoutRequestURL is an IdP-initiated LogoutRequest for nameID (and
// sessionIndex, if any) sent to the service provider's SLO URL.
func (p *IdP) LogoutRequestURL(destination, nameID, sessionIndex, relayState string, o LogoutOptions) string {
	now := o.IssueInstant
	if now.IsZero() {
		now = time.Now()
	}
	msgID := or(o.ID, id())
	idx := ""
	if sessionIndex != "" {
		idx = `<samlp:SessionIndex>` + esc(sessionIndex) + `</samlp:SessionIndex>`
	}
	x := `<samlp:LogoutRequest xmlns:samlp="` + nsP + `" xmlns:saml="` + nsA + `" ID="` + msgID + `" Version="2.0" IssueInstant="` + ts(now) +
		`" Destination="` + esc(destination) + `"><saml:Issuer>` + esc(or(o.Issuer, p.EntityID)) + `</saml:Issuer>` +
		`<saml:NameID Format="urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress">` + esc(nameID) + `</saml:NameID>` + idx + `</samlp:LogoutRequest>`
	return p.redirect(destination, "SAMLRequest", x, relayState, o)
}

// LogoutResponseURL answers the service provider's LogoutRequest inResponseTo.
func (p *IdP) LogoutResponseURL(destination, inResponseTo, relayState string, o LogoutOptions) string {
	now := o.IssueInstant
	if now.IsZero() {
		now = time.Now()
	}
	x := `<samlp:LogoutResponse xmlns:samlp="` + nsP + `" xmlns:saml="` + nsA + `" ID="` + or(o.ID, id()) + `" Version="2.0" IssueInstant="` + ts(now) +
		`" Destination="` + esc(destination) + `" InResponseTo="` + esc(inResponseTo) + `"><saml:Issuer>` + esc(or(o.Issuer, p.EntityID)) + `</saml:Issuer>` +
		`<samlp:Status><samlp:StatusCode Value="` + esc(or(o.Status, "urn:oasis:names:tc:SAML:2.0:status:Success")) + `"/></samlp:Status></samlp:LogoutResponse>`
	return p.redirect(destination, "SAMLResponse", x, relayState, o)
}

func esc(v string) string {
	var b bytes.Buffer
	_ = xml.EscapeText(&b, []byte(v))
	return strings.ReplaceAll(b.String(), `"`, "&quot;")
}

// redirect encodes a message for the HTTP-Redirect binding and signs it.
func (p *IdP) redirect(destination, param, message, relayState string, o LogoutOptions) string {
	enc := deflate(message)
	signed := param + "=" + url.QueryEscape(enc)
	if relayState != "" {
		signed += "&RelayState=" + url.QueryEscape(relayState)
	}
	q := signed
	if !o.Unsigned {
		key := o.Key
		if key == nil {
			key = p.Key
		}
		signed += "&SigAlg=" + url.QueryEscape(sigAlgRSASHA256)
		h := sha256.Sum256([]byte(signed))
		sig, _ := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, h[:])
		q = signed + "&Signature=" + url.QueryEscape(base64.StdEncoding.EncodeToString(sig))
	}
	if o.Tamper != nil {
		v, _ := url.ParseQuery(q)
		o.Tamper(v)
		q = v.Encode()
	}
	sep := "?"
	if strings.Contains(destination, "?") {
		sep = "&"
	}
	return destination + sep + q
}

func deflate(s string) string {
	var b bytes.Buffer
	w, _ := flate.NewWriter(&b, flate.BestCompression)
	_, _ = w.Write([]byte(s))
	_ = w.Close()
	return base64.StdEncoding.EncodeToString(b.Bytes())
}

func inflate(v string) ([]byte, error) {
	b, err := base64.StdEncoding.DecodeString(v)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(io.LimitReader(flate.NewReader(bytes.NewReader(b)), 1<<20))
}

// VerifyRedirect checks an HTTP-Redirect binding query string's signature
// (RSA-SHA256) against certs, over the parameters exactly as they were sent.
func VerifyRedirect(rawQuery string, certs ...*x509.Certificate) error {
	vals := map[string]string{}
	for _, pair := range strings.Split(rawQuery, "&") {
		k, v, _ := strings.Cut(pair, "=")
		vals[k] = v
	}
	if vals["Signature"] == "" || vals["SigAlg"] == "" {
		return errors.New("not signed")
	}
	alg, _ := url.QueryUnescape(vals["SigAlg"])
	if alg != sigAlgRSASHA256 {
		return fmt.Errorf("unexpected SigAlg %s", alg)
	}
	param := "SAMLRequest"
	if _, ok := vals["SAMLResponse"]; ok {
		param = "SAMLResponse"
	}
	signed := param + "=" + vals[param]
	if rs, ok := vals["RelayState"]; ok {
		signed += "&RelayState=" + rs
	}
	signed += "&SigAlg=" + vals["SigAlg"]
	raw, _ := url.QueryUnescape(vals["Signature"])
	sig, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return err
	}
	h := sha256.Sum256([]byte(signed))
	for _, c := range certs {
		if pub, ok := c.PublicKey.(*rsa.PublicKey); ok && rsa.VerifyPKCS1v15(pub, crypto.SHA256, h[:], sig) == nil {
			return nil
		}
	}
	return errors.New("no trusted key made this signature")
}

func parseLogout(v, kind string) (Logout, error) {
	raw, err := inflate(v)
	if err != nil {
		return Logout{}, err
	}
	var m struct {
		ID           string `xml:"ID,attr"`
		InResponseTo string `xml:"InResponseTo,attr"`
		Destination  string `xml:"Destination,attr"`
		Issuer       string `xml:"urn:oasis:names:tc:SAML:2.0:assertion Issuer"`
		NameID       string `xml:"urn:oasis:names:tc:SAML:2.0:assertion NameID"`
		SessionIndex string `xml:"urn:oasis:names:tc:SAML:2.0:protocol SessionIndex"`
		Status       struct {
			Code struct {
				Value string `xml:"Value,attr"`
			} `xml:"urn:oasis:names:tc:SAML:2.0:protocol StatusCode"`
		} `xml:"urn:oasis:names:tc:SAML:2.0:protocol Status"`
	}
	if err := xml.Unmarshal(raw, &m); err != nil {
		return Logout{}, err
	}
	return Logout{Kind: kind, ID: m.ID, InResponseTo: m.InResponseTo, NameID: m.NameID, SessionIndex: m.SessionIndex,
		Status: m.Status.Code.Value, Destination: m.Destination, Issuer: m.Issuer}, nil
}
