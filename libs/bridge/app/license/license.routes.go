package license

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors TS @Controller('license') — singular path segment.
var ROUTES = router.ForRoutes(
	router.Resource("license", LicenseController{}).All(),
)
