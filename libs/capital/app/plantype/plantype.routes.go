package plantype

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors TS @Controller('plan-type') — singular kebab-case segment.
var ROUTES = router.ForRoutes(
	router.Resource("plan-type", PlanTypeController{}).All(),
)
