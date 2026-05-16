package account

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors TS @Controller('account') — singular path segment.
var ROUTES = router.ForRoutes(
	router.Resource("account", AccountController{}).All(),
)
