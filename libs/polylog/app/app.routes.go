package app

import (
	"github.com/awesome-goose/goose/modules/router"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
)

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
//
// Guards (PLAN U26): the ingest routes are public. The three webhook routes
// check the source's signature rule; /ingest and /ingest/batch are the intake
// browser SDKs post to (jsx-polylog batches every few seconds, for guests too,
// with a credential shipped in the page), so they stay open but refuse any item
// for a source that has a signature rule (AppService.checkIntake). config, root-source
// (creates root-owned sources and reads any by id) and root-sink (rewrites
// where events are delivered) are platform-admin only.
var ROUTES = router.ForRoutes(
	router.Get("/", []any{AppController{}, "GetHello"}),
	router.Post("/ingest/batch", []any{AppController{}, "IngestBatch"}),
	router.Post("/ingest", []any{AppController{}, "Ingest"}),
	router.Post("/ingest/:id", []any{AppController{}, "IngestSourceAlt"}),
	router.Post("/ingest/source/:id", []any{AppController{}, "IngestSource"}),
	router.Post("/ingest/channel/:id", []any{AppController{}, "IngestChannel"}),
	router.Post("/config", []any{AppController{}, "SetConfig"}, coreauth.Admin()...),
	router.Get("/config", []any{AppController{}, "GetConfig"}, coreauth.Admin()...),
	router.Patch("/config/:id", []any{AppController{}, "UpdateConfig"}, coreauth.Admin()...),
	router.Post("/root-source", []any{AppController{}, "UpsertRootSource"}, coreauth.Admin()...),
	router.Post("/root-sink", []any{AppController{}, "UpsertRootSink"}, coreauth.Admin()...),
)
