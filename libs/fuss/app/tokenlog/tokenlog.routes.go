package tokenlog

import "github.com/awesome-goose/goose/modules/router"

var ROUTES = router.ForRoutes(
	router.Resource("token-logs", TokenLogController{}).All(),
)
