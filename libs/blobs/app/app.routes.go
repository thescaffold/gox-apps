package app

import "github.com/awesome-goose/goose/modules/router"

// ROUTES registers AppController endpoints.
// Mirrors ntx-apps/libs/blobs/src/app.controller.ts route paths:
//
//	GET    /                  → GetHello
//	POST   /upload/init       → Init
//	POST   /upload/batch      → Batch
//	POST   /upload/verify     → Verify
//	POST   /upload            → Upload (single-shot)
//	GET    /download/:id      → Download
var ROUTES = router.ForRoutes(
	router.Get("/", []any{AppController{}, "GetHello"}),
	router.Post("/upload/init", []any{AppController{}, "Init"}),
	router.Post("/upload/batch", []any{AppController{}, "Batch"}),
	router.Post("/upload/verify", []any{AppController{}, "Verify"}),
	router.Post("/upload", []any{AppController{}, "Upload"}),
	router.Get("/download/:id", []any{AppController{}, "Download"}),
)
