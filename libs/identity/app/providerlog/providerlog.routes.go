package providerlog

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors TS @Controller('provider-log') — singular kebab-case segment.
var ROUTES = router.ForRoutes(
	router.Resource("provider-log", ProviderLogController{}).All(),
)
