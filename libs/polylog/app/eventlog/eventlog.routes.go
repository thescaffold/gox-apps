package eventlog

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors TS @Controller('event-log') — singular kebab-case segment.
// /event-log/dashboard precedes the Resource block so it isn't shadowed by /:id.
var ROUTES = router.ForRoutes(
	router.Get("/event-log/dashboard", []any{EventLogController{}, "Dashboard"}),
	router.Resource("event-log", EventLogController{}).All(),
)
