package sourcetype

import "github.com/awesome-goose/goose/modules/router"

var ROUTES = router.ForRoutes(
	router.Resource("sourcetypes", SourceTypeController{}).All(),
)
