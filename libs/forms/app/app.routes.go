package app

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors ntx-apps/libs/forms/src/app.controller.ts route paths.
var ROUTES = router.ForRoutes(
	router.Get("/", []any{AppController{}, "Health"}),
	router.Get("/one/:id", []any{AppController{}, "GetForm"}),
	router.Post("/one/:id", []any{AppController{}, "SaveForm"}),
)
