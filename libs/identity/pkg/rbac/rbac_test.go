package rbac

import (
	"testing"

	"github.com/awesome-goose/goose/types"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
)

func caller(user, workspace string, roles ...any) ntxctx.NTXContext {
	return ntxctx.NTXContext{UserID: user, WorkspaceID: workspace, Roles: roles}
}

func TestLevels(t *testing.T) {
	t.Setenv("ROOT_USER_ID", "root-user")
	cases := []struct {
		name            string
		ntx             ntxctx.NTXContext
		platform, admin bool
	}{
		{"anonymous", caller("", ""), false, false},
		{"anonymous with claimed roles", caller("", "w", "global-admin", "owner"), false, false},
		{"member", caller("u", "w", "member"), false, false},
		{"owner", caller("u", "w", "owner", "admin"), false, true},
		{"admin", caller("u", "w", "admin"), false, true},
		{"owner of no workspace", caller("u", "", "owner"), false, false},
		{"root user", caller("root-user", ""), true, false},
		{"global admin", caller("u", "w", "global-admin"), true, false},
		{"global_admin spelling", caller("u", "w", "Global_Admin"), true, false},
		{"global member is not an admin", caller("u", "w", "global-member"), false, false},
	}
	for _, c := range cases {
		if got := PlatformAdmin(c.ntx); got != c.platform {
			t.Errorf("%s: PlatformAdmin = %v", c.name, got)
		}
		if got := WorkspaceAdmin(c.ntx); got != c.admin {
			t.Errorf("%s: WorkspaceAdmin = %v", c.name, got)
		}
	}
}

func TestEmptyRootIDMatchesNobody(t *testing.T) {
	t.Setenv("ROOT_USER_ID", "")
	if PlatformAdmin(caller("", "")) {
		t.Fatal("an unset ROOT_USER_ID must not match an empty caller")
	}
}

func TestPrivileged(t *testing.T) {
	for _, r := range []string{"root", "ROOT", "global-admin", "global_admin", " Global-Member "} {
		if !Privileged(r) {
			t.Errorf("%q must be privileged", r)
		}
	}
	for _, r := range []string{"owner", "admin", "Admin", "member", "site_editor", ""} {
		if Privileged(r) {
			t.Errorf("%q must not be privileged", r)
		}
	}
}

func TestStrictIsTheDefault(t *testing.T) {
	t.Setenv("ROOT_USER_ID", "root-user")
	owner := caller("u", "w", "owner")
	admin := caller("u", "w", "admin")
	for _, on := range []string{"", "true", "TRUE", "1", "yes", "anything"} {
		t.Setenv("IDENTITY_RBAC_STRICT", on)
		if Allowed(owner) || Allowed(admin) {
			t.Fatalf("strict=%q must refuse a workspace owner or admin: role and permission tables are global", on)
		}
		if !Allowed(caller("root-user", "")) {
			t.Fatalf("strict=%q must keep the platform admin", on)
		}
	}
	// A console that still edits RBAC as a workspace admin opts out explicitly.
	for _, off := range []string{"false", "FALSE", "0"} {
		t.Setenv("IDENTITY_RBAC_STRICT", off)
		if !Allowed(owner) {
			t.Fatalf("strict=%q is the opt-out: a workspace owner keeps the RBAC screens", off)
		}
	}
}

type fakeRequest struct {
	types.Request
	method types.Method
}

func (r *fakeRequest) Method() types.Method { return r.method }

type fakeResponse struct {
	types.Response
	status int
}

func (r *fakeResponse) Write(_ types.SerialType, _ []byte, status int) error {
	r.status = status
	return nil
}

type fakeContext struct {
	req    *fakeRequest
	res    *fakeResponse
	values map[any]any
}

func (c *fakeContext) Request() types.Request   { return c.req }
func (c *fakeContext) Response() types.Response { return c.res }
func (c *fakeContext) SetValue(k any, v any)    { c.values[k] = v }
func (c *fakeContext) GetValue(k any) any       { return c.values[k] }

func handle(method types.Method, ntx ntxctx.NTXContext) (error, int) {
	ctx := &fakeContext{req: &fakeRequest{method: method}, res: &fakeResponse{}, values: map[any]any{}}
	ntxctx.Set(ctx, ntx)
	err := (&Guard{}).Handle(ctx)
	return err, ctx.res.status
}

func TestGuard(t *testing.T) {
	if err, _ := handle("GET", caller("u", "w", "member")); err != nil {
		t.Fatal("reads pass")
	}
	t.Setenv("ROOT_USER_ID", "root-user")
	t.Setenv("IDENTITY_RBAC_STRICT", "")
	if err, status := handle("POST", caller("u", "w", "owner")); err == nil || status != 403 {
		t.Fatalf("by default a workspace owner may not write global RBAC: %v %d", err, status)
	}
	if err, _ := handle("POST", caller("root-user", "")); err != nil {
		t.Fatal("the platform admin may write")
	}
	t.Setenv("IDENTITY_RBAC_STRICT", "false")
	if err, _ := handle("POST", caller("u", "w", "owner")); err != nil {
		t.Fatal("with strict off, an owner may write")
	}
	err, status := handle("POST", caller("u", "w", "member"))
	if err == nil || status != 403 {
		t.Fatalf("a member's write: %v %d", err, status)
	}
	if err, status := handle("DELETE", caller("", "")); err == nil || status != 403 {
		t.Fatalf("an unverified write: %v %d", err, status)
	}
}
