package app

import (
	"github.com/awesome-goose/goose/modules/router"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
)

// ROUTES mirrors ntx-apps/libs/figs/src/app.controller.ts route paths:
//
//	GET    /          → GetHello
//	POST   /:name     → SaveFile
//
// Guards (PLAN U26): POST /:name writes a platform-wide named file, so it needs a
// platform admin; GET / is public.
var ROUTES = router.ForRoutes(
	router.Get("/", []any{AppController{}, "GetHello"}),
	router.Post("/:name", []any{AppController{}, "SaveFile"}, coreauth.Admin()...),
)
