package attribute

import "github.com/awesome-goose/goose/modules/router"

var ROUTES = router.ForRoutes(
	router.Resource("attributes", AttributeController{}).All(),
)
