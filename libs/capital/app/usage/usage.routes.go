package usage

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors TS @Controller('usage') — singular path segment.
var ROUTES = router.ForRoutes(
	router.Resource("usage", UsageController{}).All(),
)
