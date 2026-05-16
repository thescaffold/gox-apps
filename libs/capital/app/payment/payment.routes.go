package payment

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors TS @Controller('payment') — singular path segment.
var ROUTES = router.ForRoutes(
	router.Resource("payment", PaymentController{}).All(),
)
