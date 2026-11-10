package wallet

import (
	"github.com/awesome-goose/goose/modules/router"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
)

// ROUTES act on the verified caller's own wallet (coreauth.Caller): the user, client and
// workspace come from the token. Before PLAN U26 they answered anonymous callers.
var ROUTES = router.ForRoutes(
	router.Post("/wallet/init", []any{WalletController{}, "Init"}, coreauth.Caller()...),
	router.Get("/wallet/detail", []any{WalletController{}, "Detail"}, coreauth.Caller()...),
	router.Get("/wallet/balance", []any{WalletController{}, "Balance"}, coreauth.Caller()...),
	router.Post("/wallet/upgrade", []any{WalletController{}, "Upgrade"}, coreauth.Caller()...),
	router.Get("/wallet/transactions", []any{WalletController{}, "Transactions"}, coreauth.Caller()...),
)
