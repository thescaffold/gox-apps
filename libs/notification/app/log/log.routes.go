package log

import (
	"github.com/awesome-goose/goose/modules/router"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
)

// ROUTES mirrors TS @Controller('log') — singular path segment.
var ROUTES = router.ForRoutes(
	coreauth.AdminWrites(router.Resource("log", LogController{}).All()),
)
