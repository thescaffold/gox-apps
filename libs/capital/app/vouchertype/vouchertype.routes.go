package vouchertype

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors TS @Controller('voucher-type') — singular kebab-case segment.
var ROUTES = router.ForRoutes(
	router.Resource("voucher-type", VoucherTypeController{}).All(),
)
