package tests

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apikeyapp "github.com/thescaffold/gox-apps/libs/identity/app/apikey"
	"github.com/thescaffold/gox-apps/libs/identity/app/auth"
	orgapp "github.com/thescaffold/gox-apps/libs/identity/app/org"
	"github.com/thescaffold/gox-apps/libs/identity/pkg/apikey"
	"github.com/thescaffold/gox-apps/libs/identity/pkg/org"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
)

// A person manages their own API keys. Owners and admins (org.tokens.manage) also list and revoke
// the workspace's other people's keys, but never rotate them, so no secret of someone else's key
// ever reaches them. Keys of another brand (a host's machine credentials, such as service accounts)
// are not /keys' to show or change at all.
func TestApiKeys_BelongToTheirHolder(t *testing.T) {
	e := newOrgEnv(t)
	e.db.Exec(`TRUNCATE "IdentityApiKeys"`)
	svc := &apikey.Service{DB: e.db, Brand: "lgr"}
	c := apikeyapp.NewControllerForTest(svc)
	ws := newWS()
	owner := e.seat(t, ws, "Kay Owner", "owner")
	admin := e.seat(t, ws, "Kay Admin", "admin")
	member := e.seat(t, ws, "Kay Member", "member")
	guest := e.seat(t, ws, "Kay Guest", "guest")
	as := func(u string) ntxctx.NTXContext {
		return ntxctx.NTXContext{UserID: u, ClientID: "org-client", WorkspaceID: ws}
	}

	issue := func(u, name string) (id, secret string) {
		t.Helper()
		k, err := svc.Issue(context.Background(), u, "org-client", ws, name, apikey.Secret, apikey.Live, []string{"read"}, 0)
		if err != nil {
			t.Fatal(err)
		}
		return k.ID, k.Secret
	}
	memberKey, memberSecret := issue(member, "member laptop")
	guestKey, guestSecret := issue(guest, "guest laptop")
	adminKey, _ := issue(admin, "admin laptop")
	// A key of another brand: a host's service account, managed by the host's own routes.
	machine := uuid.NewString()
	saKey, err := (&apikey.Service{DB: e.db, Brand: "lgrsa"}).Issue(context.Background(), machine, "org-client", ws, "ci", apikey.Secret, apikey.Live, []string{"read"}, 0)
	if err != nil {
		t.Fatal(err)
	}

	names := func(u string) string {
		t.Helper()
		code, b := kbody(c.List(&apikeyapp.ListReq{Ctx: as(u)}))
		if code != 200 {
			t.Fatalf("list as %s: %d %s", u, code, b)
		}
		if strings.Contains(b, memberSecret) || strings.Contains(b, guestSecret) {
			t.Fatalf("a listing shows a secret: %s", b)
		}
		return b
	}
	if b := names(guest); !strings.Contains(b, "guest laptop") || strings.Contains(b, "member laptop") || strings.Contains(b, "admin laptop") || strings.Contains(b, `"ci"`) {
		t.Fatalf("a guest sees others' keys: %s", b)
	}
	if b := names(member); !strings.Contains(b, "member laptop") || strings.Contains(b, "guest laptop") {
		t.Fatalf("a member sees others' keys: %s", b)
	}
	if b := names(admin); !strings.Contains(b, "member laptop") || !strings.Contains(b, "guest laptop") || !strings.Contains(b, "admin laptop") || strings.Contains(b, `"ci"`) {
		t.Fatalf("an admin lists people's keys and not the host's machine keys: %s", b)
	}

	// A guest (or member) cannot revoke or rotate another person's key: not found, and it still works.
	if code, _ := kbody(c.Revoke(&apikeyapp.IDReq{Id: memberKey, Ctx: as(guest)})); code != 404 {
		t.Fatalf("a guest revokes a member's key: %d", code)
	}
	if code, _ := kbody(c.Rotate(&apikeyapp.IDReq{Id: memberKey, Ctx: as(guest)})); code != 404 {
		t.Fatalf("a guest rotates a member's key: %d", code)
	}
	if code, _ := kbody(c.Revoke(&apikeyapp.IDReq{Id: adminKey, Ctx: as(member)})); code != 404 {
		t.Fatalf("a member revokes an admin's key: %d", code)
	}
	if _, err := svc.Authenticate(context.Background(), memberSecret); err != nil {
		t.Fatalf("the member's key after the refused attempts: %v", err)
	}
	// Nobody reaches the machine key through /keys, an owner included.
	if code, _ := kbody(c.Revoke(&apikeyapp.IDReq{Id: saKey.ID, Ctx: as(owner)})); code != 404 {
		t.Fatalf("an owner revokes a machine key through /keys: %d", code)
	}
	// An admin never rotates someone else's key (the new secret would be theirs to see).
	if code, b := kbody(c.Rotate(&apikeyapp.IDReq{Id: guestKey, Ctx: as(admin)})); code != 403 || strings.Contains(b, "lgr_live_sk_") {
		t.Fatalf("an admin rotates a guest's key: %d %s", code, b)
	}
	// A person rotates their own; an admin revokes a member's.
	if code, b := kbody(c.Rotate(&apikeyapp.IDReq{Id: guestKey, Ctx: as(guest)})); code != 201 || !strings.Contains(b, "lgr_live_sk_") {
		t.Fatalf("a guest rotates their own key: %d %s", code, b)
	}
	if code, _ := kbody(c.Revoke(&apikeyapp.IDReq{Id: memberKey, Ctx: as(admin)})); code != 200 {
		t.Fatalf("an admin revokes a member's key: %d", code)
	}
	if _, err := svc.Authenticate(context.Background(), memberSecret); !errors.Is(err, apikey.ErrRevoked) {
		t.Fatalf("the revoked key: %v", err)
	}

	// The host's authorizer decides org.tokens.manage: it can take it from admins or give it to others.
	org.SetAuthorizer(func(_ context.Context, w, u, action string) (bool, error) {
		return action == org.ManageTokens && u == guest, nil
	})
	defer org.SetAuthorizer(nil)
	if b := names(admin); strings.Contains(b, "guest laptop") {
		t.Fatalf("an admin the host denies still lists others' keys: %s", b)
	}
	if b := names(guest); !strings.Contains(b, "admin laptop") {
		t.Fatalf("a guest the host allows does not list others' keys: %s", b)
	}
}

// identity asks the host's authorizer, when one is set, for its org-level checks; without one, it
// keeps owner-and-admin behaviour exactly. The seat rules (nobody above their own seat, only an
// owner grants owner, one owner stays) apply whatever the authorizer says.
func TestOrg_AuthorizerHook(t *testing.T) {
	e := newOrgEnv(t)
	ws := newWS()
	owner := e.seat(t, ws, "Hook Owner", "owner")
	admin := e.seat(t, ws, "Hook Admin", "admin")
	member := e.seat(t, ws, "Hook Member", "member")
	guest := e.seat(t, ws, "Hook Guest", "guest")
	c := orgapp.NewControllerForTest(e.svc)

	// Default: owners and admins only.
	for _, a := range []string{org.ManageMembers, org.ManageSecurity, org.ManageSettings, org.ManageTokens, org.ManageRoles, org.ReadAudit} {
		for u, want := range map[string]bool{owner: true, admin: true, member: false, guest: false} {
			if got, err := e.svc.Allowed(e.ctx, ws, u, a); err != nil || got != want {
				t.Fatalf("default %s for %s: %v %v, want %v", a, u, got, err, want)
			}
		}
	}
	if err := e.svc.SetSeat(e.ctx, ws, member, guest, "member"); !errors.Is(err, org.ErrNotAllowed) {
		t.Fatalf("a member changes seats by default: %v", err)
	}

	asked := []string{}
	org.SetAuthorizer(func(_ context.Context, w, u, action string) (bool, error) {
		asked = append(asked, action)
		if u == admin {
			return false, nil // the host takes everything from this admin
		}
		return u == owner || (u == member && action == org.ManageMembers), nil
	})
	defer org.SetAuthorizer(nil)

	// The member may now manage members, within their own seat; the admin may not.
	if err := e.svc.SetSeat(e.ctx, ws, member, guest, "member"); err != nil {
		t.Fatalf("a member the host allows changes a guest's seat: %v", err)
	}
	if err := e.svc.SetSeat(e.ctx, ws, member, guest, "admin"); !errors.Is(err, org.ErrNotAllowed) {
		t.Fatalf("a member grants a seat above their own: %v", err)
	}
	if err := e.svc.SetSeat(e.ctx, ws, admin, guest, "guest"); !errors.Is(err, org.ErrNotAllowed) {
		t.Fatalf("an admin the host denies changes seats: %v", err)
	}
	if err := e.svc.CheckGrant(e.ctx, ws, member, "member"); err != nil {
		t.Fatalf("a member the host allows invites: %v", err)
	}
	if got := c.CreateTeam(&orgapp.TeamReq{Ctx: ntx(member, ws), Name: "Hooked"}).Code(); got != 201 {
		t.Fatalf("a member the host allows creates a team: %d", got)
	}
	if got := c.CreateTeam(&orgapp.TeamReq{Ctx: ntx(admin, ws), Name: "Denied"}).Code(); got != 403 {
		t.Fatalf("an admin the host denies creates a team: %d", got)
	}
	if err := e.svc.Rename(e.ctx, ws, member, "Renamed"); !errors.Is(err, org.ErrNotAllowed) {
		t.Fatalf("members.manage is not settings.manage: %v", err)
	}
	if err := e.svc.Rename(e.ctx, ws, owner, "Renamed"); err != nil {
		t.Fatalf("the owner renames: %v", err)
	}
	if got := c.Activity(&orgapp.Ctx{Ctx: ntx(admin, ws)}).Code(); got != 403 {
		t.Fatalf("an admin the host denies reads the activity: %d", got)
	}
	for _, want := range []string{org.ManageMembers, org.ManageSettings, org.ReadAudit} {
		found := false
		for _, a := range asked {
			found = found || a == want
		}
		if !found {
			t.Fatalf("the authorizer was never asked for %s (asked %v)", want, asked)
		}
	}
}

// The two-factor policy and the SSO connection are org.security.manage.
func TestOrg_AuthorizerCoversSecurity(t *testing.T) {
	r := newMfaRig(t)
	_, ws := r.person("sec-owner")
	dev, _ := r.person("sec-dev")
	r.join(dev, ws, "member")
	if out := decode(t, r.ctl.MfaPolicy(&auth.MfaDto{Ctx: ctxOf(dev, ws)})); out.code != 403 {
		t.Fatalf("a member reads the 2FA policy by default: %d", out.code)
	}
	if out := decode(t, r.ctl.SsoGetConnection(&auth.SsoWorkspaceDto{Ctx: ctxOf(dev, ws)})); out.code != 403 {
		t.Fatalf("a member reads the SSO connection by default: %d", out.code)
	}
	org.SetAuthorizer(func(_ context.Context, w, u, action string) (bool, error) {
		return u == dev.Id && action == org.ManageSecurity, nil
	})
	defer org.SetAuthorizer(nil)
	if out := decode(t, r.ctl.MfaPolicy(&auth.MfaDto{Ctx: ctxOf(dev, ws)})); out.code != 200 {
		t.Fatalf("a member the host allows reads the 2FA policy: %d %v", out.code, out.body)
	}
	if out := decode(t, r.ctl.SsoGetConnection(&auth.SsoWorkspaceDto{Ctx: ctxOf(dev, ws)})); out.code == 403 {
		t.Fatalf("a member the host allows is refused the SSO connection: %v", out.body)
	}
}

// Removing someone from a workspace runs the host's clean-up in the same transaction: if it fails,
// they keep their seat.
func TestOrg_SeatRemovalHook(t *testing.T) {
	e := newOrgEnv(t)
	ws := newWS()
	owner := e.seat(t, ws, "Rm Owner", "owner")
	member := e.seat(t, ws, "Rm Member", "member")
	other := e.seat(t, ws, "Rm Other", "member")

	var got []string
	org.OnSeatRemoved(func(_ context.Context, tx *gorm.DB, w, u, actor string) error {
		if u == other {
			return errors.New("clean-up failed")
		}
		var n int64
		tx.Table(`"IdentityUserClientWorkspaces"`).Where("workspace_id = ? AND user_id = ? AND deleted_at IS NULL", w, u).Count(&n)
		got = append(got, w+"/"+u+"/"+actor)
		if n != 0 {
			return errors.New("the hook ran before the seat was gone")
		}
		return nil
	})
	defer org.OnSeatRemoved(nil)

	if err := e.svc.RemoveSeat(org.WithActor(e.ctx, owner), ws, owner, member); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != ws+"/"+member+"/"+owner {
		t.Fatalf("the hook heard %v", got)
	}
	if err := e.svc.RemoveSeat(e.ctx, ws, owner, other); err == nil {
		t.Fatal("a failed clean-up must fail the removal")
	}
	if s, _ := e.svc.Seat(e.ctx, ws, other); s != "member" {
		t.Fatalf("a failed removal took the seat: %q", s)
	}
}

// Sign-in, sign-up and passwordless codes name the app they are for; without one they are refused
// before anything happens.
func TestSignInNeedsAClientId(t *testing.T) {
	r := plRig(t)
	u, _ := r.person("noclient")
	for name, out := range map[string]interface{ Code() int }{
		"login":      r.ctl.Login(&auth.LoginDto{Ref: *u.Ref, Secret: testPassword, ClientId: "  "}),
		"initiate":   r.ctl.Initiate(&auth.InitiateDto{Ref: uniq("noclient"), Type: "email"}),
		"otp":        r.ctl.OtpRequest(&auth.OtpRequestDto{Ref: uniq("noclient")}),
		"otp/verify": r.ctl.OtpVerify(&auth.OtpVerifyDto{Ref: uniq("noclient"), Code: "424242"}),
	} {
		if out.Code() != 400 {
			t.Errorf("%s without clientId: HTTP %d, want 400", name, out.Code())
		}
	}
	// A session that names no app cannot make a workspace.
	if out := decode(t, r.ctl.CreateMyWorkspace(&auth.CreateMyWorkspaceDto{Ctx: ntxctx.NTXContext{UserID: u.Id}, Name: "Orphan"})); out.code != 400 {
		t.Fatalf("create a workspace without a client: %d", out.code)
	}
	if out := r.login(u, ""); out.code != 200 {
		t.Fatalf("login with a clientId: %d %v", out.code, out.body)
	}
}

// An owner switches into their workspace from any app; anyone else needs a membership under the
// app they signed in from, and a deleted workspace is not reopened.
func TestOwnerSwitchesInFromAnyClient(t *testing.T) {
	r := newRig(t)
	owner, ws := r.person("switcher")
	member, _ := r.person("switch-member")
	r.join(member, ws, "member")
	elsewhere := "other-" + uuid.NewString()[:8]

	if rep := decode(t, r.ctl.SwitchWorkspace(&auth.SwitchWorkspaceDto{Ctx: ntxctx.NTXContext{UserID: member.Id, ClientID: elsewhere}, WorkspaceId: ws})); rep.code != 401 {
		t.Fatalf("a member switches in from another app: %d", rep.code)
	}
	rep := decode(t, r.ctl.SwitchWorkspace(&auth.SwitchWorkspaceDto{Ctx: ntxctx.NTXContext{UserID: owner.Id, ClientID: elsewhere}, WorkspaceId: ws}))
	if rep.code != 200 || rep.data()["accessToken"] == nil {
		t.Fatalf("the owner switches in from another app: %d %v", rep.code, rep.body)
	}
	var n int64
	r.db.Table(`"IdentityUserClientWorkspaces"`).Where(`user_id = ? AND client_id = ? AND workspace_id = ? AND lower(status) = 'owner' AND deleted_at IS NULL`, owner.Id, elsewhere, ws).Count(&n)
	if n != 1 {
		t.Fatalf("the owner's membership for the other app: %d rows", n)
	}
	// Again: no second row.
	_ = decode(t, r.ctl.SwitchWorkspace(&auth.SwitchWorkspaceDto{Ctx: ntxctx.NTXContext{UserID: owner.Id, ClientID: elsewhere}, WorkspaceId: ws}))
	r.db.Table(`"IdentityUserClientWorkspaces"`).Where(`user_id = ? AND workspace_id = ? AND deleted_at IS NULL`, owner.Id, ws).Count(&n)
	if n != 2 {
		t.Fatalf("two switches made %d rows, want 2 (one per app)", n)
	}

	gone, gws := r.person("switch-gone")
	r.db.Exec(`UPDATE "IdentityWorkspaces" SET deleted_at = now() WHERE id = ?`, gws)
	if rep := decode(t, r.ctl.SwitchWorkspace(&auth.SwitchWorkspaceDto{Ctx: ntxctx.NTXContext{UserID: gone.Id, ClientID: elsewhere}, WorkspaceId: gws})); rep.code != 401 {
		t.Fatalf("an owner switches into a deleted workspace from another app: %d", rep.code)
	}
}

// A person renames themself.
func TestAccountRename(t *testing.T) {
	r := newRig(t)
	u, _ := r.person("renamer")
	me := ntxctx.NTXContext{UserID: u.Id, ClientID: r.client}
	if rep := decode(t, r.ctl.UpdateAccount(&auth.UpdateAccountDto{Ctx: me, Name: "   "})); rep.code != 400 {
		t.Fatalf("a blank name: %d", rep.code)
	}
	if rep := decode(t, r.ctl.UpdateAccount(&auth.UpdateAccountDto{Ctx: me, Name: strings.Repeat("n", 81)})); rep.code != 400 {
		t.Fatalf("an 81-character name: %d", rep.code)
	}
	rep := decode(t, r.ctl.UpdateAccount(&auth.UpdateAccountDto{Ctx: me, Name: "  Ada   Lovelace "}))
	if rep.code != 200 || rep.data()["name"] != "Ada Lovelace" {
		t.Fatalf("rename: %d %v", rep.code, rep.body)
	}
	var name string
	r.db.Raw(`SELECT name FROM "IdentityUsers" WHERE id = ?`, u.Id).Scan(&name)
	if name != "Ada Lovelace" {
		t.Fatalf("stored name %q", name)
	}
	if rep := decode(t, r.ctl.UpdateAccount(&auth.UpdateAccountDto{Name: "Nobody"})); rep.code != 401 {
		t.Fatalf("rename without a session: %d", rep.code)
	}
}

// The time zone stays a person's choice after sign-up; country and currency are still set once.
func TestTimezoneStaysEditable(t *testing.T) {
	r := newRig(t)
	u, _ := r.person("tz")
	for _, tz := range []string{"Africa/Lagos", "Europe/Berlin"} {
		if code := r.patchAttrs(u.Id, "preference", map[string]any{"timezone": tz, "country": map[string]string{"Africa/Lagos": "NG", "Europe/Berlin": "DE"}[tz]}); code != 200 {
			t.Fatalf("patch %s: %d", tz, code)
		}
	}
	if v, _ := r.attr(u.Id, "timezone"); v != "Europe/Berlin" {
		t.Fatalf("timezone %q, want Europe/Berlin", v)
	}
	if v, _ := r.attr(u.Id, "country"); v != "NG" {
		t.Fatalf("country %q, want it kept at NG", v)
	}
}
