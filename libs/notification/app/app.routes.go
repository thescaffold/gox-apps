package app

import (
	"github.com/awesome-goose/goose/modules/router"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
)

// ROUTES mirrors ntx-apps/libs/notification/src/app.controller.ts route paths:
//
//	GET    /                → Health
//	GET    /scope           → FindByScope
//	GET    /priority        → FindByPriority
//	GET    /subscription    → GetSubscription
//	PATCH  /subscription    → UpdateSubscription
//	GET    /capture         → Capture (non-production, token-guarded)
var ROUTES = router.ForRoutes(
	router.Get("/", []any{AppController{}, "GetHello"}),
	router.Get("/scope", []any{AppController{}, "FindByScope"}),
	router.Get("/priority", []any{AppController{}, "FindByPriority"}),
	router.Get("/subscription", []any{AppController{}, "GetSubscription"}, coreauth.Session()...),
	router.Patch("/subscription", []any{AppController{}, "UpdateSubscription"}, coreauth.Session()...),
	router.Get("/capture", []any{AppController{}, "Capture"}),
)
