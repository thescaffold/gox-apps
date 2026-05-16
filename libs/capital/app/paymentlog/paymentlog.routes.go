package paymentlog

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors TS @Controller('payment-log') — singular kebab-case segment.
var ROUTES = router.ForRoutes(
	router.Resource("payment-log", PaymentLogController{}).All(),
)
