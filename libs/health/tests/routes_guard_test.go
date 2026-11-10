package tests

import (
	"strings"
	"testing"

	"github.com/awesome-goose/goose/types"
	health "github.com/thescaffold/gox-apps/libs/health/app"
	healthlog "github.com/thescaffold/gox-apps/libs/health/app/log"
	service "github.com/thescaffold/gox-apps/libs/health/app/service"
	summary "github.com/thescaffold/gox-apps/libs/health/app/summary"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
)

// PLAN U26: every route of the app has the guard it must have, so that mounted with no
// middleware of its own (as every product mounts it) anonymous and ordinary signed-in
// callers reach only what is meant for them.
func TestRouteGuards(t *testing.T) {
	want := map[string]string{}
	resource(want, "log", "admin")
	resource(want, "service", "admin")
	resource(want, "summary", "admin")
	want["GET /"] = "public"
	want["POST /ping"] = "internal"
	want["POST /register"] = "internal" // the server would poll the URL it names
	checkGuards(t, routeGuards(t, health.ROUTES, healthlog.ROUTES, service.ROUTES, summary.ROUTES), want)
}

// guardOf names what a route's middleware chain (its own and its parents') demands:
// internal (a signed service call or a platform admin), admin, session, or public.
func guardOf(chain []types.Middleware) string {
	has := func(match func(types.Middleware) bool) bool {
		for _, m := range chain {
			if match(m) {
				return true
			}
		}
		return false
	}
	switch {
	case has(func(m types.Middleware) bool { _, ok := m.(*coreauth.InternalGuard); return ok }):
		return "internal"
	case has(func(m types.Middleware) bool { _, ok := m.(*coreauth.AdminGuard); return ok }):
		return "admin"
	case has(func(m types.Middleware) bool { _, ok := m.(*coreauth.AuthMiddleware); return ok }):
		if has(func(m types.Middleware) bool { _, ok := m.(*coreauth.FromClaims); return ok }) {
			return "caller"
		}
		return "session"
	}
	return "public"
}

func routeGuards(t *testing.T, mods ...types.Module) map[string]string {
	t.Helper()
	out := map[string]string{}
	var walk func(rs types.Routes, parent string, inherited []types.Middleware)
	walk = func(rs types.Routes, parent string, inherited []types.Middleware) {
		for _, r := range rs {
			path := strings.Trim(parent+"/"+strings.Trim(r.Path, "/"), "/")
			chain := append(append([]types.Middleware{}, inherited...), r.Middlewares...)
			if r.Handler != nil {
				out[strings.ToUpper(r.Method.String())+" /"+path] = guardOf(chain)
			}
			walk(r.Children, path, chain)
		}
	}
	for _, m := range mods {
		r, ok := m.(interface{ Routes() (types.Routes, error) })
		if !ok {
			t.Fatalf("%T has no route table", m)
		}
		rs, err := r.Routes()
		if err != nil {
			t.Fatal(err)
		}
		walk(rs, "", nil)
	}
	return out
}

// resource is the five routes of router.Resource(name).All(), all behind one guard.
func resource(want map[string]string, name, guard string) {
	for _, k := range []string{"POST /" + name, "GET /" + name, "GET /" + name + "/:id", "PATCH /" + name + "/:id", "DELETE /" + name + "/:id"} {
		want[k] = guard
	}
}

// adminWrites is a resource a verified caller may read and only a platform admin may write.
func adminWrites(want map[string]string, name string) {
	resource(want, name, "admin")
	want["GET /"+name] = "caller"
	want["GET /"+name+"/:id"] = "caller"
}

func checkGuards(t *testing.T, got, want map[string]string) {
	t.Helper()
	for k, w := range want {
		if g, ok := got[k]; !ok {
			t.Errorf("%s: no such route", k)
		} else if g != w {
			t.Errorf("%s is %s, want %s", k, g, w)
		}
	}
	for k, g := range got {
		if _, ok := want[k]; !ok {
			t.Errorf("%s (%s) is not in the expected table: decide its guard", k, g)
		}
	}
}
