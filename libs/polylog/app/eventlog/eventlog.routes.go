package eventlog

import (
	"github.com/awesome-goose/goose/modules/router"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
)

// ROUTES mirrors TS @Controller('event-log') — singular kebab-case segment.
// /event-log/dashboard precedes the Resource block so it isn't shadowed by /:id.
var ROUTES = router.ForRoutes(
	router.Get("/event-log/dashboard", []any{EventLogController{}, "Dashboard"}),
	coreauth.Secure(router.Resource("event-log", EventLogController{}).All()),
)
