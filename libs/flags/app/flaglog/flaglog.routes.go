package flaglog

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors TS @Controller('flag-log') — kebab-case singular path.
var ROUTES = router.ForRoutes(
	router.Resource("flag-log", FlagLogController{}).All(),
)
