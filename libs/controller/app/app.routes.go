package app

import "github.com/awesome-goose/goose/modules/router"

var ROUTES = router.ForRoutes(
	router.Get("/", []any{AppController{}, "Health"}),
	router.Get("/now", []any{AppController{}, "GetRoutes"}),
	router.Get("/route", []any{AppController{}, "GetRoute"}),
)
