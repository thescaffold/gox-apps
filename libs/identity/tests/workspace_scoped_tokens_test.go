package tests

import (
	"os"
	"testing"

	"github.com/thescaffold/gox-apps/libs/identity/app"
	identitypermission "github.com/thescaffold/gox-apps/libs/identity/app/permission"
	identityrole "github.com/thescaffold/gox-apps/libs/identity/app/role"
	"github.com/thescaffold/gox-apps/libs/identity/app/user"
	"github.com/thescaffold/gox-apps/libs/identity/pkg"
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/google/uuid"
	"github.com/thescaffold/gox-packages/libs/core/security"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// identityTestDB requires IDENTITY_TEST_DB_URL to point at a scratch
// Postgres database — runs the real app.AllMigrations itself (mirrors
// gox-apps/libs/autonomy/tests/integration_test.go's testDB), not a
// hand-copied schema. Skips, not fails, when unset.
func identityTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("IDENTITY_TEST_DB_URL")
	if dsn == "" {
		t.Skip("IDENTITY_TEST_DB_URL not set; skipping integration test")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	q := (&sql.Query{}).With(&sql.Db{DB: db})
	for _, m := range app.AllMigrations {
		if err := m.Run(q); err != nil {
			t.Fatalf("run migration %T: %v", m, err)
		}
	}
	return db
}

// TestAuthService_ResolvePermissions_ScopedToWorkspace is PLAN M1-02's own
// literal Done line made a test: a user with a permission in workspace A and
// a different permission in workspace B must, when resolved for A, see only
// A's permission (plus anything genuinely global) — never B's. Before this
// fix, an empty workspaceId meant "no filter, every workspace" (TRD F6);
// this test instead exercises the real bug shape directly: two *different*,
// non-empty workspaceIds must not see each other's rows either, which was
// never broken by F6 but is exactly what a real multi-workspace user relies
// on and deserves its own explicit coverage rather than an inference from
// the empty-string case alone.
func TestAuthService_ResolvePermissions_ScopedToWorkspace(t *testing.T) {
	db := identityTestDB(t)
	auth := pkg.NewAuthServiceForTest(db)
	permEntity := &identitypermission.PermissionEntity{Entity: (&sql.Entity[identitypermission.Permission]{}).With((&sql.Query{}).With(&sql.Db{DB: db}))}
	permEntity.OnRegister()

	userId := uuid.NewString()
	clientId := "c-" + uuid.NewString()
	wsA := uuid.NewString()
	wsB := uuid.NewString()

	permA := &identitypermission.Permission{
		UserId: &userId, ClientId: &clientId, WorkspaceId: &wsA,
		Key: "apps:systems:create", Scope: "workspace", Resource: "systems", Action: "create",
	}
	permB := &identitypermission.Permission{
		UserId: &userId, ClientId: &clientId, WorkspaceId: &wsB,
		Key: "apps:systems:delete", Scope: "workspace", Resource: "systems", Action: "delete",
	}
	if err := permEntity.Insert(permA); err != nil {
		t.Fatalf("insert permA: %v", err)
	}
	t.Cleanup(func() { _, _ = permEntity.Delete("id = ?", permA.Id) })
	if err := permEntity.Insert(permB); err != nil {
		t.Fatalf("insert permB: %v", err)
	}
	t.Cleanup(func() { _, _ = permEntity.Delete("id = ?", permB.Id) })

	forA, err := auth.ResolvePermissions(userId, clientId, wsA)
	if err != nil {
		t.Fatalf("ResolvePermissions(wsA): %v", err)
	}
	if len(forA) != 1 || forA[0].Key != "apps:systems:create" {
		t.Fatalf("ResolvePermissions(wsA) = %+v, want exactly permA", forA)
	}

	forB, err := auth.ResolvePermissions(userId, clientId, wsB)
	if err != nil {
		t.Fatalf("ResolvePermissions(wsB): %v", err)
	}
	if len(forB) != 1 || forB[0].Key != "apps:systems:delete" {
		t.Fatalf("ResolvePermissions(wsB) = %+v, want exactly permB — a token scoped to B must never carry A's permission", forB)
	}
}

// TestAuthService_ResolvePermissions_EmptyWorkspaceOnlySeesGlobalRows is TRD
// F6's own literal repro, now fixed: an empty workspaceId used to mean "no
// filter at all" (every workspace-scoped row for this user+client, from
// every workspace they ever touched) — it must now mean "global rows only"
// (workspace_id IS NULL), never a leak of a specific real workspace's rows.
func TestAuthService_ResolvePermissions_EmptyWorkspaceOnlySeesGlobalRows(t *testing.T) {
	db := identityTestDB(t)
	auth := pkg.NewAuthServiceForTest(db)
	permEntity := &identitypermission.PermissionEntity{Entity: (&sql.Entity[identitypermission.Permission]{}).With((&sql.Query{}).With(&sql.Db{DB: db}))}
	permEntity.OnRegister()

	userId := uuid.NewString()
	clientId := "c-" + uuid.NewString()
	ws := uuid.NewString()

	scoped := &identitypermission.Permission{
		UserId: &userId, ClientId: &clientId, WorkspaceId: &ws,
		Key: "apps:systems:create", Scope: "workspace", Resource: "systems", Action: "create",
	}
	global := &identitypermission.Permission{
		UserId: &userId, ClientId: &clientId, WorkspaceId: nil,
		Key: "platform:health:read", Scope: "global", Resource: "health", Action: "read",
	}
	if err := permEntity.Insert(scoped); err != nil {
		t.Fatalf("insert scoped: %v", err)
	}
	t.Cleanup(func() { _, _ = permEntity.Delete("id = ?", scoped.Id) })
	if err := permEntity.Insert(global); err != nil {
		t.Fatalf("insert global: %v", err)
	}
	t.Cleanup(func() { _, _ = permEntity.Delete("id = ?", global.Id) })

	got, err := auth.ResolvePermissions(userId, clientId, "")
	if err != nil {
		t.Fatalf("ResolvePermissions(\"\"): %v", err)
	}
	if len(got) != 1 || got[0].Key != "platform:health:read" {
		t.Fatalf("ResolvePermissions(\"\") = %+v, want exactly the global row, never the workspace-scoped one (TRD F6)", got)
	}
}

// TestAuthService_ResolveRoles_SameScoping is the same fix's other affected
// method: same real bug shape (empty workspaceId meant no filter at all),
// same expected repair.
func TestAuthService_ResolveRoles_SameScoping(t *testing.T) {
	db := identityTestDB(t)
	auth := pkg.NewAuthServiceForTest(db)
	roleEntity := &identityrole.RoleEntity{Entity: (&sql.Entity[identityrole.Role]{}).With((&sql.Query{}).With(&sql.Db{DB: db}))}
	roleEntity.OnRegister()

	clientId := "c-" + uuid.NewString()
	wsA := uuid.NewString()
	wsB := uuid.NewString()

	roleA := &identityrole.Role{Name: "owner-a", ClientId: clientId, WorkspaceId: &wsA, Type: "workspace"}
	roleB := &identityrole.Role{Name: "owner-b", ClientId: clientId, WorkspaceId: &wsB, Type: "workspace"}
	if err := roleEntity.Insert(roleA); err != nil {
		t.Fatalf("insert roleA: %v", err)
	}
	t.Cleanup(func() { _, _ = roleEntity.Delete("id = ?", roleA.Id) })
	if err := roleEntity.Insert(roleB); err != nil {
		t.Fatalf("insert roleB: %v", err)
	}
	t.Cleanup(func() { _, _ = roleEntity.Delete("id = ?", roleB.Id) })

	forA, err := auth.ResolveRoles("", clientId, wsA)
	if err != nil {
		t.Fatalf("ResolveRoles(wsA): %v", err)
	}
	if len(forA) != 1 || forA[0].Name != "owner-a" {
		t.Fatalf("ResolveRoles(wsA) = %+v, want exactly roleA", forA)
	}
}

// TestAuthService_Login_EmbedsTheGivenWorkspaceId is U-S9's other half:
// Login must actually embed whatever workspaceId its caller resolved, and
// the resulting token's claims — not just some out-of-band return value —
// must carry only that workspace's permissions when decoded and resolved
// again, matching exactly what identity/app/auth's real Login handler does
// with the tokens this method returns.
func TestAuthService_Login_EmbedsTheGivenWorkspaceId(t *testing.T) {
	db := identityTestDB(t)
	auth := pkg.NewAuthServiceForTest(db)
	userEntity := &user.UserEntity{Entity: (&sql.Entity[user.User]{}).With((&sql.Query{}).With(&sql.Db{DB: db}))}
	userEntity.OnRegister()
	permEntity := &identitypermission.PermissionEntity{Entity: (&sql.Entity[identitypermission.Permission]{}).With((&sql.Query{}).With(&sql.Db{DB: db}))}
	permEntity.OnRegister()

	hashed, err := security.Hash("correct horse battery staple")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	ref := "login-scoped-" + uuid.NewString()
	u := &user.User{Ref: &ref, Name: "Test User", Email: ref + "@example.test", Secret: &hashed}
	if err := userEntity.Insert(u); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	t.Cleanup(func() { _, _ = userEntity.Delete("id = ?", u.Id) })

	clientId := "c-" + uuid.NewString()
	wsA := uuid.NewString()
	permA := &identitypermission.Permission{
		UserId: &u.Id, ClientId: &clientId, WorkspaceId: &wsA,
		Key: "apps:systems:create", Scope: "workspace", Resource: "systems", Action: "create",
	}
	if err := permEntity.Insert(permA); err != nil {
		t.Fatalf("insert permA: %v", err)
	}
	t.Cleanup(func() { _, _ = permEntity.Delete("id = ?", permA.Id) })

	access, _, err := auth.Login(ref, "correct horse battery staple", clientId, wsA)
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	claims, err := auth.ValidateToken(access)
	if err != nil {
		t.Fatalf("ValidateToken: %v", err)
	}
	if claims["workspaceId"] != wsA {
		t.Fatalf(`claims["workspaceId"] = %v, want %q`, claims["workspaceId"], wsA)
	}
	perms, ok := claims["permissions"].([]any)
	if !ok || len(perms) != 1 || perms[0] != "apps:systems:create" {
		t.Fatalf(`claims["permissions"] = %v, want exactly ["apps:systems:create"]`, claims["permissions"])
	}
}
