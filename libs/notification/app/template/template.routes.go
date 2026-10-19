package template

import (
	"github.com/awesome-goose/goose/modules/router"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
)

// ROUTES mirrors TS @Controller('template') — singular path segment.
var ROUTES = router.ForRoutes(
	coreauth.Secure(router.Resource("template", TemplateController{}).All()),
)
