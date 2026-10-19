package app

import (
	"github.com/awesome-goose/goose/modules/router"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
)

// ROUTES mirrors ntx-apps/libs/cache/src/app.controller.ts. TS exposes neither
// a health endpoint nor a /flush endpoint — both have been removed.
var ROUTES = router.ForRoutes(
	router.Get("/get", []any{AppController{}, "Get"}, coreauth.Session()...),
	router.Post("/set", []any{AppController{}, "Set"}, coreauth.Session()...),
	router.Post("/setnx", []any{AppController{}, "SetNx"}, coreauth.Session()...),
	router.Post("/getset", []any{AppController{}, "GetSet"}, coreauth.Session()...),
	router.Delete("/del", []any{AppController{}, "Del"}, coreauth.Session()...),
	router.Get("/ttl", []any{AppController{}, "TTL"}, coreauth.Session()...),
	router.Post("/incr", []any{AppController{}, "Incr"}, coreauth.Session()...),
)
