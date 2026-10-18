package tests

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/thescaffold/gox-apps/libs/identity/app"
	orgapp "github.com/thescaffold/gox-apps/libs/identity/app/org"
	"github.com/thescaffold/gox-apps/libs/identity/pkg"
	"github.com/thescaffold/gox-apps/libs/identity/pkg/org"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
)

type orgEnv struct {
	db  *gorm.DB
	svc *org.Service
	ctx context.Context
}

func newOrgEnv(t *testing.T) *orgEnv {
	t.Helper()
	db := identityTestDB(t)
	return &orgEnv{db: db, svc: org.New(db), ctx: context.Background()}
}

// seat creates a user and gives them a seat (owner, admin, member…) in a workspace.
func (e *orgEnv) seat(t *testing.T, ws, name, status string) string {
	t.Helper()
	id := uuid.NewString()
	if err := e.db.Exec(`INSERT INTO "IdentityUsers" (id, created_at, updated_at, name, email) VALUES (?, now(), now(), ?, ?)`, id, name, strings.ToLower(strings.ReplaceAll(name, " ", "."))+"-"+id[:6]+"@org.test").Error; err != nil {
		t.Fatal(err)
	}
	e.also(t, ws, id, status)
	return id
}

func (e *orgEnv) also(t *testing.T, ws, user, status string) {
	t.Helper()
	if err := e.db.Exec(`INSERT INTO "IdentityUserClientWorkspaces" (id, created_at, updated_at, user_id, client_id, workspace_id, status) VALUES (?, now(), now(), ?, 'org-client', ?, ?)`, uuid.NewString(), user, ws, status).Error; err != nil {
		t.Fatal(err)
	}
}

func newWS() string { return uuid.NewString() }

func perms(t *testing.T, e *orgEnv, ws, user string) []string {
	t.Helper()
	g, err := e.svc.For(e.ctx, ws, user)
	if err != nil {
		t.Fatal(err)
	}
	return g.Permissions
}

func TestOrg_TeamsNestingAndIsolation(t *testing.T) {
	e := newOrgEnv(t)
	wsA, wsB := newWS(), newWS()

	eng, err := e.svc.CreateTeam(e.ctx, wsA, "Engineering", org.KindDepartment, "builds it", "")
	if err != nil {
		t.Fatal(err)
	}
	web, err := e.svc.CreateTeam(e.ctx, wsA, "Web", org.KindTeam, "", eng.ID)
	if err != nil || web.ParentID == nil || *web.ParentID != eng.ID {
		t.Fatalf("nested team: %+v %v", web, err)
	}
	if _, err := e.svc.CreateTeam(e.ctx, wsA, "engineering", "", "", ""); !errors.Is(err, org.ErrDuplicate) {
		t.Fatalf("a name differing only by case: %v", err)
	}
	// Another workspace may use the same name, and cannot parent under ours.
	if _, err := e.svc.CreateTeam(e.ctx, wsB, "Engineering", "", "", ""); err != nil {
		t.Fatalf("same name in another workspace: %v", err)
	}
	if _, err := e.svc.CreateTeam(e.ctx, wsB, "Sneaky", "", "", eng.ID); !errors.Is(err, org.ErrInvalid) {
		t.Fatalf("parenting under another workspace's team: %v", err)
	}
	for name, args := range map[string][3]string{"empty name": {"", "team", ""}, "bad kind": {"X", "squad", ""}, "long name": {strings.Repeat("n", 81), "team", ""}} {
		if _, err := e.svc.CreateTeam(e.ctx, wsA, args[0], args[1], args[2], ""); !errors.Is(err, org.ErrInvalid) {
			t.Errorf("%s: %v", name, err)
		}
	}

	// Moving a team beneath its own descendant would make a loop.
	if _, err := e.svc.UpdateTeam(e.ctx, wsA, eng.ID, nil, nil, &web.ID); !errors.Is(err, org.ErrCycle) {
		t.Fatalf("cycle: %v", err)
	}
	if _, err := e.svc.UpdateTeam(e.ctx, wsA, eng.ID, nil, nil, &eng.ID); !errors.Is(err, org.ErrCycle) {
		t.Fatalf("self parent: %v", err)
	}
	top := ""
	if moved, err := e.svc.UpdateTeam(e.ctx, wsA, web.ID, ptr("Web platform"), nil, &top); err != nil || moved.ParentID != nil || moved.Name != "Web platform" {
		t.Fatalf("move to top level: %+v %v", moved, err)
	}
	// A department with teams beneath it cannot be deleted.
	if _, err := e.svc.UpdateTeam(e.ctx, wsA, web.ID, nil, nil, &eng.ID); err != nil {
		t.Fatal(err)
	}
	if err := e.svc.DeleteTeam(e.ctx, wsA, eng.ID); !errors.Is(err, org.ErrHasChild) {
		t.Fatalf("delete with children: %v", err)
	}

	// Another workspace can neither see, change nor delete ours.
	if ts, _ := e.svc.ListTeams(e.ctx, wsB); len(ts) != 1 {
		t.Fatalf("workspace B sees %d teams", len(ts))
	}
	if _, err := e.svc.UpdateTeam(e.ctx, wsB, eng.ID, ptr("pwned"), nil, nil); !errors.Is(err, org.ErrNotFound) {
		t.Fatalf("update across workspaces: %v", err)
	}
	if err := e.svc.DeleteTeam(e.ctx, wsB, web.ID); !errors.Is(err, org.ErrNotFound) {
		t.Fatalf("delete across workspaces: %v", err)
	}
	if ts, _ := e.svc.ListTeams(e.ctx, wsA); len(ts) != 2 {
		t.Fatalf("workspace A lost teams: %d", len(ts))
	}
}

func ptr(s string) *string { return &s }

func TestOrg_MembershipNeedsASeat(t *testing.T) {
	e := newOrgEnv(t)
	ws, other := newWS(), newWS()
	ada := e.seat(t, ws, "Ada Lovelace", "owner")
	grace := e.seat(t, ws, "Grace Hopper", "member")
	eve := e.seat(t, other, "Eve", "member")
	team, _ := e.svc.CreateTeam(e.ctx, ws, "Ops", "", "", "")

	if err := e.svc.AddMember(e.ctx, ws, team.ID, eve, ""); !errors.Is(err, org.ErrNotMember) {
		t.Fatalf("someone from another workspace: %v", err)
	}
	if err := e.svc.AddMember(e.ctx, ws, team.ID, uuid.NewString(), ""); !errors.Is(err, org.ErrNotMember) {
		t.Fatalf("an unknown user: %v", err)
	}
	if err := e.svc.AddMember(e.ctx, ws, team.ID, grace, "boss"); !errors.Is(err, org.ErrInvalid) {
		t.Fatalf("bad team role: %v", err)
	}
	if err := e.svc.AddMember(e.ctx, other, team.ID, eve, ""); !errors.Is(err, org.ErrNotFound) {
		t.Fatalf("adding into another workspace's team: %v", err)
	}
	for _, u := range []string{grace, ada} {
		if err := e.svc.AddMember(e.ctx, ws, team.ID, u, ""); err != nil {
			t.Fatal(err)
		}
	}
	if err := e.svc.AddMember(e.ctx, ws, team.ID, ada, org.TeamLead); err != nil { // changes the role, no duplicate
		t.Fatal(err)
	}
	ms, err := e.svc.TeamMembers(e.ctx, ws, team.ID)
	if err != nil || len(ms) != 2 || ms[0].Name != "Ada Lovelace" || ms[0].Role != org.TeamLead || ms[1].Role != org.TeamMember {
		t.Fatalf("team members (leads first) = %+v %v", ms, err)
	}
	all, _ := e.svc.Members(e.ctx, ws)
	if len(all) != 2 {
		t.Fatalf("workspace members = %+v", all)
	}
	for _, m := range all {
		if len(m.Teams) != 1 || m.Teams[0] != team.ID {
			t.Fatalf("member %s teams = %v", m.Name, m.Teams)
		}
	}
	if ts, _ := e.svc.ListTeams(e.ctx, ws); ts[0].Members != 2 {
		t.Fatalf("member count = %d", ts[0].Members)
	}
	if err := e.svc.RemoveMember(e.ctx, ws, team.ID, grace); err != nil {
		t.Fatal(err)
	}
	if err := e.svc.RemoveMember(e.ctx, ws, team.ID, grace); err != nil {
		t.Fatalf("removing twice is a no-op: %v", err)
	}
	if ms, _ := e.svc.TeamMembers(e.ctx, ws, team.ID); len(ms) != 1 {
		t.Fatalf("after removal: %+v", ms)
	}
}

func TestOrg_PermissionValidation(t *testing.T) {
	apps := map[string]bool{"files": true, "calendar": true}
	good := map[string]string{
		"read:apps:files:*:*:workspace":            "custom:read:apps:files:*:*:workspace",
		"member:create:apps:calendar:event:*:self": "custom:create:apps:calendar:event:*:self", // role segment replaced
		"  *:apps:files:*:*:workspace ":            "custom:*:apps:files:*:*:workspace",
		"read:apps:files:folder:abc-1:self":        "custom:read:apps:files:folder:abc-1:self",
	}
	for in, want := range good {
		if got, err := org.ValidatePermission(in, apps); err != nil || got != want {
			t.Errorf("%q -> %q, %v; want %q", in, got, err, want)
		}
	}
	bad := []string{
		"", "read", "read:apps:files:*:*", // too short
		"read:apps:files:*:*:global", // a workspace cannot grant platform scope
		"read:apps:files:*:*:*",      // nor an open scope
		"read:platform:files:*:*:workspace",
		"read:apps:*:*:*:workspace",           // name the app
		"read:apps:mail-secret:*:*:workspace", // not an app in this deployment
		"Read:apps:files:*:*:workspace",       // action is lower-case
		"read:apps:files:../x:*:workspace",
		"read:apps:files:*:*:workspace:extra:more",
		"a:b:c:d:e:f:g:h",
	}
	for _, in := range bad {
		if got, err := org.ValidatePermission(in, apps); err == nil {
			t.Errorf("%q was accepted as %q", in, got)
		}
	}
	// With no catalog registered, any well-formed app name is allowed.
	if _, err := org.ValidatePermission("read:apps:whatever:*:*:workspace", nil); err != nil {
		t.Errorf("no catalog: %v", err)
	}
}

func TestOrg_RolesCatalogAndResolution(t *testing.T) {
	e := newOrgEnv(t)
	ws, other := newWS(), newWS()
	// Register a catalog as the host does. Role names are unique to this test:
	// the catalog is deployment-wide and other tests share the database.
	svc := app.NewAppServiceForTest(e.db)
	r1, r2 := "orgcat-"+uuid.NewString()[:8], "orgglobal-"+uuid.NewString()[:8]
	if err := svc.OnRBACRegister(app.RBACRegisterPayload{RolePermissions: map[string][]string{
		r1: {r1 + ":read:apps:files:*:*:workspace", r1 + ":create:apps:calendar:*:*:self"},
		r2: {r2 + ":*:apps:billing:*:*:global"},
	}}); err != nil {
		t.Fatal(err)
	}

	cat, err := e.svc.Catalog(e.ctx)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, c := range cat {
		names = append(names, c.App)
	}
	if !contains(names, "files") || !contains(names, "calendar") || contains(names, "billing") {
		t.Fatalf("catalog apps = %v (platform-scope apps must not be offered)", names)
	}
	if _, err := e.svc.CreateRole(e.ctx, ws, "Biller", "", []string{"read:apps:billing:*:*:workspace"}); !errors.Is(err, org.ErrPermission) {
		t.Fatalf("a role for an app only offered at platform scope: %v", err)
	}

	ada := e.seat(t, ws, "Ada", "owner")
	grace := e.seat(t, ws, "Grace", "member")
	linus := e.seat(t, ws, "Linus", "member")
	mallory := e.seat(t, other, "Mallory", "member")

	editor, err := e.svc.CreateRole(e.ctx, ws, "Editor", "edits files", []string{"update:apps:files:*:*:workspace", "update:apps:files:*:*:workspace", "delete:apps:files:*:*:self"})
	if err != nil || len(editor.Permissions) != 2 {
		t.Fatalf("create role: %+v %v (duplicates collapse)", editor, err)
	}
	if _, err := e.svc.CreateRole(e.ctx, ws, "EDITOR", "", nil); !errors.Is(err, org.ErrDuplicate) {
		t.Fatalf("duplicate role name: %v", err)
	}
	roleB, _ := e.svc.CreateRole(e.ctx, other, "Editor", "", []string{"*:apps:files:*:*:workspace"}) // same name, other workspace

	// Held directly.
	if err := e.svc.Assign(e.ctx, ws, editor.ID, org.SubjectUser, grace); err != nil {
		t.Fatal(err)
	}
	if got := perms(t, e, ws, grace); len(got) != 2 || got[0] != "custom:delete:apps:files:*:*:self" {
		t.Fatalf("direct grant = %v", got)
	}
	if got := perms(t, e, ws, linus); len(got) != 0 {
		t.Fatalf("someone who holds nothing got %v", got)
	}

	// Held through a team, and through the department above it.
	dept, _ := e.svc.CreateTeam(e.ctx, ws, "Engineering", org.KindDepartment, "", "")
	team, _ := e.svc.CreateTeam(e.ctx, ws, "Web", org.KindTeam, "", dept.ID)
	reader, _ := e.svc.CreateRole(e.ctx, ws, "Calendar reader", "", []string{"read:apps:calendar:*:*:workspace"})
	_ = e.svc.AddMember(e.ctx, ws, team.ID, linus, "")
	_ = e.svc.Assign(e.ctx, ws, reader.ID, org.SubjectTeam, dept.ID)
	if got := perms(t, e, ws, linus); len(got) != 1 || got[0] != "custom:read:apps:calendar:*:*:workspace" {
		t.Fatalf("a role on the department must reach the team beneath it: %v", got)
	}
	g, _ := e.svc.For(e.ctx, ws, linus)
	if len(g.Teams) != 1 || g.Teams[0] != team.ID || len(g.Roles) != 1 || g.Roles[0] != "Calendar reader" {
		t.Fatalf("granted = %+v", g)
	}
	// ...but not upward: a role on a sub-team does not reach people who are only in its department.
	webOnly, _ := e.svc.CreateRole(e.ctx, ws, "Web only", "", []string{"create:apps:files:*:*:workspace"})
	_ = e.svc.Assign(e.ctx, ws, webOnly.ID, org.SubjectTeam, team.ID)
	_ = e.svc.AddMember(e.ctx, ws, dept.ID, grace, "")
	if got := perms(t, e, ws, grace); contains(got, "custom:create:apps:files:*:*:workspace") {
		t.Fatalf("a role on a sub-team leaked up to the department's members: %v", got)
	}
	if got := perms(t, e, ws, linus); !contains(got, "custom:create:apps:files:*:*:workspace") {
		t.Fatalf("the sub-team's own member lost its role: %v", got)
	}
	_ = e.svc.Assign(e.ctx, ws, editor.ID, org.SubjectTeam, team.ID)

	// A seat in another workspace gets nothing from this one's roles, and the
	// same person gets nothing for a workspace they have no seat in.
	_ = e.svc.AddMember(e.ctx, other, uuidTeam(t, e, other), mallory, "")
	if err := e.svc.Assign(e.ctx, ws, editor.ID, org.SubjectUser, mallory); !errors.Is(err, org.ErrNotMember) {
		t.Fatalf("assigning to someone outside the workspace: %v", err)
	}
	if got := perms(t, e, ws, mallory); len(got) != 0 {
		t.Fatalf("a stranger resolved %v", got)
	}
	if err := e.svc.Assign(e.ctx, other, editor.ID, org.SubjectUser, mallory); !errors.Is(err, org.ErrNotFound) {
		t.Fatalf("assigning another workspace's role: %v", err)
	}
	if got := perms(t, e, other, mallory); len(got) != 0 {
		t.Fatalf("their own workspace's role %s must not follow: %v (%s)", roleB.Name, got, roleB.ID)
	}
	// Holders count, assignment listing with names.
	rs, _ := e.svc.ListRoles(e.ctx, ws)
	var holders int
	for _, r := range rs {
		if r.ID == editor.ID {
			holders = r.Holders
		}
	}
	if holders != 2 {
		t.Fatalf("holders = %d, want the person and the team", holders)
	}
	as, _ := e.svc.Assignments(e.ctx, ws, editor.ID)
	if len(as) != 2 || as[0].Name == "" {
		t.Fatalf("assignments = %+v", as)
	}

	// Revoking: unassign, leave the team, delete the role.
	_ = e.svc.Unassign(e.ctx, ws, editor.ID, org.SubjectUser, grace)
	_ = e.svc.RemoveMember(e.ctx, ws, dept.ID, grace)
	if got := perms(t, e, ws, grace); len(got) != 0 { // no direct role, and out of the department
		t.Fatalf("after revoking = %v", got)
	}
	_ = e.svc.AddMember(e.ctx, ws, team.ID, grace, "")
	if got := perms(t, e, ws, grace); !contains(got, "custom:update:apps:files:*:*:workspace") {
		t.Fatalf("via the team again = %v", got)
	}
	if err := e.svc.DeleteRole(e.ctx, ws, editor.ID); err != nil {
		t.Fatal(err)
	}
	if got := perms(t, e, ws, grace); contains(got, "custom:update:apps:files:*:*:workspace") {
		t.Fatalf("a deleted role still grants %v", got)
	}
	// Deleting a team takes its roles and members with it; a person who leaves the workspace loses everything.
	_ = e.svc.Assign(e.ctx, ws, reader.ID, org.SubjectUser, grace)
	e.db.Exec(`UPDATE "IdentityUserClientWorkspaces" SET deleted_at = now() WHERE user_id = ? AND workspace_id = ?`, grace, ws)
	if got := perms(t, e, ws, grace); len(got) != 0 {
		t.Fatalf("someone who left the workspace still holds %v", got)
	}
	_ = ada
}

func uuidTeam(t *testing.T, e *orgEnv, ws string) string {
	t.Helper()
	tm, err := e.svc.CreateTeam(e.ctx, ws, "Their team", "", "", "")
	if err != nil {
		t.Fatal(err)
	}
	return tm.ID
}

func contains(xs []string, x string) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}

// The sign-in path: a person's tokens carry what their organisation position grants.
func TestOrg_TokensCarryCustomRolesAndTeams(t *testing.T) {
	e := newOrgEnv(t)
	ws, other := newWS(), newWS()
	grace := e.seat(t, ws, "Grace", "member")
	e.also(t, other, grace, "member") // the same person in a second workspace
	role, _ := e.svc.CreateRole(e.ctx, ws, "Files editor", "", []string{"update:apps:files:*:*:workspace"})
	team, _ := e.svc.CreateTeam(e.ctx, ws, "Docs", "", "", "")
	_ = e.svc.AddMember(e.ctx, ws, team.ID, grace, "")
	_ = e.svc.Assign(e.ctx, ws, role.ID, org.SubjectTeam, team.ID)

	auth := pkg.NewAuthServiceForTest(e.db)
	got, err := auth.ResolvePermissions(grace, "org-client", ws)
	if err != nil {
		t.Fatal(err)
	}
	var keys []string
	for _, p := range got {
		keys = append(keys, p.Key)
	}
	if !contains(keys, "custom:update:apps:files:*:*:workspace") {
		t.Fatalf("ws permissions = %v", keys)
	}
	elsewhere, _ := auth.ResolvePermissions(grace, "org-client", other)
	for _, p := range elsewhere {
		if strings.HasPrefix(p.Key, "custom:") {
			t.Fatalf("a role in one workspace followed the person into another: %s", p.Key)
		}
	}
}

// --- the HTTP rules -------------------------------------------------------

func ntx(user, ws string) ntxctx.NTXContext {
	return ntxctx.NTXContext{UserID: user, WorkspaceID: ws, ClientID: "org-client"}
}

func TestOrgController_WhoMayDoWhat(t *testing.T) {
	e := newOrgEnv(t)
	ws, other := newWS(), newWS()
	owner := e.seat(t, ws, "Ada", "owner")
	admin := e.seat(t, ws, "Boole", "admin")
	member := e.seat(t, ws, "Grace", "member")
	stranger := e.seat(t, other, "Mallory", "owner") // owns *another* workspace
	c := orgapp.NewControllerForTest(e.svc)

	code := func(out interface{ Code() int }) int { return out.Code() }

	// Anyone signed in to the workspace may read the org chart...
	for who, u := range map[string]string{"owner": owner, "admin": admin, "member": member} {
		if got := code(c.ListTeams(&orgapp.Ctx{Ctx: ntx(u, ws)})); got != 200 {
			t.Errorf("%s lists teams: %d", who, got)
		}
		if got := code(c.Members(&orgapp.Ctx{Ctx: ntx(u, ws)})); got != 200 {
			t.Errorf("%s lists members: %d", who, got)
		}
	}
	// ...only an owner or admin may change it.
	if got := code(c.CreateTeam(&orgapp.TeamReq{Ctx: ntx(member, ws), Name: "Sneaky"})); got != 403 {
		t.Errorf("member creates a team: %d", got)
	}
	if got := code(c.ListRoles(&orgapp.Ctx{Ctx: ntx(member, ws)})); got != 403 {
		t.Errorf("member lists roles: %d", got)
	}
	if got := code(c.CreateRole(&orgapp.RoleReq{Ctx: ntx(member, ws), Name: "Root", Permissions: []string{"*:apps:files:*:*:workspace"}})); got != 403 {
		t.Errorf("a member must not make themselves a role: %d", got)
	}
	out := c.CreateTeam(&orgapp.TeamReq{Ctx: ntx(admin, ws), Name: "Design", Kind: "group"})
	if out.Code() != 201 {
		t.Fatalf("admin creates a team: %d", out.Code())
	}
	teams, _ := e.svc.ListTeams(e.ctx, ws)
	tid := teams[0].ID
	if got := code(c.AddTeamMember(&orgapp.MemberReq{Ctx: ntx(owner, ws), Id: tid, UserID: member})); got != 200 {
		t.Errorf("owner adds a member: %d", got)
	}
	if got := code(c.AddTeamMember(&orgapp.MemberReq{Ctx: ntx(owner, ws), Id: tid, UserID: stranger})); got != 400 {
		t.Errorf("adding someone from another workspace: %d", got)
	}
	rout := c.CreateRole(&orgapp.RoleReq{Ctx: ntx(admin, ws), Name: "Reviewer", Permissions: []string{"read:apps:files:*:*:workspace"}})
	if rout.Code() != 201 {
		t.Fatalf("admin creates a role: %d", rout.Code())
	}
	if got := code(c.CreateRole(&orgapp.RoleReq{Ctx: ntx(admin, ws), Name: "Platform", Permissions: []string{"read:apps:files:*:*:global"}})); got != 400 {
		t.Errorf("a workspace role with platform scope: %d", got)
	}

	// An owner of some other workspace is a stranger here: every call, read or write, is refused,
	// and ids from this workspace do nothing in theirs.
	roles, _ := e.svc.ListRoles(e.ctx, ws)
	rid := roles[0].ID
	asStranger := ntx(stranger, ws) // claims naming a workspace they hold no seat in
	for name, out := range map[string]interface{ Code() int }{
		"list teams":   c.ListTeams(&orgapp.Ctx{Ctx: asStranger}),
		"members":      c.Members(&orgapp.Ctx{Ctx: asStranger}),
		"create team":  c.CreateTeam(&orgapp.TeamReq{Ctx: asStranger, Name: "x"}),
		"delete team":  c.DeleteTeam(&orgapp.IDReq{Ctx: asStranger, Id: tid}),
		"team members": c.TeamMembers(&orgapp.IDReq{Ctx: asStranger, Id: tid}),
		"list roles":   c.ListRoles(&orgapp.Ctx{Ctx: asStranger}),
		"assign":       c.Assign(&orgapp.AssignReq{Ctx: asStranger, Id: rid, SubjectType: "user", SubjectID: stranger}),
	} {
		if got := out.Code(); got != 403 {
			t.Errorf("stranger %s: %d, want 403", name, got)
		}
	}
	own := ntx(stranger, other)
	for name, out := range map[string]interface{ Code() int }{
		"their delete of our team": c.DeleteTeam(&orgapp.IDReq{Ctx: own, Id: tid}),
		"their team list of ours":  c.TeamMembers(&orgapp.IDReq{Ctx: own, Id: tid}),
		"their update of our role": c.UpdateRole(&orgapp.RoleUpdateReq{Ctx: own, Id: rid, Name: ptr("pwned")}),
		"their assignment":         c.Assign(&orgapp.AssignReq{Ctx: own, Id: rid, SubjectType: "user", SubjectID: stranger}),
	} {
		if got := out.Code(); got != 404 {
			t.Errorf("%s: %d, want 404", name, got)
		}
	}
	if rs, _ := e.svc.ListRoles(e.ctx, ws); rs[0].Name != "Reviewer" {
		t.Fatalf("a foreign request changed our role: %+v", rs[0])
	}
	if got := code(c.ListTeams(&orgapp.Ctx{Ctx: ntx("", "")})); got != 401 {
		t.Errorf("no identity: %d", got)
	}

	// A demoted admin loses the ability at once (the seat is read live, not from the token).
	e.db.Exec(`UPDATE "IdentityUserClientWorkspaces" SET status = 'member' WHERE user_id = ? AND workspace_id = ?`, admin, ws)
	if got := code(c.CreateTeam(&orgapp.TeamReq{Ctx: ntx(admin, ws), Name: "After demotion"})); got != 403 {
		t.Errorf("a demoted admin created a team: %d", got)
	}
}
