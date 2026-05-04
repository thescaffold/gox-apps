package formfield

import "github.com/awesome-goose/goose/modules/router"

var ROUTES = router.ForRoutes(
	router.Resource("formfields", FormFieldController{}).All(),
)
