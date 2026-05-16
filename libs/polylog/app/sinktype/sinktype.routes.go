package sinktype

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors TS @Controller('sink-type') — singular kebab-case segment.
var ROUTES = router.ForRoutes(
	router.Resource("sink-type", SinkTypeController{}).All(),
)
