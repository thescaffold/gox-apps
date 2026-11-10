package app

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors ntx-apps/libs/forms/src/app.controller.ts: GET / is hello,
// GET /one/:id returns the toForm() view, POST /one/:id writes the logs.
var ROUTES = router.ForRoutes(
	router.Get("/", []any{AppController{}, "GetHello"}),
	// Public on purpose: a form is rendered and submitted by whoever it is shared with. They
	// answer the form's definition and store a submission; submissions are read back only
	// through formlogs, by the form's workspace (U26).
	router.Get("/one/:id", []any{AppController{}, "GetForm"}),
	router.Post("/one/:id", []any{AppController{}, "SaveForm"}),
)
