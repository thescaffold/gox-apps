package tests

import (
	"strings"
	"testing"

	"github.com/awesome-goose/goose/types"
	"github.com/thescaffold/gox-apps/libs/identity/app/attribute"
	"github.com/thescaffold/gox-apps/libs/identity/app/client"
	"github.com/thescaffold/gox-apps/libs/identity/app/clientlog"
	"github.com/thescaffold/gox-apps/libs/identity/app/device"
	"github.com/thescaffold/gox-apps/libs/identity/app/devicelog"
	"github.com/thescaffold/gox-apps/libs/identity/app/devicesession"
	"github.com/thescaffold/gox-apps/libs/identity/app/invite"
	"github.com/thescaffold/gox-apps/libs/identity/app/provider"
	"github.com/thescaffold/gox-apps/libs/identity/app/providerlog"
	"github.com/thescaffold/gox-apps/libs/identity/app/token"
	"github.com/thescaffold/gox-apps/libs/identity/app/user"
	"github.com/thescaffold/gox-apps/libs/identity/app/userclientworkspace"
	"github.com/thescaffold/gox-apps/libs/identity/app/workspace"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
)

// The generic resources over many people's rows take any id, so they are a platform admin's
// tool (PLAN U26, ported from Yan's fork, lane 2 3f69b49): before, any signed-in person could
// read, edit or delete another account (PATCH /user/:id), make themselves owner of anyone's
// workspace (POST /user-client-workspace), mint tokens for another user (POST /token/user), and
// anyone could list provider rows.
func TestManyPeoplesResourcesNeedAPlatformAdmin(t *testing.T) {
	tables := map[string]types.Module{
		"attribute":             attribute.ROUTES,
		"client":                client.ROUTES,
		"client-log":            clientlog.ROUTES,
		"device":                device.ROUTES,
		"device-log":            devicelog.ROUTES,
		"device-session":        devicesession.ROUTES,
		"invite":                invite.ROUTES,
		"provider":              provider.ROUTES,
		"provider-log":          providerlog.ROUTES,
		"token":                 token.ROUTES,
		"user":                  user.ROUTES,
		"user-client-workspace": userclientworkspace.ROUTES,
		"workspace":             workspace.ROUTES,
	}
	isAdmin := func(mws []types.Middleware) bool {
		for _, m := range mws {
			if _, ok := m.(*coreauth.AdminGuard); ok {
				return coreauth.RequiresSession(mws)
			}
		}
		return false
	}
	for name, mod := range tables {
		withRoutes, ok := mod.(interface{ Routes() (types.Routes, error) })
		if !ok {
			t.Fatalf("%s: %T has no route table", name, mod)
		}
		routes, err := withRoutes.Routes()
		if err != nil {
			t.Fatal(err)
		}
		resources := 0
		for _, r := range routes {
			if strings.Trim(r.Path, "/") == "token/user" && !isAdmin(r.Middlewares) {
				t.Errorf("POST /token/user mints a token for any user and must need a platform admin")
			}
			if strings.Trim(r.Path, "/") != name || len(r.Children) == 0 {
				continue
			}
			resources++
			for _, c := range r.Children {
				if !isAdmin(c.Middlewares) {
					t.Errorf("%s %s%s is not platform-admin only", c.Method.String(), name, c.Path)
				}
			}
		}
		if resources != 1 {
			t.Errorf("%s: found %d resource routes, want 1", name, resources)
		}
	}
}
