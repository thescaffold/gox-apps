package app

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors ntx-apps/libs/notification/src/app.controller.ts route paths:
//
//	GET    /                → Health
//	GET    /scope           → FindByScope
//	GET    /priority        → FindByPriority
//	GET    /subscription    → GetSubscription
//	PATCH  /subscription    → UpdateSubscription
var ROUTES = router.ForRoutes(
	router.Get("/", []any{AppController{}, "Health"}),
	router.Get("/scope", []any{AppController{}, "FindByScope"}),
	router.Get("/priority", []any{AppController{}, "FindByPriority"}),
	router.Get("/subscription", []any{AppController{}, "GetSubscription"}),
	router.Patch("/subscription", []any{AppController{}, "UpdateSubscription"}),
)
