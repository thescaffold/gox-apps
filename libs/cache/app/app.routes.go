package app

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors ntx-apps/libs/cache/src/app.controller.ts. TS exposes neither
// a health endpoint nor a /flush endpoint — both have been removed.
var ROUTES = router.ForRoutes(
	router.Get("/get", []any{AppController{}, "Get"}),
	router.Post("/set", []any{AppController{}, "Set"}),
	router.Post("/setnx", []any{AppController{}, "SetNx"}),
	router.Post("/getset", []any{AppController{}, "GetSet"}),
	router.Delete("/del", []any{AppController{}, "Del"}),
	router.Get("/ttl", []any{AppController{}, "TTL"}),
	router.Post("/incr", []any{AppController{}, "Incr"}),
)
