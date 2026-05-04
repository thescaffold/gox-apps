package app

import "github.com/awesome-goose/goose/modules/router"

var ROUTES = router.ForRoutes(
	router.Get("/",                   []any{AppController{}, "Health"}),
	router.Get("/currency/:currency", []any{AppController{}, "GetCurrency"}),
	router.Get("/location",           []any{AppController{}, "GetLocation"}),
)
