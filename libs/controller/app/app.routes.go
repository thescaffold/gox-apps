package app

import (
	"github.com/awesome-goose/goose/modules/router"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
)

// ROUTES mirrors ntx-apps/libs/controller/src/app.controller.ts — the gateway
// app exposes a hello endpoint, a route-register webhook and a /now clock.
// (RouteModule + RequestModule contribute their own /routes and /requests
// CRUD endpoints under the same `apps/controller` prefix.)
//
// Guards (PLAN U26): register is the hook a service calls to declare its routes, so it
// takes an internal call (HMAC-signed, or a platform admin); GET / and /now are public.
var ROUTES = router.ForRoutes(
	router.Get("/", []any{AppController{}, "GetHello"}),
	router.Post("/register", []any{AppController{}, "Register"}, coreauth.Internal()...),
	router.Get("/now", []any{AppController{}, "Now"}),
)
