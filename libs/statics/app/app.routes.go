package app

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors ntx-apps/libs/statics/src/app.controller.ts route paths:
//
//	GET  /                 → GetHello
//	GET  /filter/:key      → FilterByKey
//	GET  /code/:code       → FindByCode
//	GET  /value/:value     → FindByValue
var ROUTES = router.ForRoutes(
	router.Get("/", []any{AppController{}, "GetHello"}),
	router.Get("/filter/:key", []any{AppController{}, "FilterByKey"}),
	router.Get("/code/:code", []any{AppController{}, "FindByCode"}),
	router.Get("/value/:value", []any{AppController{}, "FindByValue"}),
)
