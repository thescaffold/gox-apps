package tests

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/thescaffold/gox-apps/libs/identity/app/auth"
	appinvitation "github.com/thescaffold/gox-apps/libs/identity/app/invitation"
	"github.com/thescaffold/gox-apps/libs/identity/app/user"
	"github.com/thescaffold/gox-apps/libs/identity/migrations"
	"github.com/thescaffold/gox-apps/libs/identity/pkg/invitation"
	"github.com/thescaffold/gox-apps/libs/identity/pkg/org"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
)

func TestHandles(t *testing.T) {
	for in, want := range map[string]string{"Acme Corp": "acme-corp", "Café  Ünïcode!": "cafe-unicode", "  ": "org", "a": "orga", "--Hi--": "hi"} {
		if got := org.HandleFor(in); got != want {
			t.Errorf("HandleFor(%q) = %q, want %q", in, got, want)
		}
	}
	for _, bad := range []string{"settings", "a", "-ab", "ab-", "a--b", "Abc", strings.Repeat("a", 41)} {
		if org.ValidHandle(bad) == nil {
			t.Errorf("%q accepted", bad)
		}
	}
	if org.ValidHandle("acme-2") != nil {
		t.Error("acme-2 refused")
	}
}

func TestWorkspaceHandlesAndRegion(t *testing.T) {
	r := newRig(t)
	ada, _ := r.person("ada")
	bob, _ := r.person("bob")
	ctx := ntxctx.NTXContext{UserID: ada.Id, ClientID: r.client}
	name := "Acme " + uuid.NewString()[:6]
	first := decode(t, r.ctl.CreateMyWorkspace(&auth.CreateMyWorkspaceDto{Ctx: ctx, Name: name}))
	h1, _ := first.data()["handle"].(string)
	if first.code != 201 || h1 != org.HandleFor(name) || first.data()["region"] != "eu" {
		t.Fatalf("create: %d %v", first.code, first.body)
	}
	second := decode(t, r.ctl.CreateMyWorkspace(&auth.CreateMyWorkspaceDto{Ctx: ctx, Name: name}))
	if second.data()["handle"] != h1+"-2" {
		t.Fatalf("the same name gets a free handle: %v", second.body)
	}
	if out := decode(t, r.ctl.CreateMyWorkspace(&auth.CreateMyWorkspaceDto{Ctx: ctx, Name: "x", Handle: h1})); out.code != 409 {
		t.Fatalf("a taken handle: %d", out.code)
	}
	if out := decode(t, r.ctl.CreateMyWorkspace(&auth.CreateMyWorkspaceDto{Ctx: ctx, Name: "x", Region: "us"})); out.code != 400 {
		t.Fatalf("an unknown region: %d", out.code)
	}
	ws := first.data()["id"].(string)
	r.join(bob, ws, "member")
	want := "acme-" + uuid.NewString()[:6]
	if out := decode(t, r.ctl.UpdateMyWorkspace(&auth.UpdateMyWorkspaceDto{Ctx: ntxctx.NTXContext{UserID: bob.Id}, Id: ws, Handle: &want})); out.code != 403 {
		t.Fatalf("a member changes the handle: %d", out.code)
	}
	newName := "Acme Renamed"
	out := decode(t, r.ctl.UpdateMyWorkspace(&auth.UpdateMyWorkspaceDto{Ctx: ctx, Id: ws, Handle: &want, Name: &newName}))
	if out.code != 200 || out.data()["handle"] != want || out.data()["name"] != newName {
		t.Fatalf("owner renames: %d %v", out.code, out.body)
	}
	list := decode(t, r.ctl.MyWorkspaces(&auth.MyWorkspacesDto{Ctx: ctx}))
	if !strings.Contains(string(mustJSON(list.body)), `"handle":"`+want+`"`) {
		t.Fatalf("my workspaces lack the handle: %v", list.body)
	}

	// The migration gives existing workspaces handles, unique in order of creation.
	base := "Legacy " + uuid.NewString()[:6]
	for i := 0; i < 2; i++ {
		r.db.Exec(`INSERT INTO "IdentityWorkspaces" (id, created_at, updated_at, user_id, client_id, name, slug) VALUES (?, now() + (? * interval '1 second'), now(), ?, ?, ?, ?)`,
			uuid.NewString(), i, ada.Id, r.client, base, "WS"+uuid.NewString()[:8])
	}
	if err := (&migrations.WorkspaceHandleRegion{}).Run(r.q()); err != nil {
		t.Fatal(err)
	}
	var hs []string
	r.db.Raw(`SELECT "handle" FROM "IdentityWorkspaces" WHERE "name" = ? ORDER BY "created_at"`, base).Scan(&hs)
	if len(hs) != 2 || hs[0] != org.HandleFor(base) || hs[1] != org.HandleFor(base)+"-2" {
		t.Fatalf("backfill: %v", hs)
	}
}

type seatLimit struct{}

func (seatLimit) Error() string     { return "Free includes 3 seats" }
func (seatLimit) HTTPStatus() int   { return 402 }
func (seatLimit) ErrorCode() string { return "quota_exceeded" }

func TestInvitations(t *testing.T) {
	r := newRig(t)
	t.Setenv("IDENTITY_TOKEN_KEY", "0123456789abcdef0123456789abcdef")
	t.Setenv("PUBLIC_BASE_URL", "https://app.example.test")
	owner, ws := r.person("owner")
	admin, _ := r.person("admin")
	member, _ := r.person("member")
	r.join(admin, ws, "admin")
	r.join(member, ws, "member")
	clk := &clock{t: time.Now().UTC()}
	svc := &invitation.Service{DB: r.db, Now: clk.now}

	type mailed struct{ to, token, ws string }
	var mu sync.Mutex
	var mails []mailed
	c := appinvitation.NewControllerForTest(svc, func(inv invitation.Invitation, token, wsName, _ string) {
		mu.Lock()
		mails = append(mails, mailed{inv.Email, token, wsName})
		mu.Unlock()
	})
	var events []string
	invitation.SetObserver(func(_ context.Context, e invitation.Event) { mu.Lock(); events = append(events, e.Action); mu.Unlock() })
	defer invitation.SetObserver(nil)
	as := func(u *user.User) ntxctx.NTXContext {
		return ntxctx.NTXContext{UserID: u.Id, WorkspaceID: ws, ClientID: r.client}
	}

	dana, _ := r.person("dana")
	if out := decode(t, c.Invite(&appinvitation.InviteReq{Ctx: as(owner), Emails: []string{"not an email"}, Role: "member"})); out.code != 400 {
		t.Fatalf("a bad address: %d", out.code)
	}
	if out := decode(t, c.Invite(&appinvitation.InviteReq{Ctx: as(member), Emails: []string{*dana.Ref}, Role: "member"})); out.code != 403 {
		t.Fatalf("a member invites: %d", out.code)
	}
	if out := decode(t, c.Invite(&appinvitation.InviteReq{Ctx: as(admin), Emails: []string{*dana.Ref}, Role: "owner"})); out.code != 403 {
		t.Fatalf("an admin invites an owner: %d", out.code)
	}
	out := decode(t, c.Invite(&appinvitation.InviteReq{Ctx: as(owner), Emails: []string{strings.ToUpper(*dana.Ref) + ", " + *member.Ref}, Role: "member"}))
	if out.code != 201 || len(out.data()["invited"].([]any)) != 1 || len(out.data()["alreadyMembers"].([]any)) != 1 {
		t.Fatalf("invite: %d %v", out.code, out.body)
	}
	if len(mails) != 1 || mails[0].to != strings.ToLower(*dana.Ref) || mails[0].token == "" {
		t.Fatalf("mails: %+v", mails)
	}
	tok := mails[0].token

	look := decode(t, c.Look(&appinvitation.TokenReq{Token: tok}))
	p, _ := look.data()["invitation"].(map[string]any)
	if look.code != 200 || p["role"] != "member" || p["inviterName"] != "owner" || p["email"] != strings.ToLower(*dana.Ref) {
		t.Fatalf("preview: %d %v", look.code, look.body)
	}
	if out := decode(t, c.Look(&appinvitation.TokenReq{Token: "nope"})); out.code != 404 {
		t.Fatalf("unknown token: %d", out.code)
	}
	// Only the invited address accepts an email invitation.
	stranger, _ := r.person("stranger")
	if out := decode(t, c.Accept(&appinvitation.TokenReq{Ctx: ntxctx.NTXContext{UserID: stranger.Id}, Token: tok})); out.code != 403 {
		t.Fatalf("someone else accepts: %d", out.code)
	}
	// The host's seat limit refuses the join, with its own status and code.
	invitation.SetJoinGuard(func(*gorm.DB, string, string) error { return seatLimit{} })
	if out := decode(t, c.Accept(&appinvitation.TokenReq{Ctx: ntxctx.NTXContext{UserID: dana.Id}, Token: tok})); out.code != 402 || out.body["code"] != "quota_exceeded" {
		t.Fatalf("over the seat limit: %d %v", out.code, out.body)
	}
	invitation.SetJoinGuard(nil)
	if out := decode(t, c.Accept(&appinvitation.TokenReq{Ctx: ntxctx.NTXContext{UserID: dana.Id}, Token: tok})); out.code != 200 || out.data()["workspaceId"] != ws {
		t.Fatalf("accept: %d %v", out.code, out.body)
	}
	if seat, _ := org.New(r.db).Seat(context.Background(), ws, dana.Id); seat != "member" {
		t.Fatalf("dana's seat: %q", seat)
	}
	if out := decode(t, c.Accept(&appinvitation.TokenReq{Ctx: ntxctx.NTXContext{UserID: dana.Id}, Token: tok})); out.code != 410 {
		t.Fatalf("used twice: %d", out.code)
	}

	// The link: anyone signed in joins with its role; admins can copy it again.
	link := decode(t, c.Link(&appinvitation.LinkReq{Ctx: as(admin), Role: "guest"}))
	url, _ := link.data()["url"].(string)
	if link.code != 201 || !strings.HasPrefix(url, "https://app.example.test/invite/") {
		t.Fatalf("link: %d %v", link.code, link.body)
	}
	open := decode(t, c.Open(&appinvitation.OpenReq{Ctx: as(owner)}))
	if l, _ := open.data()["link"].(map[string]any); l["url"] != url {
		t.Fatalf("open link: %v", open.body)
	}
	if out := decode(t, c.Open(&appinvitation.OpenReq{Ctx: as(member)})); out.code != 403 {
		t.Fatalf("a member lists invitations: %d", out.code)
	}
	linkTok := strings.TrimPrefix(url, "https://app.example.test/invite/")
	for _, u := range []*user.User{stranger, bob(r)} {
		if out := decode(t, c.Accept(&appinvitation.TokenReq{Ctx: ntxctx.NTXContext{UserID: u.Id}, Token: linkTok})); out.code != 200 {
			t.Fatalf("link accept: %d %v", out.code, out.body)
		}
		if seat, _ := org.New(r.db).Seat(context.Background(), ws, u.Id); seat != "guest" {
			t.Fatalf("link seat: %q", seat)
		}
	}
	// A new link replaces the old; a revoked or expired one opens nothing.
	newer := decode(t, c.Link(&appinvitation.LinkReq{Ctx: as(owner)}))
	if out := decode(t, c.Look(&appinvitation.TokenReq{Token: linkTok})); out.code != 404 {
		t.Fatalf("the replaced link: %d", out.code)
	}
	newTok := strings.TrimPrefix(newer.data()["url"].(string), "https://app.example.test/invite/")
	clk.t = clk.t.Add(invitation.TTL + time.Minute)
	if out := decode(t, c.Look(&appinvitation.TokenReq{Token: newTok})); out.code != 410 {
		t.Fatalf("an expired link: %d", out.code)
	}
	clk.t = time.Now().UTC()
	if out := decode(t, c.Revoke(&appinvitation.IDReq{Ctx: as(owner), Id: newer.data()["id"].(string)})); out.code != 200 {
		t.Fatalf("revoke: %d", out.code)
	}
	if out := decode(t, c.Look(&appinvitation.TokenReq{Token: newTok})); out.code != 404 {
		t.Fatalf("a revoked link: %d", out.code)
	}

	mu.Lock()
	defer mu.Unlock()
	if strings.Join(events, ",") != "invitation.send,invitation.accept,invitation.link,invitation.accept,invitation.accept,invitation.link,invitation.revoke" {
		t.Fatalf("events: %v", events)
	}
}

// bob is one more person, for the link.
func bob(r *rig) *user.User { u, _ := r.person("bob"); return u }
