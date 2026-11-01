package tests

import (
	"testing"

	"github.com/awesome-goose/goose/modules/sql"
	"github.com/google/uuid"
	identityrole "github.com/thescaffold/gox-apps/libs/identity/app/role"
	identityroletype "github.com/thescaffold/gox-apps/libs/identity/app/roletype"
	"github.com/thescaffold/gox-apps/libs/identity/app/user"
	"github.com/thescaffold/gox-apps/libs/identity/pkg"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"gorm.io/gorm"
)

func rbacRoleEntity(db *gorm.DB) *identityrole.RoleEntity {
	e := &identityrole.RoleEntity{Entity: (&sql.Entity[identityrole.Role]{}).With((&sql.Query{}).With(&sql.Db{DB: db}))}
	e.OnRegister()
	return e
}

func rbacRoleType(t *testing.T, db *gorm.DB, name string) *identityroletype.RoleType {
	t.Helper()
	e := &identityroletype.RoleTypeEntity{Entity: (&sql.Entity[identityroletype.RoleType]{}).With((&sql.Query{}).With(&sql.Db{DB: db}))}
	e.OnRegister()
	rt := &identityroletype.RoleType{Name: name}
	if err := e.Insert(rt); err != nil {
		t.Fatalf("insert role type: %v", err)
	}
	t.Cleanup(func() { _, _ = e.Delete("id = ?", rt.Id) })
	return rt
}

func names(rs []identityrole.Role) map[string]bool {
	out := map[string]bool{}
	for _, r := range rs {
		out[r.Name] = true
	}
	return out
}

// The hole: ResolveRoles ignored the user, so a role attached to one person
// reached everyone in the same client and workspace.
func TestResolveRoles_UserRowsStayWithTheirUser(t *testing.T) {
	db := identityTestDB(t)
	auth := pkg.NewAuthServiceForTest(db)
	roles := rbacRoleEntity(db)

	client := "c-" + uuid.NewString()
	ws := uuid.NewString()
	ada, bola := uuid.NewString(), uuid.NewString()
	entity := "user"
	team := "workspace"
	rows := []*identityrole.Role{
		{Name: "for-ada", ClientId: client, WorkspaceId: &ws, Type: "user", EntityName: &entity, EntityId: &ada},
		{Name: "for-everyone", ClientId: client, WorkspaceId: &ws, Type: "workspace"},
		{Name: "for-the-workspace", ClientId: client, WorkspaceId: &ws, Type: "workspace", EntityName: &team, EntityId: &ws},
	}
	for _, r := range rows {
		if err := roles.Insert(r); err != nil {
			t.Fatalf("insert: %v", err)
		}
		id := r.Id
		t.Cleanup(func() { _, _ = roles.Delete("id = ?", id) })
	}

	forAda, err := auth.ResolveRoles(ada, client, ws)
	if err != nil {
		t.Fatal(err)
	}
	forBola, err := auth.ResolveRoles(bola, client, ws)
	if err != nil {
		t.Fatal(err)
	}
	got := names(forAda)
	if !got["for-ada"] || !got["for-everyone"] || !got["for-the-workspace"] {
		t.Fatalf("ada lost a role: %v", got)
	}
	got = names(forBola)
	if got["for-ada"] {
		t.Fatal("a role attached to ada reached bola")
	}
	if !got["for-everyone"] || !got["for-the-workspace"] {
		t.Fatalf("shared roles must still apply: %v", got)
	}
}

// An attached row stores the type's id as its name, so the type name has to be
// added for a token to say what the role is. The id stays, for existing readers.
func TestTokensNameAttachedRoles(t *testing.T) {
	t.Setenv("JWT_SECRET", "rbac-test-secret")
	db := identityTestDB(t)
	auth := pkg.NewAuthServiceForTest(db)
	roles := rbacRoleEntity(db)
	rt := rbacRoleType(t, db, "rbac-test-"+uuid.NewString()[:8])

	users := &user.UserEntity{Entity: (&sql.Entity[user.User]{}).With((&sql.Query{}).With(&sql.Db{DB: db}))}
	users.OnRegister()
	ref := "rbac-" + uuid.NewString()
	u := &user.User{Ref: &ref, Name: "T", Email: ref + "@example.test"}
	if err := users.Insert(u); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = users.Delete("id = ?", u.Id) })

	client := "c-" + uuid.NewString()
	entity := "user"
	row := &identityrole.Role{Name: rt.Id, ClientId: client, Type: "user", EntityName: &entity, EntityId: &u.Id, RoleTypeId: &rt.Id}
	if err := roles.Insert(row); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = roles.Delete("id = ?", row.Id) })

	access, err := auth.IssueAccessToken(u, client, "")
	if err != nil {
		t.Fatal(err)
	}
	claims, err := coreauth.Verify(access, "rbac-test-secret")
	if err != nil {
		t.Fatal(err)
	}
	held := map[string]bool{}
	for _, r := range claims["roles"].([]any) {
		held[r.(string)] = true
	}
	if !held[rt.Name] || !held[rt.Id] {
		t.Fatalf("roles %v must carry both the type name and its id", held)
	}
}

func TestScopeLink(t *testing.T) {
	t.Setenv("ROOT_USER_ID", "root-user")
	db := identityTestDB(t)
	ctl := identityrole.NewControllerForTest(db)
	editor := rbacRoleType(t, db, "site_editor_"+uuid.NewString()[:6])
	platform := rbacRoleType(t, db, "global-admin")

	owner := ntxctx.NTXContext{UserID: "o", ClientID: "c1", WorkspaceID: "w1", Roles: []any{"owner"}}
	root := ntxctx.NTXContext{UserID: "root-user", ClientID: "c1", WorkspaceID: "w1"}
	other := "w2"
	own := "w1"

	refused := []struct {
		name string
		dto  identityrole.LinkRoleDto
	}{
		{"platform role by type id", identityrole.LinkRoleDto{Ctx: owner, ClientId: "c1", RoleTypeIds: []string{platform.Id}}},
		{"platform role by literal", identityrole.LinkRoleDto{Ctx: owner, ClientId: "c1", RoleTypeIds: []string{"global-admin"}}},
		{"another workspace", identityrole.LinkRoleDto{Ctx: owner, ClientId: "c1", WorkspaceId: &other, RoleTypeIds: []string{editor.Id}}},
		{"another client", identityrole.LinkRoleDto{Ctx: owner, ClientId: "c2", RoleTypeIds: []string{editor.Id}}},
	}
	for _, c := range refused {
		dto := c.dto
		if ctl.ScopeLinkForTest(&dto) == "" {
			t.Errorf("%s must be refused", c.name)
		}
	}

	// the console sends no client or workspace, and must land in the caller's
	console := identityrole.LinkRoleDto{Ctx: owner, RoleTypeIds: []string{editor.Id}}
	if reason := ctl.ScopeLinkForTest(&console); reason != "" {
		t.Fatalf("owner assigning a role in their workspace: %s", reason)
	}
	if console.ClientId != "c1" || console.WorkspaceId == nil || *console.WorkspaceId != "w1" {
		t.Fatalf("scope not filled in: %+v", console)
	}
	explicit := identityrole.LinkRoleDto{Ctx: owner, ClientId: "c1", WorkspaceId: &own, RoleTypeIds: []string{editor.Id}}
	if reason := ctl.ScopeLinkForTest(&explicit); reason != "" {
		t.Fatalf("explicit own scope: %s", reason)
	}

	// provisioning as root keeps working, in any workspace, with any role
	provision := identityrole.LinkRoleDto{Ctx: root, ClientId: "c9", WorkspaceId: &other, RoleTypeIds: []string{platform.Id}}
	if reason := ctl.ScopeLinkForTest(&provision); reason != "" {
		t.Fatalf("root provisioning: %s", reason)
	}
	if provision.ClientId != "c9" || *provision.WorkspaceId != "w2" {
		t.Fatalf("root's scope must be left alone: %+v", provision)
	}
}
