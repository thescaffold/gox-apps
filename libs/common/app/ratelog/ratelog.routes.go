package ratelog

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors TS @Controller('rate-log') — singular kebab-case segment.
var ROUTES = router.ForRoutes(
	router.Resource("rate-log", RateLogController{}).All(),
)
