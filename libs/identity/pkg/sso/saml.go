package sso

// SAML 2.0, the second protocol a connection can speak (next to OpenID
// Connect in oidc.go). Identity is the service provider; the flow is
// SP-initiated only:
//
//	the browser is sent to the IdP with an AuthnRequest (HTTP-Redirect binding);
//	the IdP posts a Response back to the ACS (HTTP-POST binding).
//
// The service provider holds its own signing key (saml_keys.go, V4-03): with it
// the AuthnRequest is signed when the connection asks, and single logout
// (saml_slo.go) works in both directions. The key is never used to decrypt:
// encrypted assertions stay refused (D-15), and the metadata publishes no
// encryption key.
//
// The XML, signature and condition checks are github.com/crewjam/saml
// (BSD-2-Clause): a signed Response or a signed Assertion from a key in the
// IdP's metadata, Destination and Recipient equal to the ACS, the audience
// equal to the SP entity ID, InResponseTo equal to the request this sign-in
// started, IssueInstant/NotBefore/NotOnOrAfter. On top of the library this
// file holds the clock to SAMLSkew (the library allows three minutes), refuses
// SHA-1 signatures and unsolicited (IdP-initiated) responses, and records every
// accepted assertion id so it is never accepted twice (Store.ReplayCache).
//
// Like oidc.go it knows nothing of identity's users or tokens.

import (
	"bytes"
	"compress/flate"
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"net/url"
	"strings"
	"time"

	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
	dsig "github.com/russellhaering/goxmldsig"
)

// Protocols a connection speaks.
const (
	ProtocolOIDC = "oidc"
	ProtocolSAML = "saml"
)

// SAMLSkew is the clock difference allowed between the IdP and this server.
const SAMLSkew = 60 * time.Second

// Errors a caller can match.
var (
	ErrSAMLMetadata = errors.New("sso: the SAML metadata is not usable")
	// ErrSAMLDenied: the IdP answered, but did not sign the person in (a
	// non-success status: they cancelled, or the IdP refused them).
	ErrSAMLDenied   = errors.New("sso: the identity provider did not sign the person in")
	ErrSAMLResponse = errors.New("sso: the SAML response is not valid")
)

// SAMLIdP is an identity provider's metadata, checked.
type SAMLIdP struct {
	EntityID string
	// SSOURL is its SingleSignOnService for the HTTP-Redirect binding.
	SSOURL string
	// SLOURL is its SingleLogoutService for the HTTP-Redirect binding; "" when
	// it offers no single logout.
	SLOURL string
	// WantsSignedRequests is the metadata's WantAuthnRequestsSigned.
	WantsSignedRequests bool
	metadata            *saml.EntityDescriptor
}

// SAMLServiceProvider is this side of a connection.
type SAMLServiceProvider struct {
	EntityID string
	ACSURL   string
	// SLOURL is this side's SingleLogoutService (HTTP-Redirect); "" offers none.
	SLOURL string
	// Key signs what this side sends (AuthnRequests when SignRequests, and
	// every logout message); Cert is its certificate. NextCert, a key staged
	// by a rotation and not yet signing, is published beside it so the IdP
	// can learn it first. With no Key nothing is signed and there is no
	// single logout.
	Key          *rsa.PrivateKey
	Cert         *x509.Certificate
	NextCert     *x509.Certificate
	SignRequests bool
}

const maxMetadata = 1 << 20

// ParseSAMLMetadata checks IdP metadata: one IdP entity with a signing
// certificate and a SingleSignOnService for the HTTP-Redirect binding (https,
// unless AllowPrivate).
func (c *Client) ParseSAMLMetadata(raw []byte) (*SAMLIdP, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || len(raw) > maxMetadata {
		return nil, fmt.Errorf("%w: empty or too large", ErrSAMLMetadata)
	}
	ed, err := samlsp.ParseMetadata(raw)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSAMLMetadata, err)
	}
	if ed.EntityID == "" || len(ed.IDPSSODescriptors) == 0 {
		return nil, fmt.Errorf("%w: not an identity provider's metadata", ErrSAMLMetadata)
	}
	idp := &SAMLIdP{EntityID: ed.EntityID, metadata: ed}
	certs := 0
	for _, d := range ed.IDPSSODescriptors {
		for _, s := range d.SingleSignOnServices {
			if s.Binding == saml.HTTPRedirectBinding && idp.SSOURL == "" {
				idp.SSOURL = s.Location
			}
		}
		for _, s := range d.SingleLogoutServices {
			if s.Binding == saml.HTTPRedirectBinding && idp.SLOURL == "" && c.checkURL(s.Location) == nil {
				idp.SLOURL = s.Location
			}
		}
		if d.WantAuthnRequestsSigned != nil && *d.WantAuthnRequestsSigned {
			idp.WantsSignedRequests = true
		}
		for _, k := range d.KeyDescriptors {
			if k.Use != "" && k.Use != "signing" {
				continue
			}
			for _, xc := range k.KeyInfo.X509Data.X509Certificates {
				if _, err := parseCert(xc.Data); err == nil {
					certs++
				}
			}
		}
	}
	if idp.SSOURL == "" {
		return nil, fmt.Errorf("%w: no SingleSignOnService for the HTTP-Redirect binding", ErrSAMLMetadata)
	}
	if err := c.checkURL(idp.SSOURL); err != nil {
		return nil, fmt.Errorf("%w: the sign-on URL must be https", ErrSAMLMetadata)
	}
	if certs == 0 {
		return nil, fmt.Errorf("%w: no signing certificate", ErrSAMLMetadata)
	}
	return idp, nil
}

func parseCert(data string) (*x509.Certificate, error) {
	if blk, _ := pem.Decode([]byte(strings.TrimSpace(data))); blk != nil {
		return x509.ParseCertificate(blk.Bytes)
	}
	b, err := base64.StdEncoding.DecodeString(strings.Join(strings.Fields(data), ""))
	if err != nil {
		return nil, err
	}
	return x509.ParseCertificate(b)
}

type metaEntry struct {
	raw []byte
	exp time.Time
}

// FetchSAMLMetadata downloads IdP metadata (cached for an hour unless fresh is
// set). The URL is untrusted input, so the request goes through the same
// non-public-address guard as OpenID discovery.
func (c *Client) FetchSAMLMetadata(ctx context.Context, raw string, fresh bool) ([]byte, error) {
	if err := c.checkURL(raw); err != nil {
		return nil, fmt.Errorf("%w: the metadata URL must be https", ErrSAMLMetadata)
	}
	c.mu.Lock()
	if c.meta == nil {
		c.meta = map[string]metaEntry{}
	}
	if e, ok := c.meta[raw]; ok && !fresh && c.now().Before(e.exp) {
		c.mu.Unlock()
		return e.raw, nil
	}
	c.mu.Unlock()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, raw, nil)
	req.Header.Set("Accept", "application/samlmetadata+xml, application/xml, text/xml")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSAMLMetadata, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %s returned %d", ErrSAMLMetadata, raw, resp.StatusCode)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, maxMetadata+1))
	if err != nil || len(b) > maxMetadata {
		return nil, fmt.Errorf("%w: unreadable or too large", ErrSAMLMetadata)
	}
	c.mu.Lock()
	c.meta[raw] = metaEntry{raw: b, exp: c.now().Add(cacheTTL)}
	c.mu.Unlock()
	return b, nil
}

// serviceProvider builds the library's service provider. signing gives it the
// key: only for what this side sends, never to read a response (a key would let
// the library decrypt an encrypted assertion, which D-15 refuses).
func (c *Client) serviceProvider(sp SAMLServiceProvider, idp *SAMLIdP, replay saml.AssertionReplayCache, signing ...bool) (*saml.ServiceProvider, error) {
	ent, err1 := url.Parse(sp.EntityID)
	acs, err2 := url.Parse(sp.ACSURL)
	if err1 != nil || err2 != nil || sp.EntityID == "" || sp.ACSURL == "" {
		return nil, fmt.Errorf("sso: bad service provider URLs")
	}
	var slo *url.URL
	if sp.SLOURL != "" {
		u, err := url.Parse(sp.SLOURL)
		if err != nil {
			return nil, fmt.Errorf("sso: bad service provider URLs")
		}
		slo = u
	}
	s := &saml.ServiceProvider{
		EntityID:                         sp.EntityID,
		MetadataURL:                      *ent,
		AcsURL:                           *acs,
		AuthnNameIDFormat:                saml.EmailAddressNameIDFormat,
		AssertionConsumerServiceBindings: []string{saml.HTTPPostBinding},
		RejectSHA1Signatures:             true,
		ReplayCache:                      replay,
	}
	if idp != nil {
		s.IDPMetadata = idp.metadata
	}
	if slo != nil {
		s.SloURL = *slo
		s.LogoutBindings = []string{saml.HTTPRedirectBinding}
	}
	if len(signing) > 0 && signing[0] && sp.Key != nil && sp.Cert != nil {
		s.Key, s.Certificate = sp.Key, sp.Cert
		s.LogoutSignatureMethod = dsig.RSASHA256SignatureMethod
		if sp.SignRequests {
			s.SignatureMethod = dsig.RSASHA256SignatureMethod
		}
	}
	if c.Now != nil {
		s.Clock = dsig.NewFakeClockAt(c.Now())
	}
	return s, nil
}

// SAMLAuthnRequest builds the HTTP-Redirect URL that sends the browser to the
// IdP, and returns the request's id: the response must answer it.
//
// With the service provider's key and SignRequests, the query string is signed
// (RSA-SHA256, saml-bindings-2.0 §3.4.4.1).
func (c *Client) SAMLAuthnRequest(sp SAMLServiceProvider, idp *SAMLIdP, relayState string) (redirect, requestID string, err error) {
	s, err := c.serviceProvider(sp, idp, nil, true)
	if err != nil {
		return "", "", err
	}
	req, err := s.MakeAuthenticationRequest(idp.SSOURL, saml.HTTPRedirectBinding, saml.HTTPPostBinding)
	if err != nil {
		return "", "", err
	}
	u, err := req.Redirect(relayState, s)
	if err != nil {
		return "", "", err
	}
	return u.String(), req.ID, nil
}

// SAMLMetadata is this service provider's metadata, for the IdP's admin: its
// entity id and ACS, its single logout service, and its signing certificate(s)
// (the current one, and one staged by a rotation). No encryption key is ever
// published, so no IdP encrypts to this side (D-15).
func (c *Client) SAMLMetadata(sp SAMLServiceProvider) ([]byte, error) {
	s, err := c.serviceProvider(sp, nil, nil, true)
	if err != nil {
		return nil, err
	}
	ed := s.Metadata()
	for i := range ed.SPSSODescriptors {
		d := &ed.SPSSODescriptors[i]
		kept := d.KeyDescriptors[:0]
		for _, kd := range d.KeyDescriptors {
			if kd.Use == "signing" {
				kept = append(kept, kd)
			}
		}
		d.KeyDescriptors = kept
		if sp.NextCert != nil && s.Certificate != nil {
			d.KeyDescriptors = append(d.KeyDescriptors, saml.KeyDescriptor{Use: "signing", KeyInfo: saml.KeyInfo{
				X509Data: saml.X509Data{X509Certificates: []saml.X509Certificate{{Data: base64.StdEncoding.EncodeToString(sp.NextCert.Raw)}}},
			}})
		}
	}
	b, err := xml.MarshalIndent(ed, "", "  ")
	if err != nil {
		return nil, err
	}
	return append([]byte(xml.Header), b...), nil
}

// SAMLVerify checks a posted SAMLResponse (base64, as the form carries it)
// against the request it must answer, and returns who it signs in. replay
// records the assertion's id; nil turns that off (never in production).
func (c *Client) SAMLVerify(sp SAMLServiceProvider, idp *SAMLIdP, samlResponse, requestID string, replay saml.AssertionReplayCache) (*Claims, error) {
	if requestID == "" {
		return nil, fmt.Errorf("%w: no request to answer", ErrSAMLResponse)
	}
	raw, err := base64.StdEncoding.DecodeString(strings.Join(strings.Fields(samlResponse), ""))
	if err != nil || len(raw) == 0 {
		return nil, fmt.Errorf("%w: not base64", ErrSAMLResponse)
	}
	// D-15: an IdP that encrypts is refused, whatever key this side holds.
	if bytes.Contains(raw, []byte("EncryptedAssertion")) || bytes.Contains(raw, []byte("EncryptedID")) || bytes.Contains(raw, []byte("EncryptedAttribute")) {
		return nil, fmt.Errorf("%w: encrypted assertions are not accepted; turn assertion encryption off at the identity provider", ErrSAMLResponse)
	}
	s, err := c.serviceProvider(sp, idp, replay)
	if err != nil {
		return nil, err
	}
	// Hold the times to SAMLSkew before the library (which allows more) looks.
	if err := c.checkTimes(raw); err != nil {
		return nil, err
	}
	a, err := s.ParseXMLResponse(raw, []string{requestID}, s.AcsURL)
	if err != nil {
		var ire *saml.InvalidResponseError
		if errors.As(err, &ire) {
			err = ire.PrivateErr
		}
		var bad saml.ErrBadStatus
		if errors.As(err, &bad) {
			return nil, fmt.Errorf("%w: %s", ErrSAMLDenied, bad.Status)
		}
		return nil, fmt.Errorf("%w: %v", ErrSAMLResponse, err)
	}
	cl := samlClaims(a)
	if cl.Email == "" {
		return nil, fmt.Errorf("%w: no email address in the assertion", ErrSAMLResponse)
	}
	return cl, nil
}

// checkTimes reads the response's own timestamps (not yet trusted; this only
// refuses earlier than the library would) and holds them to SAMLSkew.
func (c *Client) checkTimes(raw []byte) error {
	var r struct {
		IssueInstant time.Time `xml:"IssueInstant,attr"`
		Assertions   []struct {
			IssueInstant time.Time `xml:"IssueInstant,attr"`
			Conditions   *struct {
				NotBefore    time.Time `xml:"NotBefore,attr"`
				NotOnOrAfter time.Time `xml:"NotOnOrAfter,attr"`
			} `xml:"urn:oasis:names:tc:SAML:2.0:assertion Conditions"`
			Confirmations []struct {
				Data *struct {
					NotOnOrAfter time.Time `xml:"NotOnOrAfter,attr"`
				} `xml:"urn:oasis:names:tc:SAML:2.0:assertion SubjectConfirmationData"`
			} `xml:"urn:oasis:names:tc:SAML:2.0:assertion Subject>SubjectConfirmation"`
		} `xml:"urn:oasis:names:tc:SAML:2.0:assertion Assertion"`
	}
	if err := xml.Unmarshal(raw, &r); err != nil {
		return fmt.Errorf("%w: %v", ErrSAMLResponse, err)
	}
	now := c.now()
	future := func(t time.Time) bool { return !t.IsZero() && t.After(now.Add(SAMLSkew)) }
	past := func(t time.Time) bool { return !t.IsZero() && !now.Before(t.Add(SAMLSkew)) }
	if future(r.IssueInstant) {
		return fmt.Errorf("%w: issued in the future", ErrSAMLResponse)
	}
	for _, a := range r.Assertions {
		if future(a.IssueInstant) {
			return fmt.Errorf("%w: assertion issued in the future", ErrSAMLResponse)
		}
		if a.Conditions != nil {
			if future(a.Conditions.NotBefore) {
				return fmt.Errorf("%w: assertion not yet valid", ErrSAMLResponse)
			}
			if past(a.Conditions.NotOnOrAfter) {
				return fmt.Errorf("%w: assertion expired", ErrSAMLResponse)
			}
		}
		for _, sc := range a.Confirmations {
			if sc.Data != nil && past(sc.Data.NotOnOrAfter) {
				return fmt.Errorf("%w: subject confirmation expired", ErrSAMLResponse)
			}
		}
	}
	return nil
}

// Attribute names IdPs use for the address and the name (Okta, Entra ID,
// Google, ADFS, Shibboleth/eduPerson OIDs).
var (
	samlEmailAttrs = []string{"email", "mail", "emailaddress", "e-mail", "user.email",
		"http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress", "urn:oid:0.9.2342.19200300.100.1.3"}
	samlNameAttrs = []string{"displayname", "name", "cn", "fullname",
		"http://schemas.microsoft.com/identity/claims/displayname", "urn:oid:2.16.840.1.113730.3.1.241", "urn:oid:2.5.4.3"}
	samlGivenAttrs = []string{"givenname", "firstname", "first_name",
		"http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname", "urn:oid:2.5.4.42"}
	samlSurnameAttrs = []string{"surname", "sn", "lastname", "last_name", "familyname",
		"http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname", "urn:oid:2.5.4.4"}
)

func samlClaims(a *saml.Assertion) *Claims {
	attr := func(names []string) string {
		for _, want := range names {
			for _, st := range a.AttributeStatements {
				for _, at := range st.Attributes {
					if strings.EqualFold(at.Name, want) || strings.EqualFold(at.FriendlyName, want) {
						for _, v := range at.Values {
							if s := strings.TrimSpace(v.Value); s != "" {
								return s
							}
						}
					}
				}
			}
		}
		return ""
	}
	cl := &Claims{EmailVerified: true} // the company's own IdP vouches; the caller checks the domain
	if a.Subject != nil && a.Subject.NameID != nil {
		cl.Subject = a.Subject.NameID.Value
	}
	for _, st := range a.AuthnStatements {
		if st.SessionIndex != "" {
			cl.SessionIndex = st.SessionIndex
			break
		}
	}
	email := attr(samlEmailAttrs)
	if email == "" && strings.Contains(cl.Subject, "@") {
		email = cl.Subject
	}
	if addr, err := mail.ParseAddress(email); err == nil && addr.Name == "" && EmailDomain(addr.Address) != "" {
		cl.Email = strings.ToLower(addr.Address)
	}
	if cl.Subject == "" {
		cl.Subject = cl.Email
	}
	cl.Name = attr(samlNameAttrs)
	if cl.Name == "" {
		cl.Name = strings.TrimSpace(attr(samlGivenAttrs) + " " + attr(samlSurnameAttrs))
	}
	return cl
}

// DecodeSAMLRequest inflates an HTTP-Redirect SAMLRequest (tests and test
// identity providers read the request id and ACS from it).
func DecodeSAMLRequest(v string) ([]byte, error) {
	b, err := base64.StdEncoding.DecodeString(v)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(io.LimitReader(flate.NewReader(bytes.NewReader(b)), 1<<20))
}
