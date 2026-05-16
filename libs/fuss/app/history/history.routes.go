package history

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors TS @Controller('history') — singular path segment.
var ROUTES = router.ForRoutes(
	router.Resource("history", HistoryController{}).All(),
)
