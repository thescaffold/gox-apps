package environment

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors TS @Controller('environment') — singular path segment.
var ROUTES = router.ForRoutes(
	router.Resource("environment", EnvironmentController{}).All(),
)
