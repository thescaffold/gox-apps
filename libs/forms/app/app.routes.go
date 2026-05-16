package app

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors ntx-apps/libs/forms/src/app.controller.ts: GET / is hello,
// GET /one/:id returns the toForm() view, POST /one/:id writes the logs.
var ROUTES = router.ForRoutes(
	router.Get("/", []any{AppController{}, "GetHello"}),
	router.Get("/one/:id", []any{AppController{}, "GetForm"}),
	router.Post("/one/:id", []any{AppController{}, "SaveForm"}),
)
