package plantype

import (
	"github.com/awesome-goose/goose/modules/router"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
)

// ROUTES mirrors TS @Controller('plan-type') — singular kebab-case segment.
var ROUTES = router.ForRoutes(
	coreauth.Secure(router.Resource("plan-type", PlanTypeController{}).All()),
)
