package channel

import (
	"github.com/awesome-goose/goose/modules/router"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
)

// ROUTES mirrors TS @Controller('channel') — singular path segment.
var ROUTES = router.ForRoutes(
	coreauth.Secure(router.Resource("channel", ChannelController{}).All()),
)
