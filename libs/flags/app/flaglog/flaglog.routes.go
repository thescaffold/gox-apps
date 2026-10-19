package flaglog

import (
	"github.com/awesome-goose/goose/modules/router"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
)

// ROUTES mirrors TS @Controller('flag-log') — kebab-case singular path.
var ROUTES = router.ForRoutes(
	coreauth.Secure(router.Resource("flag-log", FlagLogController{}).All()),
)
