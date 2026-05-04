package formtype

import "github.com/awesome-goose/goose/modules/router"

var ROUTES = router.ForRoutes(
	router.Resource("formtypes", FormTypeController{}).All(),
)
