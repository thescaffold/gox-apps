package voucher

import "github.com/awesome-goose/goose/modules/router"

var ROUTES = router.ForRoutes(
	router.Resource("vouchers", VoucherController{}).All(),
)
