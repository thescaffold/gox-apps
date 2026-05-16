package app

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors ntx-apps/libs/queue/src/app.controller.ts: POST / pushes a
// job, GET / pops one, PATCH / logs a completion. TS exposes no health
// endpoint on the queue app.
var ROUTES = router.ForRoutes(
	router.Post("/", []any{AppController{}, "Push"}),
	router.Get("/", []any{AppController{}, "Pop"}),
	router.Patch("/", []any{AppController{}, "Log"}),
)
