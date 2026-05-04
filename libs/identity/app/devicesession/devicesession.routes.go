package devicesession

import "github.com/awesome-goose/goose/modules/router"

var ROUTES = router.ForRoutes(
	router.Resource("device-sessions", DeviceSessionController{}).All(),
)
