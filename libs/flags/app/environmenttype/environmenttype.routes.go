package environmenttype

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors TS @Controller('environment-type') — kebab-case singular path.
var ROUTES = router.ForRoutes(
	router.Resource("environment-type", EnvironmentTypeController{}).All(),
)
