package roletype

import (
	"github.com/awesome-goose/goose/modules/router"
	"github.com/thescaffold/gox-apps/libs/identity/pkg/rbac"
)

// ROUTES mirrors TS @Controller('role-type') — singular kebab-case segment.
var ROUTES = router.ForRoutes(
	rbac.Secure(router.Resource("role-type", RoleTypeController{}).All()),
)
