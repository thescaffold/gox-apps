package ratelog

import "github.com/awesome-goose/goose/modules/router"

var ROUTES = router.ForRoutes(
	router.Resource("rate-logs", RateLogController{}).All(),
)
