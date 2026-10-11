package tests

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/thescaffold/gox-apps/libs/identity/app/auth"
	"github.com/thescaffold/gox-apps/libs/identity/app/user"
	identityucw "github.com/thescaffold/gox-apps/libs/identity/app/userclientworkspace"
	identityworkspace "github.com/thescaffold/gox-apps/libs/identity/app/workspace"
	"github.com/thescaffold/gox-apps/libs/identity/pkg/sso"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/security"
)

// An OpenID provider the tests control: it hands out the ID token the test
// minted for whichever nonce the sign-in used.
type testIdP struct {
	srv  *httptest.Server
	key  *rsa.PrivateKey
	mu   sync.Mutex
	next map[string]any // claim overrides for the next token
}

func newTestIdP(t *testing.T) *testIdP {
	t.Helper()
	k, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	p := &testIdP{key: k}
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"issuer": p.srv.URL,
			"authorization_endpoint": p.srv.URL + "/authorize", "token_endpoint": p.srv.URL + "/token", "jwks_uri": p.srv.URL + "/jwks"})
	})
	mux.HandleFunc("/jwks", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"keys": []map[string]string{{"kty": "RSA", "use": "sig", "kid": "k",
			"n": base64.RawURLEncoding.EncodeToString(k.N.Bytes()), "e": base64.RawURLEncoding.EncodeToString(big.NewInt(int64(k.E)).Bytes())}}})
	})
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		// The code the test passed to the callback is "code:<nonce>".
		nonce := strings.TrimPrefix(r.PostForm.Get("code"), "code:")
		_ = json.NewEncoder(w).Encode(map[string]string{"id_token": p.mint(nonce)})
	})
	p.srv = httptest.NewServer(mux)
	t.Cleanup(p.srv.Close)
	return p
}

func (p *testIdP) mint(nonce string) string {
	p.mu.Lock()
	over := p.next
	p.mu.Unlock()
	claims := map[string]any{"iss": p.srv.URL, "aud": "oidc-client", "sub": "sub-1", "exp": time.Now().Add(time.Hour).Unix(),
		"nonce": nonce, "email_verified": true, "name": "Test Person"}
	for k, v := range over {
		claims[k] = v
	}
	seg := func(v any) string { b, _ := json.Marshal(v); return base64.RawURLEncoding.EncodeToString(b) }
	h, b := seg(map[string]string{"alg": "RS256", "kid": "k"}), seg(claims)
	sum := sha256.Sum256([]byte(h + "." + b))
	sig, _ := rsa.SignPKCS1v15(rand.Reader, p.key, crypto.SHA256, sum[:])
	return h + "." + b + "." + base64.RawURLEncoding.EncodeToString(sig)
}

func (p *testIdP) setClaims(c map[string]any) { p.mu.Lock(); p.next = c; p.mu.Unlock() }

// fakeDirectory is the host's side of invited-only membership.
type fakeDirectory struct {
	mu      sync.Mutex
	domains map[string]bool // "ws|domain"
	invited map[string]bool // "ws|email"
	joined  []string
}

func (d *fakeDirectory) OwnsDomain(_ context.Context, ws, dom string) (bool, error) {
	return d.domains[ws+"|"+dom], nil
}
func (d *fakeDirectory) Invited(_ context.Context, ws, email string) (bool, error) {
	return d.invited[ws+"|"+email], nil
}
func (d *fakeDirectory) Joined(_ context.Context, ws, userID, email string) error {
	d.mu.Lock()
	d.joined = append(d.joined, ws+"|"+userID+"|"+email)
	d.mu.Unlock()
	return nil
}

type ssoWorld struct {
	r      *rig
	idp    *testIdP
	dir    *fakeDirectory
	admin  string // identity user id (owner)
	ws     string
	domain string
}

func ssoRig(t *testing.T) *ssoWorld {
	t.Helper()
	r := newRig(t)
	t.Setenv("IDENTITY_TOKEN_KEY", "0123456789abcdef0123456789abcdef")
	t.Setenv("BASE_URL", "https://app.test")
	idp := newTestIdP(t)
	u, ws := r.person("admin")
	domain := "co-" + uuid.NewString()[:8] + ".test"
	dir := &fakeDirectory{domains: map[string]bool{ws + "|" + domain: true}, invited: map[string]bool{}}
	sso.SetDirectory(dir)
	t.Cleanup(func() {
		sso.SetDirectory(nil)
		r.db.Exec(`DELETE FROM "IdentitySsoConnections" WHERE workspace_id = ?`, ws)
	})
	return &ssoWorld{r: r, idp: idp, dir: dir, admin: u.Id, ws: ws, domain: domain}
}

func (w *ssoWorld) ctx(userID string) (c ntxctx.NTXContext) {
	c.UserID, c.ClientID, c.WorkspaceID = userID, w.r.client, w.ws
	return c
}

func (w *ssoWorld) connect(enforce bool) reply {
	w.r.t.Helper()
	return decode(w.r.t, w.r.ctl.SsoPutConnection(&auth.SsoConnectionDto{
		Ctx: w.ctx(w.admin), Issuer: w.idp.srv.URL, OIDCClientID: "oidc-client", ClientSecret: "s3cret",
		Domains: []string{w.domain}, Enforce: enforce,
	}))
}

// signIn drives start → callback and returns the callback's Location.
func (w *ssoWorld) signIn(email string) (location string) {
	w.r.t.Helper()
	start := decode(w.r.t, w.r.ctl.SsoStart(&auth.SsoStartDto{Email: email, ClientId: w.r.client, ReturnTo: "/sso/complete"}))
	if start.code != 200 {
		w.r.t.Fatalf("start: %d %v", start.code, start.body)
	}
	u, _ := url.Parse(start.data()["url"].(string))
	q := u.Query()
	if q.Get("code_challenge") == "" || q.Get("state") == "" || q.Get("login_hint") != strings.ToLower(email) {
		w.r.t.Fatalf("authorize url = %s", u)
	}
	out := w.r.ctl.SsoCallback(&auth.SsoCallbackDto{Code: "code:" + q.Get("nonce"), State: q.Get("state")})
	return out.Headers()["Location"]
}

func TestSSO_InvitedPersonSignsInAndJoins(t *testing.T) {
	w := ssoRig(t)
	if rep := w.connect(false); rep.code != 200 || rep.data()["hasSecret"] != true {
		t.Fatalf("connect: %d %v", rep.code, rep.body)
	}
	email := "dana@" + w.domain
	w.dir.invited[w.ws+"|"+email] = true
	w.idp.setClaims(map[string]any{"email": strings.ToUpper(email[:1]) + email[1:]})

	loc := w.signIn(email)
	if !strings.HasPrefix(loc, "https://app.test/sso/complete?code=") {
		t.Fatalf("callback went to %q", loc)
	}
	code := strings.TrimPrefix(loc, "https://app.test/sso/complete?code=")
	ex := decode(t, w.r.ctl.SsoExchange(&auth.SsoExchangeDto{Code: code}))
	if ex.code != 200 || ex.data()["accessToken"] == "" || ex.data()["isNew"] != true {
		t.Fatalf("exchange: %d %v", ex.code, ex.body)
	}
	if got := ex.data()["user"].(map[string]any)["email"]; got != email {
		t.Fatalf("user email = %v", got)
	}
	if len(w.dir.joined) != 1 || !strings.HasSuffix(w.dir.joined[0], "|"+email) {
		t.Fatalf("Joined calls = %v", w.dir.joined)
	}
	// The code is single use.
	again := decode(t, w.r.ctl.SsoExchange(&auth.SsoExchangeDto{Code: code}))
	if again.code != 401 {
		t.Fatalf("reused code: %d", again.code)
	}
	// A second sign-in matches the same account.
	w.idp.setClaims(map[string]any{"email": email})
	code2 := strings.TrimPrefix(w.signIn(email), "https://app.test/sso/complete?code=")
	second := decode(t, w.r.ctl.SsoExchange(&auth.SsoExchangeDto{Code: code2}))
	if second.data()["isNew"] != false || second.data()["user"].(map[string]any)["id"] != ex.data()["user"].(map[string]any)["id"] {
		t.Fatalf("second sign-in: %v", second.body)
	}
}

func TestSSO_RefusesWhatIsNotAnInvitedVerifiedAddressOfTheCompany(t *testing.T) {
	w := ssoRig(t)
	w.connect(false)
	email := "eve@" + w.domain

	cases := []struct {
		name   string
		claims map[string]any
		want   string
	}{
		{"not invited", map[string]any{"email": email}, "not_invited"},
		{"unverified email", map[string]any{"email": email, "email_verified": false}, "email_unverified"},
		{"another company's address", map[string]any{"email": "eve@other.example"}, "wrong_domain"},
		{"wrong audience", map[string]any{"email": email, "aud": "someone-else"}, "invalid_token"},
	}
	for _, c := range cases {
		if c.name == "unverified email" || c.name == "wrong audience" {
			w.dir.invited[w.ws+"|"+email] = true // otherwise it would fail earlier for the wrong reason
		}
		w.idp.setClaims(c.claims)
		loc := w.signIn(email)
		if !strings.Contains(loc, "error="+c.want) {
			t.Errorf("%s: redirected to %q, want error=%s", c.name, loc, c.want)
		}
	}
	var n int64
	w.r.db.Raw(`SELECT COUNT(*) FROM "IdentityUsers" WHERE email = ?`, email).Scan(&n)
	if n != 0 || len(w.dir.joined) != 0 {
		t.Fatalf("nothing may be created for a refused sign-in: users=%d joined=%v", n, w.dir.joined)
	}
}

func TestSSO_StateIsSingleUse(t *testing.T) {
	w := ssoRig(t)
	w.connect(false)
	start := decode(t, w.r.ctl.SsoStart(&auth.SsoStartDto{Email: "a@" + w.domain, ClientId: w.r.client}))
	u, _ := url.Parse(start.data()["url"].(string))
	q := u.Query()
	w.idp.setClaims(map[string]any{"email": "a@" + w.domain})
	first := w.r.ctl.SsoCallback(&auth.SsoCallbackDto{Code: "code:" + q.Get("nonce"), State: q.Get("state")})
	second := w.r.ctl.SsoCallback(&auth.SsoCallbackDto{Code: "code:" + q.Get("nonce"), State: q.Get("state")})
	if !strings.Contains(first.Headers()["Location"], "error=") { // not invited — but the state was consumed
		t.Fatalf("first: %v", first.Headers())
	}
	if !strings.Contains(second.Headers()["Location"], "error=expired") {
		t.Fatalf("a replayed state must be refused: %v", second.Headers())
	}
	if r := decode(t, w.r.ctl.SsoStart(&auth.SsoStartDto{Email: "a@elsewhere.test", ClientId: w.r.client})); r.code != 404 {
		t.Fatalf("an email with no SSO: %d", r.code)
	}
}

func TestSSO_OnlyAdminsOfAVerifiedDomainCanConfigure(t *testing.T) {
	w := ssoRig(t)
	stranger, _ := w.r.person("stranger") // owns a different workspace, not this one
	rep := decode(t, w.r.ctl.SsoPutConnection(&auth.SsoConnectionDto{
		Ctx: w.ctx(stranger.Id), Issuer: w.idp.srv.URL, OIDCClientID: "c", ClientSecret: "s", Domains: []string{w.domain},
	}))
	if rep.code != 403 {
		t.Fatalf("a non-member: %d", rep.code)
	}
	rep = decode(t, w.r.ctl.SsoPutConnection(&auth.SsoConnectionDto{
		Ctx: w.ctx(w.admin), Issuer: w.idp.srv.URL, OIDCClientID: "c", ClientSecret: "s", Domains: []string{"gmail.com"},
	}))
	if rep.code != 400 || !strings.Contains(rep.body["message"].(string), "not a verified domain") {
		t.Fatalf("an unowned domain: %d %v", rep.code, rep.body)
	}
	rep = decode(t, w.r.ctl.SsoPutConnection(&auth.SsoConnectionDto{
		Ctx: w.ctx(w.admin), Issuer: "http://127.0.0.1:1", OIDCClientID: "c", ClientSecret: "s", Domains: []string{w.domain},
	}))
	if rep.code != 400 {
		t.Fatalf("an unreachable provider is a clear error, got %d", rep.code)
	}
	// The secret is never read back; updating without one keeps the stored secret.
	w.connect(false)
	rep = decode(t, w.r.ctl.SsoPutConnection(&auth.SsoConnectionDto{
		Ctx: w.ctx(w.admin), Issuer: w.idp.srv.URL, OIDCClientID: "oidc-client", Domains: []string{w.domain}, Enforce: true,
	}))
	if rep.code != 200 || rep.data()["enforce"] != true {
		t.Fatalf("update: %d %v", rep.code, rep.body)
	}
	var sealed string
	w.r.db.Raw(`SELECT client_secret_sealed FROM "IdentitySsoConnections" WHERE workspace_id = ?`, w.ws).Scan(&sealed)
	if sealed == "" || strings.Contains(sealed, "s3cret") {
		t.Fatalf("secret must be stored sealed, got %q", sealed)
	}
	if got := decode(t, w.r.ctl.SsoGetConnection(&auth.SsoWorkspaceDto{Ctx: w.ctx(w.admin)})); strings.Contains(string(mustJSON(got.body)), "s3cret") {
		t.Fatal("the secret must never be returned")
	}
}

func TestSSO_EnforcementClosesThePasswordAndCodePaths(t *testing.T) {
	w := ssoRig(t)
	email := "frank-" + uuid.NewString()[:6] + "@" + w.domain
	// A person with a password, before the company turns enforcement on.
	personAt(w.r, email)
	if rep := decode(t, w.r.ctl.Login(&auth.LoginDto{Ref: email, Secret: testPassword, ClientId: w.r.client})); rep.code != 200 {
		t.Fatalf("before enforcement: %d", rep.code)
	}
	w.connect(true)
	for name, rep := range map[string]reply{
		"password": decode(t, w.r.ctl.Login(&auth.LoginDto{Ref: email, Secret: testPassword, ClientId: w.r.client})),
		"code req": decode(t, w.r.ctl.OtpRequest(&auth.OtpRequestDto{Ref: email, ClientId: w.r.client})),
		"code use": decode(t, w.r.ctl.OtpVerify(&auth.OtpVerifyDto{Ref: email, Code: "424242", ClientId: w.r.client})),
		"register": decode(t, w.r.ctl.Initiate(&auth.InitiateDto{Ref: "new@" + w.domain, Type: "email", ClientId: w.r.client})),
	} {
		if rep.code != 403 {
			t.Errorf("%s: %d, want 403", name, rep.code)
		}
	}
	// Other companies and consumer addresses are untouched.
	gina, _ := w.r.person("gina")
	if rep := decode(t, w.r.ctl.Login(&auth.LoginDto{Ref: gina.Email, Secret: testPassword, ClientId: w.r.client})); rep.code != 200 {
		t.Fatalf("an address outside the company: %d", rep.code)
	}
	// Removing the connection lifts enforcement.
	decode(t, w.r.ctl.SsoDeleteConnection(&auth.SsoWorkspaceDto{Ctx: w.ctx(w.admin)}))
	if rep := decode(t, w.r.ctl.Login(&auth.LoginDto{Ref: email, Secret: testPassword, ClientId: w.r.client})); rep.code != 200 {
		t.Fatalf("after removal: %d", rep.code)
	}
}

func mustJSON(v any) []byte { b, _ := json.Marshal(v); return b }

// personAt creates a user with a password, a workspace and the given email.
func personAt(r *rig, email string) {
	r.t.Helper()
	hashed, err := security.Hash(testPassword)
	if err != nil {
		r.t.Fatal(err)
	}
	typ := "email"
	u := &user.User{Name: "P", Email: email, Ref: &email, Type: &typ, Secret: &hashed}
	u.Id = uuid.NewString()
	if err := r.db.Create(u).Error; err != nil {
		r.t.Fatal(err)
	}
	ws := &identityworkspace.Workspace{ClientId: r.client, UserId: u.Id, Name: "w", Slug: "ws-" + uuid.NewString()[:8]}
	ws.Id = uuid.NewString()
	if err := r.db.Create(ws).Error; err != nil {
		r.t.Fatal(err)
	}
	owner := "owner"
	link := &identityucw.UserClientWorkspace{UserId: u.Id, ClientId: r.client, WorkspaceId: ws.Id, Status: &owner}
	link.Id = uuid.NewString()
	if err := r.db.Create(link).Error; err != nil {
		r.t.Fatal(err)
	}
}

func TestMyWorkspaces_AreTheCallersOwnAndCreationIsBoundToTheToken(t *testing.T) {
	r := newRig(t)
	a, _ := r.person("alice")
	b, _ := r.person("bob")
	ctxA := ntxctx.NTXContext{UserID: a.Id, ClientID: r.client}

	created := decode(t, r.ctl.CreateMyWorkspace(&auth.CreateMyWorkspaceDto{Ctx: ctxA, Name: "  Brightleaf  "}))
	if created.code != 201 || created.data()["name"] != "Brightleaf" || created.data()["role"] != "owner" {
		t.Fatalf("create: %d %v", created.code, created.body)
	}
	list := decode(t, r.ctl.MyWorkspaces(&auth.MyWorkspacesDto{Ctx: ctxA}))
	names := map[string]bool{}
	for _, w := range list.list() {
		names[w.(map[string]any)["name"].(string)] = true
	}
	if !names["Brightleaf"] || len(names) != 2 {
		t.Fatalf("alice's workspaces = %v", names)
	}
	// Bob sees none of Alice's, and the new workspace is switchable only by its owner.
	other := decode(t, r.ctl.MyWorkspaces(&auth.MyWorkspacesDto{Ctx: ntxctx.NTXContext{UserID: b.Id, ClientID: r.client}}))
	for _, w := range other.list() {
		if w.(map[string]any)["name"] == "Brightleaf" {
			t.Fatal("another person's workspace leaked")
		}
	}
	id := created.data()["id"].(string)
	if rep := decode(t, r.ctl.SwitchWorkspace(&auth.SwitchWorkspaceDto{Ctx: ntxctx.NTXContext{UserID: b.Id, ClientID: r.client}, WorkspaceId: id})); rep.code != 401 {
		t.Fatalf("bob switching into alice's workspace: %d", rep.code)
	}
	if rep := decode(t, r.ctl.SwitchWorkspace(&auth.SwitchWorkspaceDto{Ctx: ctxA, WorkspaceId: id})); rep.code != 200 {
		t.Fatalf("alice switching into her new workspace: %d", rep.code)
	}
	if rep := decode(t, r.ctl.CreateMyWorkspace(&auth.CreateMyWorkspaceDto{Ctx: ctxA, Name: ""})); rep.code != 400 {
		t.Fatalf("empty name: %d", rep.code)
	}
}
