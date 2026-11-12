package tests

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	apikeyapp "github.com/thescaffold/gox-apps/libs/identity/app/apikey"
	"github.com/thescaffold/gox-apps/libs/identity/pkg/apikey"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
)

const day = 24 * time.Hour

// V0-14 (PLT-07): a host that requires expiry (MaxTTL set) refuses a key with none, or one past
// the workspace's maximum (a year unless the workspace set a shorter one). Without MaxTTL a key
// may still have no expiry, as before.
func TestApiKey_ExpiryRequiredUpToTheWorkspaceMaximum(t *testing.T) {
	e := newOrgEnv(t)
	ctx := context.Background()
	ws := newWS()
	svc := &apikey.Service{DB: e.db, Brand: "lgr", MaxTTL: apikey.WorkspaceMaxTTL(e.db)}

	if _, err := svc.Issue(ctx, "u1", "c1", ws, "none", apikey.Secret, apikey.Live, nil, 0); !errors.Is(err, apikey.ErrExpiryRequired) {
		t.Fatalf("no expiry: %v", err)
	}
	if _, err := svc.Issue(ctx, "u1", "c1", ws, "long", apikey.Secret, apikey.Live, nil, 366*day); !errors.Is(err, apikey.ErrExpiryTooLong) {
		t.Fatalf("past a year: %v", err)
	}
	if _, err := svc.Issue(ctx, "u1", "c1", ws, "year", apikey.Secret, apikey.Live, nil, 365*day); err != nil {
		t.Fatalf("a year: %v", err)
	}

	// The workspace lowers its maximum to 30 days.
	if err := apikey.SetMaxDays(ctx, e.db, ws, "u1", 30); err != nil {
		t.Fatal(err)
	}
	if got, _ := apikey.MaxDays(ctx, e.db, ws); got != 30 {
		t.Fatalf("max days = %d", got)
	}
	if _, err := svc.Issue(ctx, "u1", "c1", ws, "90", apikey.Secret, apikey.Live, nil, 90*day); !errors.Is(err, apikey.ErrExpiryTooLong) {
		t.Fatalf("90 days past a 30-day maximum: %v", err)
	}
	if _, err := svc.Issue(ctx, "u1", "c1", ws, "30", apikey.Secret, apikey.Live, nil, 30*day); err != nil {
		t.Fatalf("30 days: %v", err)
	}
	// Another workspace keeps the default.
	if got, _ := apikey.MaxDays(ctx, e.db, newWS()); got != apikey.DefaultMaxDays {
		t.Fatalf("default max days = %d", got)
	}
	for _, bad := range []int{0, -1, apikey.DefaultMaxDays + 1} {
		if err := apikey.SetMaxDays(ctx, e.db, ws, "u1", bad); !errors.Is(err, apikey.ErrInvalid) {
			t.Fatalf("max days %d: %v", bad, err)
		}
	}

	// Rotation keeps the old key's life, so it is held to the same rule.
	k, _ := svc.Issue(ctx, "u1", "c1", ws, "rot", apikey.Secret, apikey.Live, nil, 10*day)
	if _, err := svc.RotateIn(ctx, "u1", "c1", ws, k.ID, 0); err != nil {
		t.Fatalf("rotate: %v", err)
	}

	// Without MaxTTL nothing changes: a key may have no expiry.
	if _, err := (&apikey.Service{DB: e.db, Brand: "lgr"}).Issue(ctx, "u1", "c1", ws, "free", apikey.Secret, apikey.Live, nil, 0); err != nil {
		t.Fatalf("without MaxTTL: %v", err)
	}
}

// POST /keys answers 422 for a key with no expiry or past the maximum; owners and admins
// (org.security.manage) read and set the maximum at /keys/policy, everyone else reads it.
func TestApiKeyController_ExpiryIs422AndThePolicyIsTheAdmins(t *testing.T) {
	e := newOrgEnv(t)
	ws := newWS()
	owner := e.seat(t, ws, "Exp Owner", "owner")
	member := e.seat(t, ws, "Exp Member", "member")
	as := func(u string) ntxctx.NTXContext {
		return ntxctx.NTXContext{UserID: u, ClientID: "org-client", WorkspaceID: ws}
	}
	c := apikeyapp.NewControllerForTest(&apikey.Service{DB: e.db, Brand: "lgr", MaxTTL: apikey.WorkspaceMaxTTL(e.db)})

	code, b := kbody(c.Issue(&apikeyapp.IssueReq{Name: "ci", Environment: "live", Ctx: as(member)}))
	if code != 422 || !strings.Contains(b, `"code":"expiry_required"`) || !strings.Contains(b, `"maxDays":365`) {
		t.Fatalf("no expiry: %d %s", code, b)
	}
	code, b = kbody(c.Issue(&apikeyapp.IssueReq{Name: "ci", Environment: "live", ExpiresIn: 400 * 86400, Ctx: as(member)}))
	if code != 422 || !strings.Contains(b, `"code":"expiry_too_long"`) {
		t.Fatalf("400 days: %d %s", code, b)
	}
	if code, b := kbody(c.Issue(&apikeyapp.IssueReq{Name: "ci", Environment: "live", ExpiresIn: 90 * 86400, Ctx: as(member)})); code != 201 {
		t.Fatalf("90 days: %d %s", code, b)
	}

	if code, b := kbody(c.Policy(&apikeyapp.ListReq{Ctx: as(member)})); code != 200 || !strings.Contains(b, `"maxDays":365`) || !strings.Contains(b, `"required":true`) || !strings.Contains(b, `"canManage":false`) {
		t.Fatalf("a member reads the policy: %d %s", code, b)
	}
	if code, _ := kbody(c.SetPolicy(&apikeyapp.PolicyReq{MaxDays: 30, Ctx: as(member)})); code != 403 {
		t.Fatalf("a member sets the policy: %d", code)
	}
	if code, b := kbody(c.SetPolicy(&apikeyapp.PolicyReq{MaxDays: 400, Ctx: as(owner)})); code != 422 {
		t.Fatalf("400 days as the maximum: %d %s", code, b)
	}
	if code, b := kbody(c.SetPolicy(&apikeyapp.PolicyReq{MaxDays: 30, Ctx: as(owner)})); code != 200 || !strings.Contains(b, `"maxDays":30`) {
		t.Fatalf("the owner sets 30 days: %d %s", code, b)
	}
	if code, b := kbody(c.Issue(&apikeyapp.IssueReq{Name: "ci", Environment: "live", ExpiresIn: 90 * 86400, Ctx: as(member)})); code != 422 || !strings.Contains(b, `"maxDays":30`) {
		t.Fatalf("90 days past the new maximum: %d %s", code, b)
	}
}
