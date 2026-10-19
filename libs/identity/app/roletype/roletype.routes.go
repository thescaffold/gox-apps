package roletype

import (
	"github.com/awesome-goose/goose/modules/router"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
)

// ROUTES mirrors TS @Controller('role-type') — singular kebab-case segment.
var ROUTES = router.ForRoutes(
	coreauth.Secure(router.Resource("role-type", RoleTypeController{}).All()),
)
