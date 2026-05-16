package user

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors TS @Controller('user') — singular path segment.
var ROUTES = router.ForRoutes(
	router.Resource("user", UserController{}).All(),
)
