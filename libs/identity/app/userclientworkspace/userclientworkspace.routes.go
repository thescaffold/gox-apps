package userclientworkspace

import "github.com/awesome-goose/goose/modules/router"

var ROUTES = router.ForRoutes(
	router.Resource("user-client-workspaces", UserClientWorkspaceController{}).All(),
)
