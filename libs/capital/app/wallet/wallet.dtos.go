package wallet

import ntxctx "github.com/thescaffold/gox-packages/libs/core/context"

type WalletInitDto struct {
	NTX   ntxctx.NTXContext `context:"ntx"`
	Type  string            `json:"type"`
	Label string            `json:"label"`
}

type WalletDetailDto struct {
	NTX ntxctx.NTXContext `context:"ntx"`
}

type WalletBalanceDto struct {
	NTX ntxctx.NTXContext `context:"ntx"`
}

type WalletUpgradeDto struct {
	NTX          ntxctx.NTXContext `context:"ntx"`
	DailyLimit   int               `json:"dailyLimit"`
	MonthlyLimit int               `json:"monthlyLimit"`
}

type WalletTransactionsDto struct {
	NTX     ntxctx.NTXContext `context:"ntx"`
	Page    int               `form:"page"`
	PerPage int               `form:"perPage"`
}
