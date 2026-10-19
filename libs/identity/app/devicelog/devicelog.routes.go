package devicelog

import (
	"github.com/awesome-goose/goose/modules/router"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
)

// ROUTES mirrors TS @Controller('device-log') — singular kebab-case segment.
var ROUTES = router.ForRoutes(
	coreauth.Secure(router.Resource("device-log", DeviceLogController{}).All()),
)
