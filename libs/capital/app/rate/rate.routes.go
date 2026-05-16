package rate

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors TS @Controller('rate') — singular path segment.
var ROUTES = router.ForRoutes(
	router.Resource("rate", RateController{}).All(),
)
