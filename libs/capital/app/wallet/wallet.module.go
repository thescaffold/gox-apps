package wallet

import (
	"github.com/awesome-goose/goose/types"
	capitalaccount "github.com/thescaffold/gox-apps-capital/app/account"
	capitaltransaction "github.com/thescaffold/gox-apps-capital/app/transaction"
)

type WalletModule struct{}

func (m *WalletModule) Imports() []types.Module {
	return []types.Module{
		ROUTES,
		&capitalaccount.AccountModule{},
		&capitaltransaction.TransactionModule{},
	}
}

func (m *WalletModule) Exports() []any { return []any{&WalletService{}} }

func (m *WalletModule) Declarations() []any {
	return []any{&WalletController{}, &WalletService{}}
}
