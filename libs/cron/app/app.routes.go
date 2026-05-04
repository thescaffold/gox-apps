package app

import "github.com/awesome-goose/goose/modules/router"

var ROUTES = router.ForRoutes(
	router.Get("/",       []any{AppController{}, "Health"}),
	router.Post("/",      []any{AppController{}, "Register"}),
	router.Get("/select", []any{AppController{}, "Select"}),
	router.Patch("/log",  []any{AppController{}, "Log"}),
)
