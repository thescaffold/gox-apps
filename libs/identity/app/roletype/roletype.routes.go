package roletype

import "github.com/awesome-goose/goose/modules/router"

var ROUTES = router.ForRoutes(
	router.Resource("role-types", RoleTypeController{}).All(),
)
