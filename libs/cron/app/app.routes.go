package app

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors ntx-apps/libs/cron/src/app.controller.ts: POST / registers a
// new job, GET / pops one from the pending queue, PATCH / writes a run log.
// TS does not expose a health/hello endpoint on the cron app.
var ROUTES = router.ForRoutes(
	router.Post("/", []any{AppController{}, "Register"}),
	router.Get("/", []any{AppController{}, "Select"}),
	router.Patch("/", []any{AppController{}, "Log"}),
)
