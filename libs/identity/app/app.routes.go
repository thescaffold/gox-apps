package app

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors ntx-apps/libs/identity/src/app.controller.ts:
//
//	GET  /  → GetHello
//
// The TS controller declares many more endpoints (initiate/verify/secret/
// login/logout/reset-secret/sessions/now/verify-token/etc) that depend on
// unported AuthService/ProviderService.authorize/etc. Those land in
// gox-apps/libs/identity/pkg/auth.controller.go and pkg/oauth.controller.go.
var ROUTES = router.ForRoutes(
	router.Get("/", []any{AppController{}, "GetHello"}),
)
