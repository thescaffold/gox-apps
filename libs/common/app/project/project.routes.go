package project

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors TS @Controller('project') — singular path segment.
var ROUTES = router.ForRoutes(
	router.Resource("project", ProjectController{}).All(),
)
