package permissiontype

import "github.com/awesome-goose/goose/modules/router"

var ROUTES = router.ForRoutes(
	router.Resource("permission-types", PermissionTypeController{}).All(),
)
