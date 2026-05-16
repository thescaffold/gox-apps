package permissiontype

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors TS @Controller('permission-type') — singular kebab-case segment.
var ROUTES = router.ForRoutes(
	router.Resource("permission-type", PermissionTypeController{}).All(),
)
