package message

import (
	"github.com/awesome-goose/goose/modules/router"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
)

// ROUTES mirrors TS @Controller('message') — singular path segment.
var ROUTES = router.ForRoutes(
	coreauth.Secure(router.Resource("message", MessageController{}).All()),
)
