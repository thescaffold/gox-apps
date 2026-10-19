package environmenttype

import (
	"github.com/awesome-goose/goose/modules/router"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
)

// ROUTES mirrors TS @Controller('environment-type') — kebab-case singular path.
var ROUTES = router.ForRoutes(
	coreauth.Secure(router.Resource("environment-type", EnvironmentTypeController{}).All()),
)
