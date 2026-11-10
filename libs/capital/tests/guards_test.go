package tests

import (
	"strings"
	"testing"

	"github.com/awesome-goose/goose/types"
	"github.com/thescaffold/gox-apps/libs/capital/app"
	capitalaccount "github.com/thescaffold/gox-apps/libs/capital/app/account"
	capitalpayment "github.com/thescaffold/gox-apps/libs/capital/app/payment"
	capitalpaymentlog "github.com/thescaffold/gox-apps/libs/capital/app/paymentlog"
	capitalplan "github.com/thescaffold/gox-apps/libs/capital/app/plan"
	capitalplantype "github.com/thescaffold/gox-apps/libs/capital/app/plantype"
	capitalprovider "github.com/thescaffold/gox-apps/libs/capital/app/provider"
	capitalrate "github.com/thescaffold/gox-apps/libs/capital/app/rate"
	capitaltransaction "github.com/thescaffold/gox-apps/libs/capital/app/transaction"
	capitalusage "github.com/thescaffold/gox-apps/libs/capital/app/usage"
	capitalvoucher "github.com/thescaffold/gox-apps/libs/capital/app/voucher"
	capitalvouchertype "github.com/thescaffold/gox-apps/libs/capital/app/vouchertype"
	capitalwallet "github.com/thescaffold/gox-apps/libs/capital/app/wallet"
	flagsapp "github.com/thescaffold/gox-apps/libs/flags/app/flag"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
)

// PLAN U26: every capital route says who may call it. Before, the app routes, the wallet and the
// voucher list/verify answered anonymous callers, and every resource let any signed-in caller
// read and write every tenant's rows.

type routed interface{ Routes() (types.Routes, error) }

// guardOf names the guard a middleware chain applies: public, caller, admin or internal.
func guardOf(mws []types.Middleware) string {
	has := map[string]bool{}
	for _, m := range mws {
		switch m.(type) {
		case *coreauth.AuthMiddleware:
			has["session"] = true
		case *coreauth.FromClaims:
			has["claims"] = true
		case *coreauth.AdminGuard:
			has["admin"] = true
		case *coreauth.InternalGuard:
			has["internal"] = true
		}
	}
	switch {
	case has["internal"]:
		return "internal"
	case has["admin"] && has["session"] && has["claims"]:
		return "admin"
	case has["session"] && has["claims"]:
		return "caller"
	case has["session"]:
		return "session-only"
	case len(mws) == 0:
		return "public"
	}
	return "other"
}

// guards maps "METHOD /path Handler" to its guard, across nested routes.
func guards(t *testing.T, mods ...routed) map[string]string {
	t.Helper()
	out := map[string]string{}
	var walk func(rs types.Routes, prefix string, inherited []types.Middleware)
	walk = func(rs types.Routes, prefix string, inherited []types.Middleware) {
		for _, r := range rs {
			path := strings.TrimRight(prefix, "/") + "/" + strings.Trim(r.Path, "/")
			chain := append(append([]types.Middleware{}, inherited...), r.Middlewares...)
			if tuple, ok := r.Handler.([]any); ok && len(tuple) == 2 {
				name, _ := tuple[1].(string)
				p := strings.TrimRight(path, "/")
				if p == "" {
					p = "/"
				}
				out[strings.ToUpper(string(r.Method))+" "+p+" "+name] = guardOf(chain)
			}
			walk(r.Children, path, chain)
		}
	}
	for _, m := range mods {
		rs, err := m.Routes()
		if err != nil {
			t.Fatal(err)
		}
		walk(rs, "", nil)
	}
	return out
}

func expectGuards(t *testing.T, got map[string]string, want map[string]string) {
	t.Helper()
	for route, guard := range want {
		g, ok := got[route]
		if !ok {
			t.Errorf("%s: no such route (have %v)", route, keys(got))
			continue
		}
		if g != guard {
			t.Errorf("%s is %s, want %s", route, g, guard)
		}
	}
}

func keys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func TestAppRoutesGuards(t *testing.T) {
	got := guards(t, app.ROUTES.(routed))
	want := map[string]string{
		"GET / GetHello":                                "public",
		"GET /providers GetProviders":                   "public",
		"GET /debt Debt":                                "caller",
		"GET /accrual Accrual":                          "caller",
		"GET /status Status":                            "caller",
		"GET /user-plan UserPlan":                       "caller",
		"POST /default-provider GetDefaultProvider":     "caller",
		"POST /init Init":                               "caller",
		"POST /pay Pay":                                 "caller",
		"POST /user-plan/compute-upgrade ComputeUpgrade": "caller",
		"POST /verify/:type Verify":                     "caller",
		"GET /:type/reference/:reference Reference":     "caller",
		"POST /:type/charge/:providerId Charge":         "caller",
		// these name the user in the body: another service or an admin only
		"POST /usage/start UsageStart":       "internal",
		"POST /usage/update UsageUpdate":     "internal",
		"POST /usage/stop UsageStop":         "internal",
		"POST /plan/change PlanChange":       "internal",
		"POST /plan/subscribe PlanSubscribe": "internal",
	}
	expectGuards(t, got, want)
	if len(got) != len(want) {
		t.Errorf("%d app routes, %d checked: a new route needs a guard decision", len(got), len(want))
	}
}

func TestWalletAndVoucherRoutesNeedTheCaller(t *testing.T) {
	got := guards(t, capitalwallet.ROUTES.(routed), capitalvoucher.ROUTES.(routed))
	for route, g := range got {
		if strings.Contains(route, "/wallet/") || strings.Contains(route, "ListAvailable") || strings.Contains(route, "voucher/verify") {
			if g != "caller" {
				t.Errorf("%s is %s, want caller", route, g)
			}
		}
	}
}

func resourceGuards(t *testing.T, m routed) map[string]string {
	t.Helper()
	out := map[string]string{}
	for route, g := range guards(t, m) {
		parts := strings.Fields(route)
		out[parts[len(parts)-1]] = g
	}
	return out
}

func TestResourceGuards(t *testing.T) {
	adminWrites := map[string]string{"List": "caller", "Get": "caller", "Create": "admin", "Update": "admin", "Delete": "admin"}
	adminOnly := map[string]string{"List": "admin", "Get": "admin", "Create": "admin", "Update": "admin", "Delete": "admin"}
	cases := map[string]struct {
		mod  routed
		want map[string]string
	}{
		// catalog: anyone signed in reads, an admin writes
		"rate":      {capitalrate.ROUTES.(routed), adminWrites},
		"plan-type": {capitalplantype.ROUTES.(routed), adminWrites},
		// a person's own ledger rows (scoped in the controllers): read yes, write never
		"account":     {capitalaccount.ROUTES.(routed), adminWrites},
		"payment":     {capitalpayment.ROUTES.(routed), adminWrites},
		"plan":        {capitalplan.ROUTES.(routed), adminWrites},
		"usage":       {capitalusage.ROUTES.(routed), adminWrites},
		"transaction": {capitaltransaction.ROUTES.(routed), adminWrites},
		// raw provider request/response logs and voucher-type redeem tokens: admin only
		"payment-log":  {capitalpaymentlog.ROUTES.(routed), adminOnly},
		"voucher-type": {capitalvouchertype.ROUTES.(routed), adminOnly},
		// saved cards: the owner manages their own, creating one by hand is an admin's
		"provider": {capitalprovider.ROUTES.(routed), map[string]string{"List": "caller", "Get": "caller", "Create": "admin", "Update": "caller", "Delete": "caller"}},
	}
	for name, c := range cases {
		got := resourceGuards(t, c.mod)
		for h, want := range c.want {
			if got[h] != want {
				t.Errorf("%s %s is %q, want %q", name, h, got[h], want)
			}
		}
	}
	voucher := resourceGuards(t, capitalvoucher.ROUTES.(routed))
	for h, want := range adminWrites {
		if voucher[h] != want {
			t.Errorf("voucher %s is %q, want %q", h, voucher[h], want)
		}
	}
}

// The flags app's flag resource used to be served under /apps/capital/flag, unguarded.
func TestFlagsModuleIsServicesOnly(t *testing.T) {
	for _, m := range (&app.AppModule{}).Imports() {
		if _, whole := m.(*flagsapp.FlagModule); whole {
			t.Fatal("capital imports the flags FlagModule whole, serving its CRUD under /apps/capital/flag")
		}
	}
}
