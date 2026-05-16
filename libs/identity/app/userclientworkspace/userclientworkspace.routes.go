package userclientworkspace

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors TS @Controller('user-client-workspace') — singular kebab-case segment.
var ROUTES = router.ForRoutes(
	router.Resource("user-client-workspace", UserClientWorkspaceController{}).All(),
)
