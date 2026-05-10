package app

import "github.com/awesome-goose/goose/modules/router"

var ROUTES = router.ForRoutes(
	router.Get("/", []any{AppController{}, "Health"}),
	router.Post("/", []any{AppController{}, "Push"}),
	router.Get("/pop", []any{AppController{}, "Pop"}),
	router.Patch("/log", []any{AppController{}, "Log"}),
)
