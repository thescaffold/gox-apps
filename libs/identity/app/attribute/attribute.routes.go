package attribute

import (
	"github.com/awesome-goose/goose/modules/router"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
)

// ROUTES mirrors TS @Controller('attribute') — singular path segment.
var ROUTES = router.ForRoutes(
	coreauth.Secure(router.Resource("attribute", AttributeController{}).All()),
)
