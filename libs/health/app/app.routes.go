package app

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors ntx-apps/libs/health/src/app.controller.ts route paths.
var ROUTES = router.ForRoutes(
	router.Get("/", []any{AppController{}, "Health"}),
	router.Post("/check/:id", []any{AppController{}, "Check"}),
	router.Post("/ping", []any{AppController{}, "Ping"}),
	router.Post("/register", []any{AppController{}, "Register"}),
)
