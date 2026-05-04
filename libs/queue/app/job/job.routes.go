package job

import "github.com/awesome-goose/goose/modules/router"

var ROUTES = router.ForRoutes(
	router.Resource("jobs", JobController{}).All(),
)
