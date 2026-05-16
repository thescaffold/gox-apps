package app

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors ntx-apps/libs/controller/src/app.controller.ts — the gateway
// app exposes a hello endpoint, a route-register webhook and a /now clock.
// (RouteModule + RequestModule contribute their own /routes and /requests
// CRUD endpoints under the same `apps/controller` prefix.)
var ROUTES = router.ForRoutes(
	router.Get("/", []any{AppController{}, "GetHello"}),
	router.Post("/register", []any{AppController{}, "Register"}),
	router.Get("/now", []any{AppController{}, "Now"}),
)
