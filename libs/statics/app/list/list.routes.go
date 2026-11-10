package list

import (
	"github.com/awesome-goose/goose/modules/router"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
)

var ROUTES = router.ForRoutes(
	coreauth.SecureAdmin(router.Resource("list", ListController{}).All(), "List", "Get"),
)
