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
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

// fakeIdP is a minimal OpenID provider: discovery, keys and a token endpoint
// that returns whatever ID token the test minted.
type fakeIdP struct {
	srv   *httptest.Server
	key   *rsa.PrivateKey
	kid   string
	token string // the id_token the token endpoint hands out
	hits  map[string]int
}

func newIdP(t *testing.T) *fakeIdP {
	t.Helper()
	k, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	p := &fakeIdP{key: k, kid: "k1", hits: map[string]int{}}
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		p.hits["discovery"]++
		_ = json.NewEncoder(w).Encode(map[string]string{
			"issuer": p.srv.URL, "authorization_endpoint": p.srv.URL + "/authorize",
			"token_endpoint": p.srv.URL + "/token", "jwks_uri": p.srv.URL + "/jwks",
		})
	})
	mux.HandleFunc("/jwks", func(w http.ResponseWriter, r *http.Request) {
		p.hits["jwks"]++
		_ = json.NewEncoder(w).Encode(map[string]any{"keys": []map[string]string{{
			"kty": "RSA", "use": "sig", "kid": p.kid, "alg": "RS256",
			"n": base64.RawURLEncoding.EncodeToString(p.key.N.Bytes()),
			"e": base64.RawURLEncoding.EncodeToString(big.NewInt(int64(p.key.E)).Bytes()),
		}}})
	})
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		p.hits["token:"+r.PostForm.Get("code_verifier")]++
		_ = json.NewEncoder(w).Encode(map[string]string{"id_token": p.token})
	})
	p.srv = httptest.NewServer(mux)
	t.Cleanup(p.srv.Close)
	return p
}

func b64(v any) string {
	b, _ := json.Marshal(v)
	return base64.RawURLEncoding.EncodeToString(b)
}

// mint signs an ID token with claims overridden by over.
func (p *fakeIdP) mint(over map[string]any) string {
	claims := map[string]any{
		"iss": p.srv.URL, "aud": "client-1", "sub": "u-1", "exp": time.Now().Add(time.Hour).Unix(),
		"nonce": "n1", "email": "Dana@Brightleaf.co", "email_verified": true, "name": "Dana",
	}
	for k, v := range over {
		if v == nil {
			delete(claims, k)
		} else {
			claims[k] = v
		}
	}
	head := b64(map[string]string{"alg": "RS256", "kid": p.kid, "typ": "JWT"})
	body := b64(claims)
	sum := sha256.Sum256([]byte(head + "." + body))
	sig, _ := rsa.SignPKCS1v15(rand.Reader, p.key, crypto.SHA256, sum[:])
	return head + "." + body + "." + base64.RawURLEncoding.EncodeToString(sig)
}

func verifyWith(t *testing.T, p *fakeIdP, tok string) (*Claims, error) {
	t.Helper()
	c := New(true)
	d, err := c.Discover(context.Background(), p.srv.URL)
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	return c.Verify(context.Background(), d, tok, "client-1", "n1")
}

func TestVerifyAcceptsAGoodTokenAndNormalizesTheEmail(t *testing.T) {
	p := newIdP(t)
	cl, err := verifyWith(t, p, p.mint(nil))
	if err != nil {
		t.Fatal(err)
	}
	if cl.Email != "dana@brightleaf.co" || !cl.EmailVerified || cl.Subject != "u-1" {
		t.Fatalf("claims = %+v", cl)
	}
}

func TestVerifyRejectsEveryWayATokenCanBeWrong(t *testing.T) {
	p := newIdP(t)
	cases := map[string]map[string]any{
		"wrong audience": {"aud": "someone-else"},
		"wrong issuer":   {"iss": "https://evil.example"},
		"expired":        {"exp": time.Now().Add(-time.Hour).Unix()},
		"wrong nonce":    {"nonce": "n2"},
		"no nonce":       {"nonce": nil},
		"no email":       {"email": nil},
	}
	for name, over := range cases {
		if _, err := verifyWith(t, p, p.mint(over)); !errors.Is(err, ErrIDToken) {
			t.Errorf("%s: want ErrIDToken, got %v", name, err)
		}
	}

	// An unverified email is a valid token but not a valid sign-in: the
	// caller must check EmailVerified, and Verify reports it faithfully.
	cl, err := verifyWith(t, p, p.mint(map[string]any{"email_verified": false}))
	if err != nil || cl.EmailVerified {
		t.Fatalf("email_verified=false: %+v %v", cl, err)
	}
	cl, _ = verifyWith(t, p, p.mint(map[string]any{"email_verified": "true"}))
	if cl == nil || !cl.EmailVerified {
		t.Fatal(`email_verified:"true" (string, as some providers send it) must count`)
	}
}

func TestVerifyRejectsForgedAndDowngradedTokens(t *testing.T) {
	p := newIdP(t)
	good := p.mint(nil)

	// Signed by a different key.
	other, _ := rsa.GenerateKey(rand.Reader, 2048)
	forger := &fakeIdP{srv: p.srv, key: other, kid: p.kid}
	if _, err := verifyWith(t, p, forger.mint(nil)); !errors.Is(err, ErrIDToken) {
		t.Errorf("forged signature: %v", err)
	}

	// Tampered payload keeps the old signature.
	parts := strings.Split(good, ".")
	evil := b64(map[string]any{"iss": p.srv.URL, "aud": "client-1", "sub": "admin", "exp": time.Now().Add(time.Hour).Unix(),
		"nonce": "n1", "email": "ceo@brightleaf.co", "email_verified": true})
	if _, err := verifyWith(t, p, parts[0]+"."+evil+"."+parts[2]); !errors.Is(err, ErrIDToken) {
		t.Errorf("tampered payload: %v", err)
	}

	// alg=none and HS256 downgrades.
	for _, alg := range []string{"none", "HS256"} {
		h := b64(map[string]string{"alg": alg, "kid": p.kid})
		if _, err := verifyWith(t, p, h+"."+parts[1]+"."+parts[2]); !errors.Is(err, ErrIDToken) {
			t.Errorf("alg %s: %v", alg, err)
		}
	}
	if _, err := verifyWith(t, p, "not.a.jwt.at.all"); !errors.Is(err, ErrIDToken) {
		t.Errorf("garbage: %v", err)
	}
}

func TestDiscoveryRequiresMatchingIssuerAndIsCached(t *testing.T) {
	p := newIdP(t)
	c := New(true)
	for i := 0; i < 3; i++ {
		if _, err := c.Discover(context.Background(), p.srv.URL); err != nil {
			t.Fatal(err)
		}
	}
	if p.hits["discovery"] != 1 {
		t.Fatalf("discovery fetched %d times, want 1", p.hits["discovery"])
	}

	// A server whose document names a different issuer is refused.
	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"issuer": "https://other.example",
			"authorization_endpoint": "http://x/a", "token_endpoint": "http://x/t", "jwks_uri": "http://x/j"})
	}))
	defer bad.Close()
	if _, err := New(true).Discover(context.Background(), bad.URL); !errors.Is(err, ErrDiscovery) {
		t.Fatalf("issuer mismatch: %v", err)
	}
}

func TestOutboundRequestsRefusePrivateAddressesByDefault(t *testing.T) {
	p := newIdP(t) // listens on 127.0.0.1
	if _, err := New(false).Discover(context.Background(), p.srv.URL); !errors.Is(err, ErrDiscovery) {
		t.Fatalf("loopback issuer must be refused outside dev mode, got %v", err)
	}
	// And plain http is refused even for a public name.
	if _, err := New(false).Discover(context.Background(), "http://idp.example.com"); !errors.Is(err, ErrDiscovery) {
		t.Fatalf("http issuer must be refused, got %v", err)
	}
}

func TestKeyRotationRefreshesOnUnknownKid(t *testing.T) {
	p := newIdP(t)
	c := New(true)
	d, _ := c.Discover(context.Background(), p.srv.URL)
	if _, err := c.Verify(context.Background(), d, p.mint(nil), "client-1", "n1"); err != nil {
		t.Fatal(err)
	}
	// The provider rotates to a new key id; the cached key set is stale.
	nk, _ := rsa.GenerateKey(rand.Reader, 2048)
	p.key, p.kid = nk, "k2"
	if _, err := c.Verify(context.Background(), d, p.mint(nil), "client-1", "n1"); err != nil {
		t.Fatalf("rotation: %v", err)
	}
	if p.hits["jwks"] != 2 {
		t.Fatalf("jwks fetched %d times, want 2", p.hits["jwks"])
	}
}

func TestAuthURLCarriesPKCEStateAndNonce(t *testing.T) {
	d := &Discovery{AuthorizationEndpoint: "https://idp.example/authorize"}
	v, ch, _ := PKCE()
	u, _ := url.Parse(AuthURL(d, "cid", "https://app/cb", "S", "N", ch, "dana@brightleaf.co"))
	q := u.Query()
	h := sha256.Sum256([]byte(v))
	if q.Get("code_challenge") != base64.RawURLEncoding.EncodeToString(h[:]) || q.Get("code_challenge_method") != "S256" ||
		q.Get("state") != "S" || q.Get("nonce") != "N" || q.Get("response_type") != "code" ||
		!strings.Contains(q.Get("scope"), "openid") || q.Get("login_hint") != "dana@brightleaf.co" {
		t.Fatalf("auth url = %s", u)
	}
}

func TestEmailDomain(t *testing.T) {
	for in, want := range map[string]string{"A@B.co": "b.co", " x@y.z ": "y.z", "nope": "", "@x.y": "", "a@": ""} {
		if got := EmailDomain(in); got != want {
			t.Errorf("EmailDomain(%q) = %q, want %q", in, got, want)
		}
	}
}
