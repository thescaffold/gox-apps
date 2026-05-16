package ip

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors TS @Controller('ip') — singular path segment.
var ROUTES = router.ForRoutes(
	router.Resource("ip", IpController{}).All(),
)
