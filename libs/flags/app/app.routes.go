package app

import (
	"github.com/awesome-goose/goose/modules/router"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
)

// ROUTES mirrors ntx-apps/libs/flags/src/app.controller.ts route paths.
//
// Guards (PLAN U26): register rewrites flag definitions and log spends a quota, and
// status and limit evaluate for the user, client and workspace the body names, so all
// four are a service's calls (coreauth.Internal: HMAC-signed, or a platform admin).
var ROUTES = router.ForRoutes(
	router.Get("/", []any{AppController{}, "GetHello"}),
	router.Post("/register", []any{AppController{}, "Register"}, coreauth.Internal()...),
	router.Post("/log", []any{AppController{}, "Log"}, coreauth.Internal()...),
	router.Post("/status", []any{AppController{}, "Status"}, coreauth.Internal()...),
	router.Post("/limit", []any{AppController{}, "Limit"}, coreauth.Internal()...),
)
