package app

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors ntx-apps/libs/health/src/app.controller.ts: GET / is hello,
// POST /ping writes a Log + bumps the Service state, POST /register upserts a
// Service row. TS exposes no /check/:id endpoint, so that route is omitted.
var ROUTES = router.ForRoutes(
	router.Get("/", []any{AppController{}, "GetHello"}),
	router.Post("/ping", []any{AppController{}, "Ping"}),
	router.Post("/register", []any{AppController{}, "Register"}),
)
