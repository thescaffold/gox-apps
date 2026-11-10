package app

import (
	"github.com/awesome-goose/goose/modules/router"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
)

// ROUTES mirrors ntx-apps/libs/queue/src/app.controller.ts: POST / pushes a
// job, GET / pops one, PATCH / logs a completion. TS exposes no health
// endpoint on the queue app.
//
// Guards (PLAN U26): push, pop and log are a queue worker's hooks, and jobs carry
// other apps' payloads, so they take an internal call (HMAC-signed, or a platform admin).
var ROUTES = router.ForRoutes(
	router.Post("/", []any{AppController{}, "Push"}, coreauth.Internal()...),
	router.Get("/", []any{AppController{}, "Pop"}, coreauth.Internal()...),
	router.Patch("/", []any{AppController{}, "Log"}, coreauth.Internal()...),
)
