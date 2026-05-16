package log

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors TS @Controller('log') — singular path segment.
var ROUTES = router.ForRoutes(
	router.Resource("log", LogController{}).All(),
)
