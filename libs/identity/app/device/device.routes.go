package device

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors TS @Controller('device') — singular path segment.
var ROUTES = router.ForRoutes(
	router.Resource("device", DeviceController{}).All(),
)
