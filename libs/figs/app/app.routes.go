package app

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors ntx-apps/libs/figs/src/app.controller.ts route paths:
//
//	GET    /          → GetHello
//	POST   /:name     → SaveFile
var ROUTES = router.ForRoutes(
	router.Get("/", []any{AppController{}, "GetHello"}),
	router.Post("/:name", []any{AppController{}, "SaveFile"}),
)
