package transaction

import "github.com/awesome-goose/goose/modules/router"

var ROUTES = router.ForRoutes(
	router.Resource("transactions", TransactionController{}).All(),
)
