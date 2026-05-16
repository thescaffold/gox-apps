package devicelog

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors TS @Controller('device-log') — singular kebab-case segment.
var ROUTES = router.ForRoutes(
	router.Resource("device-log", DeviceLogController{}).All(),
)
