package app

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors ntx-apps/libs/polylog/src/app.controller.ts route paths:
//
//	GET    /                   → GetHello
//	POST   /ingest/batch       → IngestBatch
//	POST   /ingest             → Ingest
//	POST   /ingest/:id         → IngestSourceAlt
//	POST   /ingest/source/:id  → IngestSource
//	POST   /ingest/channel/:id → IngestChannel
//	POST   /config             → SetConfig
//	GET    /config             → GetConfig
//	PATCH  /config/:id         → UpdateConfig
//	POST   /root-source        → UpsertRootSource
//	POST   /root-sink          → UpsertRootSink
var ROUTES = router.ForRoutes(
	router.Get("/", []any{AppController{}, "GetHello"}),
	router.Post("/ingest/batch", []any{AppController{}, "IngestBatch"}),
	router.Post("/ingest", []any{AppController{}, "Ingest"}),
	router.Post("/ingest/:id", []any{AppController{}, "IngestSourceAlt"}),
	router.Post("/ingest/source/:id", []any{AppController{}, "IngestSource"}),
	router.Post("/ingest/channel/:id", []any{AppController{}, "IngestChannel"}),
	router.Post("/config", []any{AppController{}, "SetConfig"}),
	router.Get("/config", []any{AppController{}, "GetConfig"}),
	router.Patch("/config/:id", []any{AppController{}, "UpdateConfig"}),
	router.Post("/root-source", []any{AppController{}, "UpsertRootSource"}),
	router.Post("/root-sink", []any{AppController{}, "UpsertRootSink"}),
)
