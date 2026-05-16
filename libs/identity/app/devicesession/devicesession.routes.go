package devicesession

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors TS @Controller('device-session') — singular kebab-case segment.
var ROUTES = router.ForRoutes(
	router.Resource("device-session", DeviceSessionController{}).All(),
)
