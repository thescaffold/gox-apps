package token

import (
	"github.com/awesome-goose/goose/modules/router"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
)

// ROUTES mirrors TS @Controller('token') — singular path segment.
var ROUTES = router.ForRoutes(
	coreauth.SecureCaller(router.Resource("token", TokenController{}).All()),
)
