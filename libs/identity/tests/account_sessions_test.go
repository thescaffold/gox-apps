package tests

import (
	"strings"
	"testing"

	"github.com/thescaffold/gox-apps/libs/identity/app/auth"
	"github.com/thescaffold/gox-apps/libs/identity/app/user"
	"github.com/thescaffold/gox-apps/libs/identity/pkg/revocation"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/utils"
)

// V0-15 (PLT-02): Account › Password and 2FA lists a person's own device sessions across devices
// (browser, place, last seen) with Sign out and Sign out everywhere else. A session's tokens carry
// its id (sid), so signing it out ends them at once, not 15 minutes later.

const (
	chromeMac = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36"
	safariIOS = "Mozilla/5.0 (iPhone; CPU iPhone OS 17_5 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.5 Mobile/15E148 Safari/604.1"
)

func (r *rig) loginFrom(u *user.User, agent, ip string) reply {
	r.t.Helper()
	rep := decode(r.t, r.ctl.Login(&auth.LoginDto{Ref: *u.Ref, Secret: testPassword, ClientId: r.client, UserAgent: agent, Ctx: ntxctx.NTXContext{IP: ip}}))
	if rep.code != 200 {
		r.t.Fatalf("login: %d %v", rep.code, rep.body)
	}
	return rep
}

// as is the context a verified token gives a handler (claimsToContext).
func as(t *testing.T, token string) ntxctx.NTXContext {
	t.Helper()
	claims, err := coreauth.Verify(token, "changeme")
	if err != nil {
		t.Fatal(err)
	}
	sub, _ := claims["sub"].(string)
	client, _ := claims["clientId"].(string)
	sid, _ := claims["sid"].(string)
	return ntxctx.NTXContext{UserID: sub, ClientID: client, User: utils.KeyValue{"sid": sid}}
}

func sidOf(t *testing.T, token string) string {
	t.Helper()
	claims, err := coreauth.Verify(token, "changeme")
	if err != nil {
		t.Fatal(err)
	}
	sid, _ := claims["sid"].(string)
	return sid
}

func TestAccountSessions_ListAndSignOutOnlyOnesOwn(t *testing.T) {
	r := newRig(t)
	revocation.SetStore(&revocation.Store{DB: r.db})
	t.Cleanup(func() { revocation.SetStore(nil) })
	ada, _ := r.person("ada")
	eve, _ := r.person("eve")

	laptop := r.loginFrom(ada, chromeMac, "203.0.113.4")
	phone := r.loginFrom(ada, safariIOS, "198.51.100.7")
	evil := r.loginFrom(eve, chromeMac, "192.0.2.1")
	laptopDevice, phoneDevice := deviceFrom(t, laptop), deviceFrom(t, phone)
	laptopTok, phoneTok := laptop.data()["accessToken"].(string), phone.data()["accessToken"].(string)
	eveTok := evil.data()["accessToken"].(string)

	// Every token of a sign-in names its session, the workspace tokens too.
	laptopSid := sidOf(t, laptopTok)
	if laptopSid == "" || sidOf(t, phoneTok) == "" || laptopSid == sidOf(t, phoneTok) {
		t.Fatalf("sign-in tokens carry no or the same session id: %q %q", laptopSid, sidOf(t, phoneTok))
	}
	ws := laptop.data()["workspaces"].([]any)[0].(map[string]any)["workspace"].(map[string]any)["token"].(string)
	if sidOf(t, ws) != laptopSid || sidOf(t, laptop.data()["refreshToken"].(string)) != laptopSid {
		t.Fatal("the workspace or refresh token does not carry the session id")
	}
	// The device's own re-mint (GET /sessions) carries it too.
	reminted := r.sessions(cookieFor(laptopDevice))[0]["user"].(map[string]any)["token"].(string)
	if sidOf(t, reminted) != laptopSid {
		t.Fatal("a re-minted token does not carry the session id")
	}

	list := func(tok, device string) []map[string]any {
		t.Helper()
		rep := decode(t, r.ctl.AccountSessions(&auth.AccountSessionsDto{Ctx: as(t, tok), Cookie: cookieFor(device)}))
		if rep.code != 200 {
			t.Fatalf("list: %d %v", rep.code, rep.body)
		}
		var out []map[string]any
		for _, s := range rep.list() {
			out = append(out, s.(map[string]any))
		}
		return out
	}
	mine := list(laptopTok, laptopDevice)
	if len(mine) != 2 {
		t.Fatalf("ada has %d sessions listed, want her 2: %v", len(mine), mine)
	}
	var cur, other map[string]any
	for _, s := range mine {
		if s["current"] == true {
			cur = s
		} else {
			other = s
		}
	}
	if cur == nil || other == nil || cur["id"] != laptopSid {
		t.Fatalf("the current session is not the laptop's: %v", mine)
	}
	if cur["browser"] != "Chrome" || cur["os"] != "macOS" || cur["ip"] != "203.0.113.4" || cur["lastSeenAt"] == nil {
		t.Fatalf("the laptop session reads %v", cur)
	}
	if other["browser"] != "Safari" || other["os"] != "iOS" || other["ip"] != "198.51.100.7" {
		t.Fatalf("the phone session reads %v", other)
	}
	for _, s := range mine {
		if _, leaked := s["token"]; leaked {
			t.Fatalf("a session's secret token is listed: %v", s)
		}
	}
	if strings.Contains(stringify(list(eveTok, deviceFrom(t, evil))), ada.Id) {
		t.Fatal("eve's list shows ada's sessions")
	}

	// Nobody signs out someone else's session: not found, and it still works.
	if rep := decode(t, r.ctl.AccountSessionRevoke(&auth.AccountSessionDto{Ctx: as(t, eveTok), Id: laptopSid})); rep.code != 404 {
		t.Fatalf("eve signed out ada's session: %d", rep.code)
	}
	if len(list(laptopTok, laptopDevice)) != 2 {
		t.Fatal("ada lost a session to eve's attempt")
	}

	// Ada signs her phone out: its token stops at once, the laptop's keeps working.
	phoneSid := other["id"].(string)
	if rep := decode(t, r.ctl.AccountSessionRevoke(&auth.AccountSessionDto{Ctx: as(t, laptopTok), Id: phoneSid})); rep.code != 200 {
		t.Fatalf("sign out the phone: %d %v", rep.code, rep.body)
	}
	if claims, _ := coreauth.Verify(phoneTok, "changeme"); !revocation.Check(claims) {
		t.Fatal("the signed-out phone's token is still accepted")
	}
	if claims, _ := coreauth.Verify(laptopTok, "changeme"); revocation.Check(claims) {
		t.Fatal("the laptop's token was refused")
	}
	if got := r.sessions(cookieFor(phoneDevice)); len(got) != 0 {
		t.Fatalf("the phone still re-mints tokens: %v", got)
	}

	// Sign out everywhere else: two more devices go, this one stays; eve's is untouched.
	r.loginFrom(ada, safariIOS, "198.51.100.7")
	r.loginFrom(ada, chromeMac, "198.51.100.8")
	if n := len(list(laptopTok, laptopDevice)); n != 3 {
		t.Fatalf("%d sessions before signing the others out, want 3", n)
	}
	rep := decode(t, r.ctl.AccountSessionsRevokeOthers(&auth.AccountSessionsDto{Ctx: as(t, laptopTok), Cookie: cookieFor(laptopDevice)}))
	if rep.code != 200 || rep.data()["ended"] != float64(2) {
		t.Fatalf("sign out everywhere else: %d %v", rep.code, rep.body)
	}
	left := list(laptopTok, laptopDevice)
	if len(left) != 1 || left[0]["id"] != laptopSid {
		t.Fatalf("left %v, want only this session", left)
	}
	if claims, _ := coreauth.Verify(laptopTok, "changeme"); revocation.Check(claims) {
		t.Fatal("signing the others out ended this session too")
	}
	if len(list(eveTok, deviceFrom(t, evil))) != 1 {
		t.Fatal("eve's session went with ada's")
	}

	// Without a session id or a device cookie, there is no "this session" to keep.
	if rep := decode(t, r.ctl.AccountSessionsRevokeOthers(&auth.AccountSessionsDto{Ctx: ntxctx.NTXContext{UserID: ada.Id}})); rep.code != 400 {
		t.Fatalf("sign out the others with no current session: %d", rep.code)
	}
	if rep := decode(t, r.ctl.AccountSessions(&auth.AccountSessionsDto{})); rep.code != 401 {
		t.Fatalf("list with no user: %d", rep.code)
	}
}

// Refreshing carries the session id, so a refreshed token ends with its session.
func TestAccountSessions_RefreshKeepsTheSession(t *testing.T) {
	r := newRig(t)
	revocation.SetStore(&revocation.Store{DB: r.db})
	t.Cleanup(func() { revocation.SetStore(nil) })
	ada, _ := r.person("ada")
	in := r.loginFrom(ada, chromeMac, "203.0.113.4")
	sid := sidOf(t, in.data()["accessToken"].(string))
	rep := decode(t, r.ctl.Refresh(&auth.RefreshDto{Token: in.data()["refreshToken"].(string)}))
	if rep.code != 200 {
		t.Fatalf("refresh: %d %v", rep.code, rep.body)
	}
	fresh := rep.data()["accessToken"].(string)
	if sidOf(t, fresh) != sid {
		t.Fatal("the refreshed token lost its session id")
	}
	r.ctl.AccountSessionRevoke(&auth.AccountSessionDto{Ctx: as(t, fresh), Id: sid})
	if claims, _ := coreauth.Verify(fresh, "changeme"); !revocation.Check(claims) {
		t.Fatal("the refreshed token outlives its session")
	}
	if rep := decode(t, r.ctl.Refresh(&auth.RefreshDto{Token: in.data()["refreshToken"].(string)})); rep.code != 401 {
		t.Fatalf("a signed-out session refreshes: %d", rep.code)
	}
}

func stringify(v any) string {
	b := new(strings.Builder)
	for _, m := range v.([]map[string]any) {
		for k, x := range m {
			b.WriteString(k)
			if s, ok := x.(string); ok {
				b.WriteString(s)
			}
		}
	}
	return b.String()
}
