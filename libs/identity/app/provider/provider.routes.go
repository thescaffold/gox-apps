package provider

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors TS @Controller('provider') — singular path segment.
// /provider/providers + /provider/:provider/provider precede the Resource
// catch-all so they aren't shadowed by /provider/:id.
// POST /provider/:provider/authorize wires the full OAuth login flow.
var ROUTES = router.ForRoutes(
	router.Get("/provider/providers", []any{ProviderController{}, "Providers"}),
	router.Get("/provider/:provider/provider", []any{ProviderController{}, "Provider"}),
	router.Post("/provider/:provider/authorize", []any{AuthorizeController{}, "Authorize"}),
	router.Resource("provider", ProviderController{}).All(),
)
