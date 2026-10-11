// Package sso is enterprise single sign-on for identity: OpenID Connect
// (authorization-code flow with PKCE) against a company's own identity
// provider. It is per-workspace configuration: a workspace registers its IdP
// and the email domains it owns; people with those emails sign in through it.
//
// This file is the protocol client. It has no knowledge of identity's users
// or tokens, so it can be tested against any OIDC server.
package sso

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// Errors a caller can match.
var (
	ErrDiscovery = errors.New("sso: the identity provider could not be reached or is misconfigured")
	ErrToken     = errors.New("sso: the identity provider rejected the sign-in")
	ErrIDToken   = errors.New("sso: the identity token is not valid")
)

// Discovery is the subset of the OpenID configuration the flow needs.
type Discovery struct {
	Issuer                string `json:"issuer"`
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	JWKSURI               string `json:"jwks_uri"`
}

// Claims are the verified identity-token claims the sign-in uses.
type Claims struct {
	Subject       string
	Email         string
	EmailVerified bool
	Name          string
}

// Client talks to identity providers. Zero value is not usable; use New.
type Client struct {
	// AllowPrivate lets issuer, token and key URLs resolve to loopback and
	// private addresses over plain HTTP. For tests and local development only.
	AllowPrivate bool
	// Now overrides the clock (tests).
	Now func() time.Time

	http *http.Client
	mu   sync.Mutex
	disc map[string]discEntry
	jwks map[string]jwksEntry
}

type discEntry struct {
	d   *Discovery
	exp time.Time
}

type jwksEntry struct {
	keys map[string]*rsa.PublicKey
	exp  time.Time
}

const cacheTTL = time.Hour

// New returns a Client whose outbound requests refuse private addresses
// unless allowPrivate is set (issuer URLs are supplied by workspace admins,
// so they are untrusted input: SSRF).
func New(allowPrivate bool) *Client {
	c := &Client{AllowPrivate: allowPrivate, disc: map[string]discEntry{}, jwks: map[string]jwksEntry{}}
	dialer := &net.Dialer{Timeout: 5 * time.Second}
	c.http = &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				host, port, err := net.SplitHostPort(addr)
				if err != nil {
					return nil, err
				}
				ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
				if err != nil {
					return nil, err
				}
				for _, ip := range ips {
					if !c.AllowPrivate && !publicIP(ip.IP) {
						return nil, fmt.Errorf("sso: %s resolves to a non-public address", host)
					}
				}
				if len(ips) == 0 {
					return nil, fmt.Errorf("sso: %s did not resolve", host)
				}
				return dialer.DialContext(ctx, network, net.JoinHostPort(ips[0].IP.String(), port))
			},
		},
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
	return c
}

func publicIP(ip net.IP) bool {
	return !(ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() ||
		ip.IsUnspecified() || ip.IsMulticast())
}

func (c *Client) now() time.Time {
	if c.Now != nil {
		return c.Now()
	}
	return time.Now()
}

func (c *Client) checkURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return fmt.Errorf("%w: bad url", ErrDiscovery)
	}
	if u.Scheme != "https" && !(c.AllowPrivate && u.Scheme == "http") {
		return fmt.Errorf("%w: https is required", ErrDiscovery)
	}
	return nil
}

func (c *Client) getJSON(ctx context.Context, raw string, out any) error {
	if err := c.checkURL(raw); err != nil {
		return err
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, raw, nil)
	req.Header.Set("Accept", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDiscovery, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: %s returned %d", ErrDiscovery, raw, resp.StatusCode)
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(out); err != nil {
		return fmt.Errorf("%w: %v", ErrDiscovery, err)
	}
	return nil
}

// Discover loads (and caches for an hour) an issuer's OpenID configuration.
// The configuration's own `issuer` must equal the one asked for.
func (c *Client) Discover(ctx context.Context, issuer string) (*Discovery, error) {
	issuer = strings.TrimRight(issuer, "/")
	c.mu.Lock()
	if e, ok := c.disc[issuer]; ok && c.now().Before(e.exp) {
		c.mu.Unlock()
		return e.d, nil
	}
	c.mu.Unlock()

	var d Discovery
	if err := c.getJSON(ctx, issuer+"/.well-known/openid-configuration", &d); err != nil {
		return nil, err
	}
	if strings.TrimRight(d.Issuer, "/") != issuer {
		return nil, fmt.Errorf("%w: issuer mismatch", ErrDiscovery)
	}
	for _, u := range []string{d.AuthorizationEndpoint, d.TokenEndpoint, d.JWKSURI} {
		if err := c.checkURL(u); err != nil {
			return nil, err
		}
	}
	c.mu.Lock()
	c.disc[issuer] = discEntry{d: &d, exp: c.now().Add(cacheTTL)}
	c.mu.Unlock()
	return &d, nil
}

// PKCE returns a verifier and its S256 challenge.
func PKCE() (verifier, challenge string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}
	verifier = base64.RawURLEncoding.EncodeToString(b)
	h := sha256.Sum256([]byte(verifier))
	return verifier, base64.RawURLEncoding.EncodeToString(h[:]), nil
}

// Random returns n random bytes, URL-safe base64 (state, nonce, codes).
func Random(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// AuthURL builds the authorization request.
func AuthURL(d *Discovery, clientID, redirectURI, state, nonce, challenge, loginHint string) string {
	q := url.Values{
		"response_type":         {"code"},
		"client_id":             {clientID},
		"redirect_uri":          {redirectURI},
		"scope":                 {"openid email profile"},
		"state":                 {state},
		"nonce":                 {nonce},
		"code_challenge":        {challenge},
		"code_challenge_method": {"S256"},
	}
	if loginHint != "" {
		q.Set("login_hint", loginHint)
	}
	sep := "?"
	if strings.Contains(d.AuthorizationEndpoint, "?") {
		sep = "&"
	}
	return d.AuthorizationEndpoint + sep + q.Encode()
}

// Exchange redeems an authorization code and returns the identity token.
func (c *Client) Exchange(ctx context.Context, d *Discovery, clientID, secret, code, redirectURI, verifier string) (string, error) {
	form := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {redirectURI},
		"client_id":     {clientID},
		"client_secret": {secret},
		"code_verifier": {verifier},
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, d.TokenEndpoint, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrToken, err)
	}
	defer resp.Body.Close()
	var body struct {
		IDToken string `json:"id_token"`
		Error   string `json:"error"`
	}
	_ = json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&body)
	if resp.StatusCode != http.StatusOK || body.IDToken == "" {
		return "", fmt.Errorf("%w: %d %s", ErrToken, resp.StatusCode, body.Error)
	}
	return body.IDToken, nil
}

// Verify checks an identity token: RS256 signature against the issuer's
// keys, issuer, audience, expiry (with 60 s of skew), nonce, and that the
// provider vouches for the email. Anything else is rejected.
func (c *Client) Verify(ctx context.Context, d *Discovery, idToken, clientID, nonce string) (*Claims, error) {
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return nil, ErrIDToken
	}
	var hdr struct {
		Alg string `json:"alg"`
		Kid string `json:"kid"`
	}
	if err := decodePart(parts[0], &hdr); err != nil || hdr.Alg != "RS256" {
		return nil, fmt.Errorf("%w: unsupported algorithm", ErrIDToken)
	}
	key, err := c.key(ctx, d, hdr.Kid)
	if err != nil {
		return nil, err
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, ErrIDToken
	}
	sum := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	if err := rsa.VerifyPKCS1v15(key, crypto.SHA256, sum[:], sig); err != nil {
		return nil, fmt.Errorf("%w: bad signature", ErrIDToken)
	}

	var cl struct {
		Iss           string          `json:"iss"`
		Aud           json.RawMessage `json:"aud"`
		Exp           float64         `json:"exp"`
		Nonce         string          `json:"nonce"`
		Sub           string          `json:"sub"`
		Email         string          `json:"email"`
		EmailVerified any             `json:"email_verified"`
		Name          string          `json:"name"`
	}
	if err := decodePart(parts[1], &cl); err != nil {
		return nil, ErrIDToken
	}
	if strings.TrimRight(cl.Iss, "/") != strings.TrimRight(d.Issuer, "/") {
		return nil, fmt.Errorf("%w: wrong issuer", ErrIDToken)
	}
	if !audienceHas(cl.Aud, clientID) {
		return nil, fmt.Errorf("%w: wrong audience", ErrIDToken)
	}
	if c.now().Unix() > int64(cl.Exp)+60 {
		return nil, fmt.Errorf("%w: expired", ErrIDToken)
	}
	if cl.Nonce == "" || cl.Nonce != nonce {
		return nil, fmt.Errorf("%w: nonce mismatch", ErrIDToken)
	}
	if cl.Sub == "" || cl.Email == "" {
		return nil, fmt.Errorf("%w: missing subject or email", ErrIDToken)
	}
	verified := false
	switch v := cl.EmailVerified.(type) {
	case bool:
		verified = v
	case string:
		verified = v == "true"
	}
	return &Claims{Subject: cl.Sub, Email: strings.ToLower(strings.TrimSpace(cl.Email)), EmailVerified: verified, Name: cl.Name}, nil
}

func decodePart(p string, out any) error {
	b, err := base64.RawURLEncoding.DecodeString(p)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, out)
}

func audienceHas(raw json.RawMessage, want string) bool {
	var one string
	if json.Unmarshal(raw, &one) == nil {
		return one == want
	}
	var many []string
	if json.Unmarshal(raw, &many) == nil {
		for _, a := range many {
			if a == want {
				return true
			}
		}
	}
	return false
}

func (c *Client) key(ctx context.Context, d *Discovery, kid string) (*rsa.PublicKey, error) {
	get := func(force bool) (map[string]*rsa.PublicKey, error) {
		c.mu.Lock()
		e, ok := c.jwks[d.JWKSURI]
		c.mu.Unlock()
		if ok && !force && c.now().Before(e.exp) {
			return e.keys, nil
		}
		var set struct {
			Keys []struct {
				Kty string `json:"kty"`
				Kid string `json:"kid"`
				Use string `json:"use"`
				N   string `json:"n"`
				E   string `json:"e"`
			} `json:"keys"`
		}
		if err := c.getJSON(ctx, d.JWKSURI, &set); err != nil {
			return nil, err
		}
		keys := map[string]*rsa.PublicKey{}
		for _, k := range set.Keys {
			if k.Kty != "RSA" || (k.Use != "" && k.Use != "sig") {
				continue
			}
			nb, err1 := base64.RawURLEncoding.DecodeString(k.N)
			eb, err2 := base64.RawURLEncoding.DecodeString(k.E)
			if err1 != nil || err2 != nil || len(nb) < 256 { // ≥ 2048-bit
				continue
			}
			keys[k.Kid] = &rsa.PublicKey{N: new(big.Int).SetBytes(nb), E: int(new(big.Int).SetBytes(eb).Int64())}
		}
		c.mu.Lock()
		c.jwks[d.JWKSURI] = jwksEntry{keys: keys, exp: c.now().Add(cacheTTL)}
		c.mu.Unlock()
		return keys, nil
	}
	keys, err := get(false)
	if err != nil {
		return nil, err
	}
	if k, ok := keys[kid]; ok {
		return k, nil
	}
	// Unknown kid: the provider may have rotated keys. Refresh once.
	if keys, err = get(true); err != nil {
		return nil, err
	}
	if k, ok := keys[kid]; ok {
		return k, nil
	}
	if kid == "" && len(keys) == 1 {
		for _, k := range keys {
			return k, nil
		}
	}
	return nil, fmt.Errorf("%w: unknown signing key", ErrIDToken)
}
