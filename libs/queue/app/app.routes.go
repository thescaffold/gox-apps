package app

import (
	"github.com/awesome-goose/goose/modules/router"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
)

// ROUTES mirrors ntx-apps/libs/queue/src/app.controller.ts: POST / pushes a
// job, GET / pops one, PATCH / logs a completion. TS exposes no health
// endpoint on the queue app.
var ROUTES = router.ForRoutes(
	router.Post("/", []any{AppController{}, "Push"}, coreauth.Session()...),
	router.Get("/", []any{AppController{}, "Pop"}, coreauth.Session()...),
	router.Patch("/", []any{AppController{}, "Log"}, coreauth.Session()...),
)
