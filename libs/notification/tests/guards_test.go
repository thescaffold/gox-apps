package tests

import (
	"testing"

	"github.com/awesome-goose/goose/types"
	"github.com/thescaffold/gox-apps/libs/notification/app"
	notificationlog "github.com/thescaffold/gox-apps/libs/notification/app/log"
	notificationmessage "github.com/thescaffold/gox-apps/libs/notification/app/message"
	notificationrule "github.com/thescaffold/gox-apps/libs/notification/app/rule"
	notificationtemplate "github.com/thescaffold/gox-apps/libs/notification/app/template"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
)

// Only GET / and the token-guarded capture are open; rules and templates need a platform
// admin, and log writes too (PLAN U26).
func TestNotificationRoutesGuards(t *testing.T) {
	open := map[string]bool{"/": true, "/capture": true}
	hasAdmin := func(mws []types.Middleware) bool {
		for _, m := range mws {
			if _, ok := m.(*coreauth.AdminGuard); ok {
				return true
			}
		}
		return false
	}
	for m, adminFor := range map[types.Module]map[string]bool{
		app.ROUTES:                  {},
		notificationmessage.ROUTES:  {},
		notificationlog.ROUTES:      {"Create": true, "Update": true, "Delete": true},
		notificationrule.ROUTES:     {"Create": true, "List": true, "Get": true, "Update": true, "Delete": true},
		notificationtemplate.ROUTES: {"Create": true, "List": true, "Get": true, "Update": true, "Delete": true},
	} {
		rs, err := m.(interface{ Routes() (types.Routes, error) }).Routes()
		if err != nil {
			t.Fatal(err)
		}
		var walk func(rs types.Routes, inherited []types.Middleware, path string)
		walk = func(rs types.Routes, inherited []types.Middleware, path string) {
			for _, r := range rs {
				chain := append(append([]types.Middleware{}, inherited...), r.Middlewares...)
				full := path + r.Path
				if r.Handler != nil {
					if !open[full] && !coreauth.RequiresSession(chain) {
						t.Errorf("%s is open to anonymous callers", full)
					}
					if name := r.Handler.([]any)[1].(string); adminFor[name] && !hasAdmin(chain) {
						t.Errorf("%s %s should need a platform admin", name, full)
					}
				}
				walk(r.Children, chain, full)
			}
		}
		walk(rs, nil, "")
	}
}
