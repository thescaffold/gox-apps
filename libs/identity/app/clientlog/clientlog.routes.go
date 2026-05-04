package clientlog

import "github.com/awesome-goose/goose/modules/router"

var ROUTES = router.ForRoutes(
	router.Resource("client-logs", ClientLogController{}).All(),
)
