package environment

import (
	"github.com/awesome-goose/goose/modules/router"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
)

// ROUTES mirrors TS @Controller('environment') — singular path segment.
var ROUTES = router.ForRoutes(
	coreauth.Secure(router.Resource("environment", EnvironmentController{}).All()),
)
