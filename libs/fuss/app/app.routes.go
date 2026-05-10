package app

import "github.com/awesome-goose/goose/modules/router"

var ROUTES = router.ForRoutes(
	router.Get("/", []any{AppController{}, "Health"}),
	router.Get("/recent", []any{AppController{}, "Recent"}),
	router.Get("/search", []any{AppController{}, "Search"}),
	router.Get("/lookup", []any{AppController{}, "Lookup"}),
)
