package webhooklog

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors TS @Controller('webhook-log') — singular kebab-case segment.
var ROUTES = router.ForRoutes(
	router.Resource("webhook-log", WebhookLogController{}).All(),
)
