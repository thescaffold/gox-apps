package flag

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors TS @Controller('flag') — singular path segment.
var ROUTES = router.ForRoutes(
	router.Resource("flag", FlagController{}).All(),
)
