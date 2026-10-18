package tests

import (
	"sort"
	"testing"

	"github.com/awesome-goose/goose/modules/sql"
	"github.com/google/uuid"
	"github.com/thescaffold/gox-apps/libs/identity/app"
	identityucw "github.com/thescaffold/gox-apps/libs/identity/app/userclientworkspace"
	"github.com/thescaffold/gox-apps/libs/identity/pkg"
)

func keysOf(t *testing.T, a *pkg.AuthService, user, client, ws string) []string {
	t.Helper()
	ps, err := a.ResolvePermissions(user, client, ws)
	if err != nil {
		t.Fatalf("ResolvePermissions: %v", err)
	}
	out := make([]string, 0, len(ps))
	for _, p := range ps {
		out = append(out, p.Key)
	}
	sort.Strings(out)
	return out
}

func eq(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// A registered role's grants reach exactly the members of a workspace who hold
// that role — never a non-member, never another workspace's members.
func TestRolePermissions_FollowMembership(t *testing.T) {
	db := identityTestDB(t)
	auth := pkg.NewAuthServiceForTest(db)
	svc := app.NewAppServiceForTest(db)
	ucw := &identityucw.UserClientWorkspaceEntity{Entity: (&sql.Entity[identityucw.UserClientWorkspace]{}).With((&sql.Query{}).With(&sql.Db{DB: db}))}
	ucw.OnRegister()

	role := "r" + uuid.NewString()[:8] // unique role names keep the shared test DB clean
	adminRole := "a" + uuid.NewString()[:8]
	if err := svc.OnRBACRegister(app.RBACRegisterPayload{RolePermissions: map[string][]string{
		role:      {role + ":read:apps:files:*:*:workspace"},
		adminRole: {adminRole + ":*:apps:files:*:*:workspace"},
	}}); err != nil {
		t.Fatal(err)
	}

	client := "c-" + uuid.NewString()
	wsA, wsB := uuid.NewString(), uuid.NewString()
	member, owner, stranger := uuid.NewString(), uuid.NewString(), uuid.NewString()
	add := func(u, ws, status string) {
		s := status
		if err := ucw.Insert(&identityucw.UserClientWorkspace{UserId: u, ClientId: client, WorkspaceId: ws, Status: &s}); err != nil {
			t.Fatal(err)
		}
	}
	add(member, wsA, role)
	add(owner, wsA, "owner")
	add(member, wsB, "unlisted")

	want := []string{role + ":read:apps:files:*:*:workspace"}
	if got := keysOf(t, auth, member, client, wsA); !eq(got, want) {
		t.Fatalf("member in A = %v, want %v", got, want)
	}
	if got := keysOf(t, auth, member, client, wsB); len(got) != 0 {
		t.Fatalf("member's role in A leaked into B: %v", got)
	}
	if got := keysOf(t, auth, stranger, client, wsA); len(got) != 0 {
		t.Fatalf("non-member of A got grants: %v", got)
	}
	if got := keysOf(t, auth, member, client, ""); len(got) != 0 {
		t.Fatalf("no-workspace session got grants: %v", got)
	}

	// An owner carries the admin grants too.
	if err := svc.OnRBACRegister(app.RBACRegisterPayload{RolePermissions: map[string][]string{
		"admin": {"admin:*:apps:" + role + ":*:*:workspace"},
	}}); err != nil {
		t.Fatal(err)
	}
	if got := keysOf(t, auth, owner, client, wsA); !eq(got, []string{"admin:*:apps:" + role + ":*:*:workspace"}) {
		t.Fatalf("owner in A = %v, want the admin grant", got)
	}

	// Re-registering replaces the list: dropped grants stop being issued.
	if err := svc.OnRBACRegister(app.RBACRegisterPayload{RolePermissions: map[string][]string{
		role: {role + ":create:apps:files:*:*:workspace"},
	}}); err != nil {
		t.Fatal(err)
	}
	want = []string{role + ":create:apps:files:*:*:workspace"}
	if got := keysOf(t, auth, member, client, wsA); !eq(got, want) {
		t.Fatalf("after re-register = %v, want %v", got, want)
	}
}
