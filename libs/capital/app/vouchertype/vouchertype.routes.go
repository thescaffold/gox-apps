package vouchertype

import (
	"github.com/awesome-goose/goose/modules/router"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
)

// ROUTES mirrors TS @Controller('voucher-type') — singular kebab-case segment.
var ROUTES = router.ForRoutes(
	coreauth.Secure(router.Resource("voucher-type", VoucherTypeController{}).All()),
)
