package app

import "github.com/awesome-goose/goose/modules/router"

var ROUTES = router.ForRoutes(
	router.Get("/",             []any{AppController{}, "Health"}),
	router.Get("/filter/:key",  []any{AppController{}, "FilterByKey"}),
	router.Get("/code/:code",   []any{AppController{}, "FindByCode"}),
	router.Get("/value/:value", []any{AppController{}, "FindByValue"}),
)
