package sink

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors TS @Controller('sink') — singular path segment.
var ROUTES = router.ForRoutes(
	router.Resource("sink", SinkController{}).All(),
)
