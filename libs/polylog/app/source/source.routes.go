package source

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors TS @Controller('source') — singular path segment.
var ROUTES = router.ForRoutes(
	router.Resource("source", SourceController{}).All(),
)
