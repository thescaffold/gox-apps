package tokenlog

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors TS @Controller('token-log') — singular path segment.
var ROUTES = router.ForRoutes(
	router.Resource("token-log", TokenLogController{}).All(),
)
