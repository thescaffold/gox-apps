package userclientworkspace

import (
	"github.com/awesome-goose/goose/modules/router"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
)

// ROUTES mirrors TS @Controller('user-client-workspace') — singular kebab-case segment.
var ROUTES = router.ForRoutes(
	coreauth.Secure(router.Resource("user-client-workspace", UserClientWorkspaceController{}).All()),
)
