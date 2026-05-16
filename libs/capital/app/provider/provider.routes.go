package provider

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors TS @Controller('provider') — singular path segment.
var ROUTES = router.ForRoutes(
	router.Resource("provider", ProviderController{}).All(),
)
