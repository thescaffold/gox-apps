package client

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors TS @Controller('client') — singular path segment.
var ROUTES = router.ForRoutes(
	router.Resource("client", ClientController{}).All(),
)
