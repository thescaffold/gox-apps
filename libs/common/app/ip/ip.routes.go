package ip

import (
	"github.com/awesome-goose/goose/modules/router"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
)

// ROUTES mirrors TS @Controller('ip') — singular path segment.
var ROUTES = router.ForRoutes(
	coreauth.SecureAdmin(router.Resource("ip", IpController{}).All()),
)
