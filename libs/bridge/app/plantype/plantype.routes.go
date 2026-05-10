package plantype

import "github.com/awesome-goose/goose/modules/router"

var ROUTES = router.ForRoutes(router.Resource("plantypes", PlanTypeController{}).All())
