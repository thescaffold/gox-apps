package clientlog

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors TS @Controller('client-log') — singular kebab-case segment.
var ROUTES = router.ForRoutes(
	router.Resource("client-log", ClientLogController{}).All(),
)
