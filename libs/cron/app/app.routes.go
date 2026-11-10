package app

import (
	"github.com/awesome-goose/goose/modules/router"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
)

// ROUTES mirrors ntx-apps/libs/cron/src/app.controller.ts: POST / registers a
// new job, GET / pops one from the pending queue, PATCH / writes a run log.
// TS does not expose a health/hello endpoint on the cron app.
//
// Guards (PLAN U26): register, select and log are the hooks a cron runner calls,
// so they take an internal call (coreauth.Internal: HMAC-signed, or a platform admin).
var ROUTES = router.ForRoutes(
	router.Post("/", []any{AppController{}, "Register"}, coreauth.Internal()...),
	router.Get("/", []any{AppController{}, "Select"}, coreauth.Internal()...),
	router.Patch("/", []any{AppController{}, "Log"}, coreauth.Internal()...),
)
