package voucher

import (
	"github.com/awesome-goose/goose/modules/router"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
)

// ROUTES mirrors TS @Controller('voucher') — singular path segment.
// /voucher/list and /voucher/verify/:token precede the Resource catch-all
// so they aren't shadowed by /:id.
var ROUTES = router.ForRoutes(
	router.Get("/voucher/list", []any{VoucherController{}, "ListAvailable"}, coreauth.Caller()...),
	router.Post("/voucher/verify/:token", []any{VoucherController{}, "Verify"}, coreauth.Caller()...),
	coreauth.AdminWrites(router.Resource("voucher", VoucherController{}).All()),
)
