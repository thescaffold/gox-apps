package permission

import "github.com/awesome-goose/goose/modules/router"

var ROUTES = router.ForRoutes(
	router.Resource("permissions", PermissionController{}).All(),
)
