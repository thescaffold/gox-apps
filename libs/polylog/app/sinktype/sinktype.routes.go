package sinktype

import (
	"github.com/awesome-goose/goose/modules/router"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
)

// ROUTES mirrors TS @Controller('sink-type') — singular kebab-case segment.
var ROUTES = router.ForRoutes(
	coreauth.SecureAdmin(router.Resource("sink-type", SinkTypeController{}).All()),
)
