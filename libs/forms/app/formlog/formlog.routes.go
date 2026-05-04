package formlog

import "github.com/awesome-goose/goose/modules/router"

var ROUTES = router.ForRoutes(
	router.Resource("formlogs", FormLogController{}).All(),
)
