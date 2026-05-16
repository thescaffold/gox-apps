package voucher

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors TS @Controller('voucher') — singular path segment.
// /voucher/list and /voucher/verify/:token precede the Resource catch-all
// so they aren't shadowed by /:id.
var ROUTES = router.ForRoutes(
	router.Get("/voucher/list", []any{VoucherController{}, "ListAvailable"}),
	router.Post("/voucher/verify/:token", []any{VoucherController{}, "Verify"}),
	router.Resource("voucher", VoucherController{}).All(),
)
