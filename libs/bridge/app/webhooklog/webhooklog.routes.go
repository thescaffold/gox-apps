package webhooklog

import "github.com/awesome-goose/goose/modules/router"

var ROUTES = router.ForRoutes(router.Resource("webhooklogs", WebhookLogController{}).All())
