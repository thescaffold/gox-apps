package app

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors ntx-apps/libs/common/src/app.controller.ts route paths:
//
//	GET  /                      → GetHello
//	GET  /currency/:currency    → GetCurrency
//	GET  /location              → GetLocation
var ROUTES = router.ForRoutes(
	router.Get("/", []any{AppController{}, "GetHello"}),
	router.Get("/currency/:currency", []any{AppController{}, "GetCurrency"}),
	router.Get("/location", []any{AppController{}, "GetLocation"}),
)
