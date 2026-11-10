package tests

import (
	"testing"

	"github.com/awesome-goose/goose/types"
	"github.com/thescaffold/gox-apps/libs/common/app"
	commonip "github.com/thescaffold/gox-apps/libs/common/app/ip"
	commonproject "github.com/thescaffold/gox-apps/libs/common/app/project"
	commonprojecttype "github.com/thescaffold/gox-apps/libs/common/app/projecttype"
	commonrate "github.com/thescaffold/gox-apps/libs/common/app/rate"
	commonratelog "github.com/thescaffold/gox-apps/libs/common/app/ratelog"
	commontag "github.com/thescaffold/gox-apps/libs/common/app/tag"
	commontagtype "github.com/thescaffold/gox-apps/libs/common/app/tagtype"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
)

type routed interface{ Routes() (types.Routes, error) }

func guardOf(mws []types.Middleware) string {
	var session, claims, admin bool
	for _, m := range mws {
		switch m.(type) {
		case *coreauth.AuthMiddleware:
			session = true
		case *coreauth.FromClaims:
			claims = true
		case *coreauth.AdminGuard:
			admin = true
		}
	}
	switch {
	case admin && session && claims:
		return "admin"
	case session && claims:
		return "caller"
	case session:
		return "session-only"
	case len(mws) == 0:
		return "public"
	}
	return "other"
}

// handlerGuards maps each handler method of a route module to its guard.
func handlerGuards(t *testing.T, m routed) map[string]string {
	t.Helper()
	rs, err := m.Routes()
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]string{}
	var walk func(rs types.Routes, inherited []types.Middleware)
	walk = func(rs types.Routes, inherited []types.Middleware) {
		for _, r := range rs {
			chain := append(append([]types.Middleware{}, inherited...), r.Middlewares...)
			if tuple, ok := r.Handler.([]any); ok && len(tuple) == 2 {
				out[tuple[1].(string)] = guardOf(chain)
			}
			walk(r.Children, chain)
		}
	}
	walk(rs, nil)
	return out
}

// PLAN U26: every common route says who may call it.
func TestCommonRouteGuards(t *testing.T) {
	all := func(g string) map[string]string {
		return map[string]string{"List": g, "Get": g, "Create": g, "Update": g, "Delete": g}
	}
	adminWrites := map[string]string{"List": "caller", "Get": "caller", "Create": "admin", "Update": "admin", "Delete": "admin"}
	links := map[string]string{"Sync": "caller", "Attach": "caller", "Detach": "caller"}
	merge := func(a, b map[string]string) map[string]string {
		out := map[string]string{}
		for k, v := range a {
			out[k] = v
		}
		for k, v := range b {
			out[k] = v
		}
		return out
	}
	cases := map[string]struct {
		mod  routed
		want map[string]string
	}{
		"app":          {app.ROUTES.(routed), map[string]string{"GetHello": "public", "GetCurrency": "public", "GetLocation": "public"}},
		"ip":           {commonip.ROUTES.(routed), all("admin")},
		"rate-log":     {commonratelog.ROUTES.(routed), all("admin")},
		"rate":         {commonrate.ROUTES.(routed), adminWrites},
		"project":      {commonproject.ROUTES.(routed), all("caller")},
		"tag":          {commontag.ROUTES.(routed), all("caller")},
		"project-type": {commonprojecttype.ROUTES.(routed), merge(all("caller"), links)},
		"tag-type":     {commontagtype.ROUTES.(routed), merge(all("caller"), links)},
	}
	for name, c := range cases {
		got := handlerGuards(t, c.mod)
		for h, want := range c.want {
			if got[h] != want {
				t.Errorf("%s %s is %q, want %q", name, h, got[h], want)
			}
		}
		if len(got) != len(c.want) {
			t.Errorf("%s: %d handlers routed, %d checked", name, len(got), len(c.want))
		}
	}
}
