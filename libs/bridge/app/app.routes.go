package app

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors ntx-apps/libs/bridge/src/app.controller.ts:
//
//	GET  /  → GetHello
var ROUTES = router.ForRoutes(
	router.Get("/", []any{AppController{}, "GetHello"}),
)
