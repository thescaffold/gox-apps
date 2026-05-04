package tagtype

import "github.com/awesome-goose/goose/modules/router"

var ROUTES = router.ForRoutes(
	router.Resource("tag-types", TagTypeController{}).All(),
)
