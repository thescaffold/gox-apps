package licensetype

import "github.com/awesome-goose/goose/modules/router"

var ROUTES = router.ForRoutes(
	router.Resource("licensetypes", LicenseTypeController{}).All(),
)
