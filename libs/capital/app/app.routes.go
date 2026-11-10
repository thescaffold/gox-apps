package app

import (
	"github.com/awesome-goose/goose/modules/router"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
)

// ROUTES mirrors ntx-apps/libs/capital/src/app.controller.ts:
//
//	GET    /                              GetHello            public
//	GET    /:type/reference/:reference    Reference           caller
//	POST   /default-provider              GetDefaultProvider  caller (deprecated)
//	GET    /providers                     GetProviders        public (publishable keys)
//	POST   /init                          Init                caller
//	POST   /verify/:type                  Verify              caller
//	POST   /:type/charge/:providerId      Charge              caller
//	POST   /usage/start                   UsageStart          internal
//	POST   /usage/update                  UsageUpdate         internal
//	POST   /usage/stop                    UsageStop           internal
//	GET    /debt                          Debt                caller
//	GET    /accrual                       Accrual             caller
//	POST   /pay                           Pay                 caller
//	GET    /status                        Status              caller
//	GET    /user-plan                     UserPlan            caller
//	POST   /user-plan/compute-upgrade     ComputeUpgrade      caller
//	POST   /plan/change                   PlanChange          internal
//	POST   /plan/subscribe                PlanSubscribe       internal
//
// "caller" routes act on the verified caller's own plan, payments and cards (coreauth.Caller: the
// user, client and workspace come from the token). The usage and plan hooks name the user in the
// body, so only another service (core/http.Client.Internal, signed with HMAC_KEY) or a platform
// admin may call them (coreauth.Internal). Before PLAN U26 every route here answered anonymous
// callers.
//
// Specific paths precede `/:param` catch-alls so e.g. /providers isn't
// shadowed by /:type/reference/:reference.
var ROUTES = router.ForRoutes(
	router.Get("/", []any{AppController{}, "GetHello"}),
	router.Get("/providers", []any{AppController{}, "GetProviders"}),
	router.Get("/debt", []any{AppController{}, "Debt"}, coreauth.Caller()...),
	router.Get("/accrual", []any{AppController{}, "Accrual"}, coreauth.Caller()...),
	router.Get("/status", []any{AppController{}, "Status"}, coreauth.Caller()...),
	router.Get("/user-plan", []any{AppController{}, "UserPlan"}, coreauth.Caller()...),
	router.Post("/default-provider", []any{AppController{}, "GetDefaultProvider"}, coreauth.Caller()...),
	router.Post("/init", []any{AppController{}, "Init"}, coreauth.Caller()...),
	router.Post("/pay", []any{AppController{}, "Pay"}, coreauth.Caller()...),
	router.Post("/usage/start", []any{AppController{}, "UsageStart"}, coreauth.Internal()...),
	router.Post("/usage/update", []any{AppController{}, "UsageUpdate"}, coreauth.Internal()...),
	router.Post("/usage/stop", []any{AppController{}, "UsageStop"}, coreauth.Internal()...),
	router.Post("/user-plan/compute-upgrade", []any{AppController{}, "ComputeUpgrade"}, coreauth.Caller()...),
	router.Post("/plan/change", []any{AppController{}, "PlanChange"}, coreauth.Internal()...),
	router.Post("/plan/subscribe", []any{AppController{}, "PlanSubscribe"}, coreauth.Internal()...),
	router.Post("/verify/:type", []any{AppController{}, "Verify"}, coreauth.Caller()...),
	router.Get("/:type/reference/:reference", []any{AppController{}, "Reference"}, coreauth.Caller()...),
	router.Post("/:type/charge/:providerId", []any{AppController{}, "Charge"}, coreauth.Caller()...),
)
