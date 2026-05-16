package sourcetype

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors TS @Controller('source-type') — singular kebab-case segment.
var ROUTES = router.ForRoutes(
	router.Resource("source-type", SourceTypeController{}).All(),
)
