package template

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors TS @Controller('template') — singular path segment.
var ROUTES = router.ForRoutes(
	router.Resource("template", TemplateController{}).All(),
)
