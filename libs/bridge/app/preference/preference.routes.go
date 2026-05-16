package preference

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors TS @Controller('preference') — singular path segment.
var ROUTES = router.ForRoutes(
	router.Resource("preference", PreferenceController{}).All(),
)
