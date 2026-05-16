package app

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors ntx-apps/libs/fuss/src/app.controller.ts: hello, recent
// search-history, full-text search, lookup by entity tuple.
var ROUTES = router.ForRoutes(
	router.Get("/", []any{AppController{}, "GetHello"}),
	router.Get("/recent", []any{AppController{}, "Recent"}),
	router.Get("/search", []any{AppController{}, "Search"}),
	router.Get("/lookup", []any{AppController{}, "Lookup"}),
)
