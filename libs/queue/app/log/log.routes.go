package log

import "github.com/awesome-goose/goose/modules/router"

var ROUTES = router.ForRoutes(
	router.Resource("logs", LogController{}).All(),
)
