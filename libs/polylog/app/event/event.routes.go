package event

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors TS @Controller('event') — singular path segment.
// /event/dashboard precedes the Resource block so it isn't shadowed by /event/:id.
var ROUTES = router.ForRoutes(
	router.Get("/event/dashboard", []any{EventController{}, "Dashboard"}),
	router.Resource("event", EventController{}).All(),
)
