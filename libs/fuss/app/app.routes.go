package app

import (
	"github.com/awesome-goose/goose/modules/router"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
)

// ROUTES mirrors ntx-apps/libs/fuss/src/app.controller.ts: hello, recent
// search-history, full-text search, lookup by entity tuple.
var ROUTES = router.ForRoutes(
	router.Get("/", []any{AppController{}, "GetHello"}),
	router.Get("/recent", []any{AppController{}, "Recent"}, coreauth.Caller()...),
	router.Get("/search", []any{AppController{}, "Search"}, coreauth.Caller()...),
	router.Get("/lookup", []any{AppController{}, "Lookup"}, coreauth.Caller()...),
)
