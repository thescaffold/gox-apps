package app

import (
	"github.com/awesome-goose/goose/modules/router"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
)

// ROUTES mirrors ntx-apps/libs/cache/src/app.controller.ts. TS exposes neither
// a health endpoint nor a /flush endpoint — both have been removed.
//
// Guards (PLAN U26): the cache holds codes and session data for every app, so its
// routes are for a platform admin only.
var ROUTES = router.ForRoutes(
	router.Get("/get", []any{AppController{}, "Get"}, coreauth.Admin()...),
	router.Post("/set", []any{AppController{}, "Set"}, coreauth.Admin()...),
	router.Post("/setnx", []any{AppController{}, "SetNx"}, coreauth.Admin()...),
	router.Post("/getset", []any{AppController{}, "GetSet"}, coreauth.Admin()...),
	router.Delete("/del", []any{AppController{}, "Del"}, coreauth.Admin()...),
	router.Get("/ttl", []any{AppController{}, "TTL"}, coreauth.Admin()...),
	router.Post("/incr", []any{AppController{}, "Incr"}, coreauth.Admin()...),
)
