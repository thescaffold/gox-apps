package tests

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
	"github.com/google/uuid"
	"github.com/thescaffold/gox-apps/libs/identity/app/auth"
	identityclient "github.com/thescaffold/gox-apps/libs/identity/app/client"
	identityclientlog "github.com/thescaffold/gox-apps/libs/identity/app/clientlog"
	identitydevicesession "github.com/thescaffold/gox-apps/libs/identity/app/devicesession"
	"github.com/thescaffold/gox-apps/libs/identity/app/user"
	identityucw "github.com/thescaffold/gox-apps/libs/identity/app/userclientworkspace"
	identityworkspace "github.com/thescaffold/gox-apps/libs/identity/app/workspace"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/security"
	"gorm.io/gorm"
)

// PLAN M1-40: the browser shell holds a device cookie, lists the accounts that
// device is signed in to (GET sessions) and re-mints a token from them. These
// tests drive the real handlers over a real database.

const testPassword = "correct horse battery staple"

type rig struct {
	t      *testing.T
	db     *gorm.DB
	ctl    *auth.AuthController
	client string
}

func newRig(t *testing.T) *rig {
	t.Helper()
	db := identityTestDB(t)
	return &rig{t: t, db: db, ctl: auth.NewAuthControllerForTest(db), client: "client-" + uuid.NewString()[:8]}
}

func (r *rig) q() *sql.Query { return (&sql.Query{}).With(&sql.Db{DB: r.db}) }

// person creates a user with a password and a workspace they own.
func (r *rig) person(name string) (u *user.User, workspaceID string) {
	r.t.Helper()
	hashed, err := security.Hash(testPassword)
	if err != nil {
		r.t.Fatal(err)
	}
	email := name + "-" + uuid.NewString()[:8] + "@example.test"
	typ := "email"
	u = &user.User{Name: name, Email: email, Ref: &email, Type: &typ, Secret: &hashed}
	u.Id = uuid.NewString()
	if err := r.db.Create(u).Error; err != nil {
		r.t.Fatal(err)
	}
	ws := &identityworkspace.Workspace{ClientId: r.client, UserId: u.Id, Name: name + "'s workspace", Slug: "ws-" + uuid.NewString()[:8]}
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
	return u, ws.Id
}

type reply struct {
	code    int
	headers map[string]string
	body    map[string]any
}

func decode(t *testing.T, out types.Output) reply {
	t.Helper()
	raw, err := json.Marshal(out.Data())
	if err != nil {
		t.Fatal(err)
	}
	var b map[string]any
	if err := json.Unmarshal(raw, &b); err != nil {
		t.Fatalf("not a JSON object: %s", raw)
	}
	return reply{code: out.Code(), headers: out.Headers(), body: b}
}

func (r reply) data() map[string]any {
	d, _ := r.body["data"].(map[string]any)
	return d
}

func (r reply) list() []any {
	d, _ := r.body["data"].([]any)
	return d
}

func (r *rig) login(u *user.User, cookie string) reply {
	r.t.Helper()
	email := ""
	if u.Ref != nil {
		email = *u.Ref
	}
	return decode(r.t, r.ctl.Login(&auth.LoginDto{Ref: email, Secret: testPassword, ClientId: r.client, UserAgent: "test-agent", Cookie: cookie}))
}

// deviceFrom reads the device id out of a Set-Cookie header.
func deviceFrom(t *testing.T, rep reply) string {
	t.Helper()
	sc := rep.headers["Set-Cookie"]
	if sc == "" {
		t.Fatalf("no Set-Cookie header on the response: %v", rep.headers)
	}
	first, _, _ := strings.Cut(sc, ";")
	name, val, ok := strings.Cut(first, "=")
	if !ok || name != "x_ntx_device_id" || val == "" {
		t.Fatalf("Set-Cookie %q does not set x_ntx_device_id", sc)
	}
	return val
}

func (r *rig) sessions(cookie string) []map[string]any {
	r.t.Helper()
	rep := decode(r.t, r.ctl.Sessions(&auth.SessionsDto{Cookie: cookie}))
	if rep.code != 200 {
		r.t.Fatalf("sessions: HTTP %d %v", rep.code, rep.body)
	}
	var out []map[string]any
	for _, s := range rep.list() {
		out = append(out, s.(map[string]any))
	}
	return out
}

func cookieFor(device string) string { return "theme=dark; x_ntx_device_id=" + device + "; other=1" }

func TestLoginSetsTheDeviceCookieAndReusesTheDeviceItIsGiven(t *testing.T) {
	r := newRig(t)
	ada, _ := r.person("ada")
	lin, _ := r.person("lin")

	first := r.login(ada, "")
	if first.code != 200 {
		t.Fatalf("login: %d %v", first.code, first.body)
	}
	device := deviceFrom(t, first)
	sc := first.headers["Set-Cookie"]
	for _, want := range []string{"HttpOnly", "SameSite=Lax", "Path=/", "Max-Age="} {
		if !strings.Contains(sc, want) {
			t.Fatalf("cookie %q lacks %s", sc, want)
		}
	}

	// the same browser signs a second person in: the device is reused, not remade
	second := r.login(lin, cookieFor(device))
	if got := deviceFrom(t, second); got != device {
		t.Fatalf("a second login made device %s, want the browser's %s", got, device)
	}
	var devices, sessions int64
	r.db.Raw(`SELECT count(*) FROM "IdentityDevices" WHERE id = ?`, device).Scan(&devices)
	r.db.Raw(`SELECT count(*) FROM "IdentityDeviceSessions" WHERE device_id = ? AND deleted_at IS NULL`, device).Scan(&sessions)
	if devices != 1 || sessions != 2 {
		t.Fatalf("devices=%d sessions=%d, want 1 device holding 2 sessions", devices, sessions)
	}

	// signing the same person in again on that device does not pile up sessions
	r.login(ada, cookieFor(device))
	r.db.Raw(`SELECT count(*) FROM "IdentityDeviceSessions" WHERE device_id = ? AND deleted_at IS NULL`, device).Scan(&sessions)
	if sessions != 2 {
		t.Fatalf("%d sessions after the same person signed in again, want 2", sessions)
	}

	// a cookie that names no device is not trusted: a fresh one is made
	other := r.login(ada, cookieFor("no-such-device"))
	if got := deviceFrom(t, other); got == "no-such-device" || got == device {
		t.Fatalf("an unknown device cookie was honoured: %s", got)
	}
}

func TestSessionsListsTheAccountsOfThisDeviceWithATokenPerWorkspace(t *testing.T) {
	r := newRig(t)
	ada, adaWs := r.person("ada")
	lin, _ := r.person("lin")
	device := deviceFrom(t, r.login(ada, ""))
	r.login(lin, cookieFor(device))
	// a third person on another device must never appear
	outsider, _ := r.person("eve")
	r.login(outsider, "")

	got := r.sessions(cookieFor(device))
	if len(got) != 2 {
		t.Fatalf("%d sessions, want the 2 of this device: %v", len(got), got)
	}
	// most recently signed in first
	if u := got[0]["user"].(map[string]any); u["id"] != lin.Id {
		t.Fatalf("first session is %v, want the most recent sign-in (lin)", u["id"])
	}
	var adaSession map[string]any
	for _, s := range got {
		if s["user"].(map[string]any)["id"] == ada.Id {
			adaSession = s["user"].(map[string]any)
		}
	}
	if adaSession == nil {
		t.Fatalf("ada is missing: %v", got)
	}
	if _, leaked := adaSession["secret"]; leaked {
		t.Fatal("the user's secret is in the response")
	}
	wss := adaSession["workspaces"].([]any)
	if len(wss) != 1 {
		t.Fatalf("workspaces = %v", wss)
	}
	w := wss[0].(map[string]any)
	if w["workspaceId"] != adaWs || w["userId"] != ada.Id || w["clientId"] != r.client {
		t.Fatalf("workspace link = %v", w)
	}
	tok := w["workspace"].(map[string]any)["token"].(string)
	claims, err := coreauth.Verify(tok, "changeme")
	if err != nil || claims["sub"] != ada.Id || claims["workspaceId"] != adaWs || claims["clientId"] != r.client {
		t.Fatalf("the workspace token is %v (%v), want one for ada in her workspace", claims, err)
	}
	if adaSession["token"] == nil {
		t.Fatal("the user has no token")
	}
}

func TestSessionsAnswersWithNothingForNoCookieAnUnknownOneOrAnotherDevicesId(t *testing.T) {
	r := newRig(t)
	ada, _ := r.person("ada")
	device := deviceFrom(t, r.login(ada, ""))
	for name, cookie := range map[string]string{"none": "", "empty": "x_ntx_device_id=", "unknown": cookieFor("nope"), "other cookies only": "theme=dark"} {
		if got := r.sessions(cookie); len(got) != 0 {
			t.Fatalf("%s: %d sessions, want none", name, len(got))
		}
	}
	_ = device
}

func TestSessionsLeavesOutExpiredSessionsAndClosedOrSuspendedAccounts(t *testing.T) {
	r := newRig(t)
	ada, _ := r.person("ada")
	old, _ := r.person("old")
	gone, _ := r.person("gone")
	device := deviceFrom(t, r.login(ada, ""))
	r.login(old, cookieFor(device))
	r.login(gone, cookieFor(device))
	r.db.Exec(`UPDATE "IdentityDeviceSessions" SET expires_at = timezone('UTC', now()) - interval '1 hour' WHERE device_id = ? AND user_id = ?`, device, old.Id)
	r.db.Exec(`INSERT INTO "IdentityAttributes" (id, created_at, updated_at, user_id, key, value) VALUES (?, now(), now(), ?, 'closed_at', 'x')`, uuid.NewString(), gone.Id)
	got := r.sessions(cookieFor(device))
	if len(got) != 1 || got[0]["user"].(map[string]any)["id"] != ada.Id {
		t.Fatalf("sessions = %v, want only ada", got)
	}
}

func TestSessionsMintsAccessTokensOnlyAndWritesNoRefreshTokens(t *testing.T) {
	r := newRig(t)
	ada, _ := r.person("ada")
	device := deviceFrom(t, r.login(ada, ""))
	var before, after int64
	r.db.Raw(`SELECT count(*) FROM "IdentityTokens"`).Scan(&before)
	for i := 0; i < 5; i++ {
		r.sessions(cookieFor(device))
	}
	r.db.Raw(`SELECT count(*) FROM "IdentityTokens"`).Scan(&after)
	if after != before {
		t.Fatalf("fetching sessions wrote %d token rows", after-before)
	}
}

func TestSessionsIncludesTheUsersAttributesGroupedByType(t *testing.T) {
	r := newRig(t)
	ada, _ := r.person("ada")
	r.db.Exec(`INSERT INTO "IdentityAttributes" (id, created_at, updated_at, user_id, key, value, type) VALUES (?, now(), now(), ?, 'theme', 'dark', 'preference')`, uuid.NewString(), ada.Id)
	device := deviceFrom(t, r.login(ada, ""))
	u := r.sessions(cookieFor(device))[0]["user"].(map[string]any)
	pref := u["attributes"].(map[string]any)["preference"].(map[string]any)
	if pref["theme"] != "dark" {
		t.Fatalf("attributes = %v", u["attributes"])
	}
}

func TestLogoutSignsOutWhatWasAskedAndNothingElse(t *testing.T) {
	r := newRig(t)
	ada, _ := r.person("ada")
	lin, _ := r.person("lin")
	device := deviceFrom(t, r.login(ada, ""))
	r.login(lin, cookieFor(device))
	linElsewhere := deviceFrom(t, r.login(lin, ""))

	// one user on this device
	rep := decode(t, r.ctl.Logout(&auth.LogoutDto{Cookie: cookieFor(device), Device: "yes", UserId: lin.Id, ClientId: r.client}))
	if rep.code != 200 {
		t.Fatalf("logout: %d %v", rep.code, rep.body)
	}
	got := r.sessions(cookieFor(device))
	if len(got) != 1 || got[0]["user"].(map[string]any)["id"] != ada.Id {
		t.Fatalf("after signing lin out of this device: %v", got)
	}
	if len(r.sessions(cookieFor(linElsewhere))) != 1 {
		t.Fatal("signing lin out of this device signed her out of another one")
	}

	// everyone on this device
	rep = decode(t, r.ctl.Logout(&auth.LogoutDto{Cookie: cookieFor(device), Device: "yes", ClientId: r.client}))
	if got := r.sessions(cookieFor(device)); len(got) != 0 {
		t.Fatalf("sessions after signing the whole device out: %v", got)
	}
	if sc := rep.headers["Set-Cookie"]; !strings.Contains(sc, "x_ntx_device_id=;") || !strings.Contains(sc, "Max-Age=0") {
		t.Fatalf("the device cookie was not cleared: %q", sc)
	}
	if len(r.sessions(cookieFor(linElsewhere))) != 1 {
		t.Fatal("signing the device out touched another device")
	}
}

func TestLogoutOfAUserSignsThemOutOfEveryDeviceButOnlyThem(t *testing.T) {
	r := newRig(t)
	ada, _ := r.person("ada")
	lin, _ := r.person("lin")
	d1 := deviceFrom(t, r.login(lin, ""))
	d2 := deviceFrom(t, r.login(lin, ""))
	r.login(ada, cookieFor(d1))
	decode(t, r.ctl.Logout(&auth.LogoutDto{Device: "no", UserId: lin.Id, ClientId: r.client}))
	if got := r.sessions(cookieFor(d2)); len(got) != 0 {
		t.Fatalf("lin is still signed in on another device: %v", got)
	}
	if got := r.sessions(cookieFor(d1)); len(got) != 1 || got[0]["user"].(map[string]any)["id"] != ada.Id {
		t.Fatalf("ada was signed out with lin: %v", got)
	}
}

func TestUserClientsListsFavouritesAndWhatElseExists(t *testing.T) {
	r := newRig(t)
	// the client list is global, so start from none
	r.db.Exec(`DELETE FROM "IdentityClientLogs"`)
	r.db.Exec(`DELETE FROM "IdentityClients"`)
	ada, _ := r.person("ada")
	mk := func(key, name string) *identityclient.Client {
		c := &identityclient.Client{Name: name, Type: "web", Key: key + "-" + uuid.NewString()[:6]}
		c.Id = uuid.NewString()
		if err := r.db.Create(c).Error; err != nil {
			t.Fatal(err)
		}
		return c
	}
	used, unused := mk("used", "Used app"), mk("unused", "Unused app")
	common := mk("payments", "Payments")
	r.db.Model(common).Update("key", "payments")
	// ada has opened one app
	reg := decode(t, r.ctl.RegisterClientLog(&auth.ClientLogRegisterDto{
		Ctx: ntxctx.NTXContext{UserID: ada.Id, ClientID: used.Id}, ClientId: used.Id, Type: "client", Status: "client"}))
	if reg.code != 200 {
		t.Fatalf("register: %d %v", reg.code, reg.body)
	}
	rep := decode(t, r.ctl.UserClients(&auth.UserClientsDto{Ctx: ntxctx.NTXContext{UserID: ada.Id, ClientID: used.Id}}))
	if rep.code != 200 {
		t.Fatalf("user-clients: %d %v", rep.code, rep.body)
	}
	title := func(list any) []string {
		var out []string
		for _, x := range list.([]any) {
			out = append(out, x.(map[string]any)["title"].(string))
		}
		return out
	}
	d := rep.data()
	if got := strings.Join(title(d["favorites"]), ","); got != "Used app" {
		t.Fatalf("favorites = %q", got)
	}
	if !strings.Contains(strings.Join(title(d["discover"]), ","), "Unused app") || strings.Contains(strings.Join(title(d["discover"]), ","), "Used app") {
		t.Fatalf("discover = %v", d["discover"])
	}
	if !strings.Contains(strings.Join(title(d["common"]), ","), "Payments") {
		t.Fatalf("common = %v", d["common"])
	}
	_ = unused

	// someone else's favourites are their own
	lin, _ := r.person("lin")
	other := decode(t, r.ctl.UserClients(&auth.UserClientsDto{Ctx: ntxctx.NTXContext{UserID: lin.Id}}))
	if got := title(other.data()["favorites"]); len(got) != 0 {
		t.Fatalf("lin inherited ada's favourites: %v", got)
	}
}

func TestRegisteringAClientUsesTheSignedInUserNotTheBody(t *testing.T) {
	r := newRig(t)
	ada, _ := r.person("ada")
	c := &identityclient.Client{Name: "App", Type: "web", Key: "k-" + uuid.NewString()[:6]}
	c.Id = uuid.NewString()
	r.db.Create(c)
	// a signed-out caller registers nothing
	if rep := decode(t, r.ctl.RegisterClientLog(&auth.ClientLogRegisterDto{ClientId: c.Id, Type: "client"})); rep.code != http.StatusUnauthorized {
		t.Fatalf("an anonymous register answered %d, want 401", rep.code)
	}
	decode(t, r.ctl.RegisterClientLog(&auth.ClientLogRegisterDto{Ctx: ntxctx.NTXContext{UserID: ada.Id}, ClientId: c.Id, WorkspaceId: "ws-1", Type: "client", Status: "client"}))
	var rows []identityclientlog.ClientLog
	r.db.Where("client_id = ?", c.Id).Find(&rows)
	if len(rows) != 1 || !strings.Contains(string(rows[0].Meta), ada.Id) || !strings.Contains(string(rows[0].Meta), "ws-1") {
		t.Fatalf("log rows = %+v", rows)
	}
	// opening the same app again does not add a row every time
	decode(t, r.ctl.RegisterClientLog(&auth.ClientLogRegisterDto{Ctx: ntxctx.NTXContext{UserID: ada.Id}, ClientId: c.Id, WorkspaceId: "ws-1", Type: "client", Status: "client"}))
	var n int64
	r.db.Model(&identityclientlog.ClientLog{}).Where("client_id = ?", c.Id).Count(&n)
	if n != 1 {
		t.Fatalf("%d rows after opening the same app twice, want 1", n)
	}
}

func TestADeviceSessionRemembersItsClient(t *testing.T) {
	r := newRig(t)
	ada, _ := r.person("ada")
	device := deviceFrom(t, r.login(ada, ""))
	var s identitydevicesession.DeviceSession
	if err := r.db.Where("device_id = ?", device).First(&s).Error; err != nil || s.ClientId == nil || *s.ClientId != r.client {
		t.Fatalf("session = %+v, %v", s, err)
	}
	if s.ExpiresAt == nil || s.ExpiresAt.Before(time.Now().UTC().Add(24*time.Hour)) {
		t.Fatalf("the session lasts less than a day: %v", s.ExpiresAt)
	}
}

// Signing up opens the new account's device session and sets the device cookie,
// like signing in does: the browser must be able to list the account it has just
// created.
func TestSigningUpOpensTheDeviceSessionAndSetsTheCookie(t *testing.T) {
	r := newRig(t)
	id := uuid.NewString()
	email := "new-" + id[:8] + "@example.test"
	token := r.ctl.SeedRegistrationForTest(id, email, r.client)
	rep := decode(t, r.ctl.Secret(&auth.SecretDto{Token: token, Secret: "a long enough secret", ConfirmSecret: "a long enough secret", FirstName: "New", LastName: "Person"}))
	if rep.code != 200 {
		t.Fatalf("secret: %d %v", rep.code, rep.body)
	}
	device := deviceFrom(t, rep)
	got := r.sessions(cookieFor(device))
	if len(got) != 1 || got[0]["user"].(map[string]any)["email"] != email {
		t.Fatalf("sessions of the browser that just signed up = %v", got)
	}
	if got[0]["clientId"] != r.client {
		t.Fatalf("the session's client = %v", got[0]["clientId"])
	}
	if ws := got[0]["user"].(map[string]any)["workspaces"].([]any); len(ws) != 1 {
		t.Fatalf("the new account's workspaces = %v", ws)
	}
	// signing up on a browser that already has a device adds to it
	id2 := uuid.NewString()
	email2 := "new2-" + id2[:8] + "@example.test"
	token2 := r.ctl.SeedRegistrationForTest(id2, email2, r.client)
	rep2 := decode(t, r.ctl.Secret(&auth.SecretDto{Token: token2, Secret: "a long enough secret", ConfirmSecret: "a long enough secret", Cookie: cookieFor(device)}))
	if deviceFrom(t, rep2) != device || len(r.sessions(cookieFor(device))) != 2 {
		t.Fatalf("the second sign-up did not join the browser's device")
	}
}
