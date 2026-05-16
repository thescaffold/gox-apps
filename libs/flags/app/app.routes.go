package app

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors ntx-apps/libs/flags/src/app.controller.ts route paths.
var ROUTES = router.ForRoutes(
	router.Get("/", []any{AppController{}, "GetHello"}),
	router.Post("/register", []any{AppController{}, "Register"}),
	router.Post("/log", []any{AppController{}, "Log"}),
	router.Post("/status", []any{AppController{}, "Status"}),
	router.Post("/limit", []any{AppController{}, "Limit"}),
)
