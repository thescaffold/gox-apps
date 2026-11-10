package rate

import (
	"github.com/awesome-goose/goose/modules/router"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
)

// ROUTES mirrors TS @Controller('rate') — singular path segment.
var ROUTES = router.ForRoutes(
	coreauth.AdminWrites(router.Resource("rate", RateController{}).All()),
)
