package licensetype

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors TS @Controller('license-type') — singular kebab-case segment.
var ROUTES = router.ForRoutes(
	router.Resource("license-type", LicenseTypeController{}).All(),
)
