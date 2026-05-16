package roletype

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors TS @Controller('role-type') — singular kebab-case segment.
var ROUTES = router.ForRoutes(
	router.Resource("role-type", RoleTypeController{}).All(),
)
