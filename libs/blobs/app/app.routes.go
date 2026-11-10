package app

import (
	"github.com/awesome-goose/goose/modules/router"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
)

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
	router.Post("/upload/init", []any{AppController{}, "Init"}, coreauth.Caller()...),
	router.Post("/upload/batch", []any{AppController{}, "Batch"}, coreauth.Caller()...),
	router.Post("/upload/verify", []any{AppController{}, "Verify"}, coreauth.Caller()...),
	router.Post("/upload", []any{AppController{}, "Upload"}, coreauth.Caller()...),
	router.Get("/download/:id", []any{AppController{}, "Download"}, coreauth.Caller()...),
)
