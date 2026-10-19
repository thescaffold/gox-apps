package page

import (
	"github.com/awesome-goose/goose/modules/router"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
)

var ROUTES = router.ForRoutes(
	coreauth.Secure(router.Resource("page", PageController{}).All()),
)
