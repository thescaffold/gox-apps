package wallet

import "github.com/awesome-goose/goose/modules/router"

var ROUTES = router.ForRoutes(
	router.Post("/wallet/init", []any{WalletController{}, "Init"}),
	router.Get("/wallet/detail", []any{WalletController{}, "Detail"}),
	router.Get("/wallet/balance", []any{WalletController{}, "Balance"}),
	router.Post("/wallet/upgrade", []any{WalletController{}, "Upgrade"}),
	router.Get("/wallet/transactions", []any{WalletController{}, "Transactions"}),
)
