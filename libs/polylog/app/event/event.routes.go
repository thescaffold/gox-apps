package event

import (
	"github.com/awesome-goose/goose/modules/router"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
)

// ROUTES mirrors TS @Controller('event') — singular path segment.
// /event/dashboard precedes the Resource block so it isn't shadowed by /event/:id.
var ROUTES = router.ForRoutes(
	router.Get("/event/dashboard", []any{EventController{}, "Dashboard"}),
	coreauth.Secure(router.Resource("event", EventController{}).All()),
)
