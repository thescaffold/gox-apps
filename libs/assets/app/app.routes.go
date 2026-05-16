package app

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors ntx-apps/libs/assets/src/app.controller.ts: GET / returns
// getHello, GET /dynamic generates or looks up a dynamic SVG.
var ROUTES = router.ForRoutes(
	router.Get("/", []any{AppController{}, "GetHello"}),
	router.Get("/dynamic", []any{AppController{}, "DynamicFile"}),
)
