package projecttype

import "github.com/awesome-goose/goose/modules/router"

var ROUTES = router.ForRoutes(
	router.Resource("project-types", ProjectTypeController{}).All(),
)
