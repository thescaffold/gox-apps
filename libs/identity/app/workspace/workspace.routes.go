package workspace

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors TS @Controller('workspace') — singular path segment.
var ROUTES = router.ForRoutes(
	router.Resource("workspace", WorkspaceController{}).All(),
)
