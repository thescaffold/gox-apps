package attribute

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors TS @Controller('attribute') — singular path segment.
var ROUTES = router.ForRoutes(
	router.Resource("attribute", AttributeController{}).All(),
)
