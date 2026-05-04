package provider

import "github.com/awesome-goose/goose/modules/router"

var ROUTES = router.ForRoutes(
	router.Resource("providers", ProviderController{}).All(),
)
