package app

import "github.com/awesome-goose/goose/modules/router"

var ROUTES = router.ForRoutes(
	router.Get("/",      []any{AppController{}, "Health"}),
	router.Get("/get",   []any{AppController{}, "Get"}),
	router.Post("/set",  []any{AppController{}, "Set"}),
	router.Delete("/del",   []any{AppController{}, "Del"}),
	router.Delete("/flush", []any{AppController{}, "Flush"}),
)
