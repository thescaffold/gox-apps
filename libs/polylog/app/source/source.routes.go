package source

import (
	"github.com/awesome-goose/goose/modules/router"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
)

// ROUTES mirrors TS @Controller('source') — singular path segment.
var ROUTES = router.ForRoutes(
	coreauth.Secure(router.Resource("source", SourceController{}).All()),
)
