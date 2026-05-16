package transaction

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors TS @Controller('transaction') — singular path segment.
var ROUTES = router.ForRoutes(
	router.Resource("transaction", TransactionController{}).All(),
)
