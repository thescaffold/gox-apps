package app

import (
	"github.com/awesome-goose/goose/modules/router"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
)

// ROUTES mirrors ntx-apps/libs/health/src/app.controller.ts: GET / is hello,
// POST /ping writes a Log + bumps the Service state, POST /register upserts a
// Service row. TS exposes no /check/:id endpoint, so that route is omitted.
//
// Guards (PLAN U26): register makes the server poll the URL it names and ping
// records a service's state, so both take an internal call (HMAC-signed, or a
// platform admin); GET / is public.
var ROUTES = router.ForRoutes(
	router.Get("/", []any{AppController{}, "GetHello"}),
	router.Post("/ping", []any{AppController{}, "Ping"}, coreauth.Internal()...),
	router.Post("/register", []any{AppController{}, "Register"}, coreauth.Internal()...),
)
