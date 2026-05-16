package app

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors ntx-apps/libs/capital/src/app.controller.ts:
//
//	GET    /                              GetHello
//	GET    /:type/reference/:reference    Reference
//	POST   /default-provider              GetDefaultProvider (deprecated)
//	GET    /providers                     GetProviders
//	POST   /init                          Init
//	POST   /verify/:type                  Verify
//	POST   /:type/charge/:providerId      Charge
//	POST   /usage/start                   UsageStart
//	POST   /usage/update                  UsageUpdate
//	POST   /usage/stop                    UsageStop
//	GET    /debt                          Debt
//	GET    /accrual                       Accrual
//	POST   /pay                           Pay
//	GET    /status                        Status
//	GET    /user-plan                     UserPlan
//	POST   /user-plan/compute-upgrade     ComputeUpgrade
//	POST   /plan/change                   PlanChange
//	POST   /plan/subscribe                PlanSubscribe
//
// Specific paths precede `/:param` catch-alls so e.g. /providers isn't
// shadowed by /:type/reference/:reference.
var ROUTES = router.ForRoutes(
	router.Get("/", []any{AppController{}, "GetHello"}),
	router.Get("/providers", []any{AppController{}, "GetProviders"}),
	router.Get("/debt", []any{AppController{}, "Debt"}),
	router.Get("/accrual", []any{AppController{}, "Accrual"}),
	router.Get("/status", []any{AppController{}, "Status"}),
	router.Get("/user-plan", []any{AppController{}, "UserPlan"}),
	router.Post("/default-provider", []any{AppController{}, "GetDefaultProvider"}),
	router.Post("/init", []any{AppController{}, "Init"}),
	router.Post("/pay", []any{AppController{}, "Pay"}),
	router.Post("/usage/start", []any{AppController{}, "UsageStart"}),
	router.Post("/usage/update", []any{AppController{}, "UsageUpdate"}),
	router.Post("/usage/stop", []any{AppController{}, "UsageStop"}),
	router.Post("/user-plan/compute-upgrade", []any{AppController{}, "ComputeUpgrade"}),
	router.Post("/plan/change", []any{AppController{}, "PlanChange"}),
	router.Post("/plan/subscribe", []any{AppController{}, "PlanSubscribe"}),
	router.Post("/verify/:type", []any{AppController{}, "Verify"}),
	router.Get("/:type/reference/:reference", []any{AppController{}, "Reference"}),
	router.Post("/:type/charge/:providerId", []any{AppController{}, "Charge"}),
)
