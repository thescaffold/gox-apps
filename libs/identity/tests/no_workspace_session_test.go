package tests

import (
	"testing"

	"github.com/google/uuid"
	"github.com/thescaffold/gox-apps/libs/identity/app/auth"
	"github.com/thescaffold/gox-apps/libs/identity/app/user"
	identityucw "github.com/thescaffold/gox-apps/libs/identity/app/userclientworkspace"
	identityworkspace "github.com/thescaffold/gox-apps/libs/identity/app/workspace"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
	"github.com/thescaffold/gox-packages/libs/core/security"
)

// A person who belongs to no workspace of the client (never invited, or their
// only workspace was deleted) can still sign in: the answer carries a token for
// the account alone. Their browser must hold a device session like everybody
// else's, or the sign-in ends with the access token's 15 minutes; and once they
// create or join a workspace, the same device session hands out its token.

// loner creates a user with a password and no workspace at all.
func (r *rig) loner(name string) *user.User {
	r.t.Helper()
	hashed, err := security.Hash(testPassword)
	if err != nil {
		r.t.Fatal(err)
	}
	email := name + "-" + uuid.NewString()[:8] + "@example.test"
	typ := "email"
	u := &user.User{Name: name, Email: email, Ref: &email, Type: &typ, Secret: &hashed}
	u.Id = uuid.NewString()
	if err := r.db.Create(u).Error; err != nil {
		r.t.Fatal(err)
	}
	return u
}

// joinNewWorkspace gives u a workspace of the rig's client, as creating or joining one does.
func (r *rig) joinNewWorkspace(u *user.User) string {
	r.t.Helper()
	ws := &identityworkspace.Workspace{ClientId: r.client, UserId: u.Id, Name: u.Name + "'s workspace", Slug: "ws-" + uuid.NewString()[:8]}
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
	return ws.Id
}

// accountToken checks that tok is a token for u alone (no workspace) on the rig's client.
func (r *rig) accountToken(tok any, u *user.User) {
	r.t.Helper()
	s, _ := tok.(string)
	claims, err := coreauth.Verify(s, "changeme")
	if err != nil || claims["sub"] != u.Id || claims["clientId"] != r.client {
		r.t.Fatalf("token %q: %v (%v), want one for %s", s, claims, err, u.Id)
	}
	if w, _ := claims["workspaceId"].(string); w != "" {
		r.t.Fatalf("the account token names workspace %q", w)
	}
}

// signedInWithoutWorkspace checks a sign-in answer for a person with no
// workspace: an account token, a device cookie, and a device session that
// lists them with a fresh account token. It returns the device.
func (r *rig) signedInWithoutWorkspace(rep reply, u *user.User) string {
	r.t.Helper()
	if rep.code != 200 {
		r.t.Fatalf("sign-in: %d %v", rep.code, rep.body)
	}
	d := rep.data()
	if ws, _ := d["workspaces"].([]any); len(ws) != 0 {
		r.t.Fatalf("workspaces = %v, want none", ws)
	}
	r.accountToken(d["accessToken"], u)
	if d["refreshToken"] == nil || d["refreshToken"] == "" {
		r.t.Fatal("no refresh token")
	}
	device := deviceFrom(r.t, rep)
	got := r.sessions(cookieFor(device))
	if len(got) != 1 {
		r.t.Fatalf("%d sessions on the device, want the one just opened: %v", len(got), got)
	}
	su := got[0]["user"].(map[string]any)
	if su["id"] != u.Id || got[0]["clientId"] != r.client {
		r.t.Fatalf("the device's session = %v", got[0])
	}
	if ws := su["workspaces"].([]any); len(ws) != 0 {
		r.t.Fatalf("the session's workspaces = %v, want none", ws)
	}
	r.accountToken(su["token"], u)
	return device
}

func TestSigningInWithNoWorkspaceOpensADeviceSessionThatUpgradesWhenTheyJoinOne(t *testing.T) {
	r := newRig(t)
	ada := r.loner("ada")

	device := r.signedInWithoutWorkspace(r.login(ada, ""), ada)

	// the device session outlives the access token: it is what the browser
	// re-mints from, and it lasts as long as anybody's
	var sessions int64
	r.db.Raw(`SELECT count(*) FROM "IdentityDeviceSessions" WHERE device_id = ? AND user_id = ? AND deleted_at IS NULL AND expires_at > now() + interval '29 days'`, device, ada.Id).Scan(&sessions)
	if sessions != 1 {
		t.Fatalf("%d live 30-day device sessions for ada, want 1", sessions)
	}

	// the refresh token works too, still without a workspace
	again := decode(t, r.ctl.Refresh(&auth.RefreshDto{Token: r.login(ada, cookieFor(device)).data()["refreshToken"].(string)}))
	if again.code != 200 {
		t.Fatalf("refresh: %d %v", again.code, again.body)
	}
	r.accountToken(again.data()["accessToken"], ada)
	if n := len(r.sessions(cookieFor(device))); n != 1 {
		t.Fatalf("signing in again on the device made %d sessions, want 1", n)
	}

	// she creates (or joins) a workspace: the same device session now hands
	// out a token for it, as for everyone else
	wsID := r.joinNewWorkspace(ada)
	got := r.sessions(cookieFor(device))
	if len(got) != 1 {
		t.Fatalf("%d sessions after joining, want 1", len(got))
	}
	wss := got[0]["user"].(map[string]any)["workspaces"].([]any)
	if len(wss) != 1 {
		t.Fatalf("workspaces after joining = %v", wss)
	}
	tok := wss[0].(map[string]any)["workspace"].(map[string]any)["token"].(string)
	claims, err := coreauth.Verify(tok, "changeme")
	if err != nil || claims["sub"] != ada.Id || claims["workspaceId"] != wsID {
		t.Fatalf("the workspace token = %v (%v), want one for ada in %s", claims, err, wsID)
	}
}

func TestPasswordlessSignInWithNoWorkspaceOpensADeviceSession(t *testing.T) {
	r := plRig(t)
	lin := r.loner("lin")
	rep := decode(t, r.ctl.OtpVerify(&auth.OtpVerifyDto{Ref: *lin.Ref, Code: "424242", ClientId: r.client, UserAgent: "test-agent"}))
	r.signedInWithoutWorkspace(rep, lin)
}
