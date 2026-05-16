package webhook

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors TS @Controller('webhook') — singular path segment.
var ROUTES = router.ForRoutes(
	router.Resource("webhook", WebhookController{}).All(),
)
