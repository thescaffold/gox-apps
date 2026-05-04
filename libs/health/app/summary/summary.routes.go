package summary

import "github.com/awesome-goose/goose/modules/router"

var ROUTES = router.ForRoutes(
	router.Resource("summary", SummaryController{}).All(),
)
