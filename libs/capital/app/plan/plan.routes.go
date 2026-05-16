package plan

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors TS @Controller('plan') — singular path segment.
var ROUTES = router.ForRoutes(
	router.Resource("plan", PlanController{}).All(),
)
