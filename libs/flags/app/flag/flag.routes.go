package flag

import (
	"github.com/awesome-goose/goose/modules/router"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
)

// ROUTES mirrors TS @Controller('flag') — singular path segment.
var ROUTES = router.ForRoutes(
	coreauth.Secure(router.Resource("flag", FlagController{}).All()),
)
