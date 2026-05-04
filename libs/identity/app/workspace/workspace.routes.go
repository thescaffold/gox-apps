package workspace

import "github.com/awesome-goose/goose/modules/router"

var ROUTES = router.ForRoutes(
	router.Resource("workspaces", WorkspaceController{}).All(),
)
