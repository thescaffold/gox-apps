package channel

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors TS @Controller('channel') — singular path segment.
var ROUTES = router.ForRoutes(
	router.Resource("channel", ChannelController{}).All(),
)
