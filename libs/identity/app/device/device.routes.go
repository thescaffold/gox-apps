package device

import (
	"github.com/awesome-goose/goose/modules/router"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
)

// ROUTES mirrors TS @Controller('device') — singular path segment.
var ROUTES = router.ForRoutes(
	coreauth.Secure(router.Resource("device", DeviceController{}).All()),
)
