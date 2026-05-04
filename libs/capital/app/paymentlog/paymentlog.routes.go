package paymentlog
import "github.com/awesome-goose/goose/modules/router"
var ROUTES = router.ForRoutes(router.Resource("paymentlogs", PaymentLogController{}).All())
