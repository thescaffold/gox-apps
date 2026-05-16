package rule

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors TS @Controller('rule') — singular path segment.
var ROUTES = router.ForRoutes(
	router.Resource("rule", RuleController{}).All(),
)
