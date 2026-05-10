package app

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
	capitalaccount "github.com/thescaffold/gox-apps/libs/capital/app/account"
	capitalpayment "github.com/thescaffold/gox-apps/libs/capital/app/payment"
	capitalpaymentlog "github.com/thescaffold/gox-apps/libs/capital/app/paymentlog"
	capitalplan "github.com/thescaffold/gox-apps/libs/capital/app/plan"
	capitalplantype "github.com/thescaffold/gox-apps/libs/capital/app/plantype"
	capitalprovider "github.com/thescaffold/gox-apps/libs/capital/app/provider"
	capitalrate "github.com/thescaffold/gox-apps/libs/capital/app/rate"
	capitaltransaction "github.com/thescaffold/gox-apps/libs/capital/app/transaction"
	capitalusage "github.com/thescaffold/gox-apps/libs/capital/app/usage"
	capitalvoucher "github.com/thescaffold/gox-apps/libs/capital/app/voucher"
	capitalvouchertype "github.com/thescaffold/gox-apps/libs/capital/app/vouchertype"
	capitalwallet "github.com/thescaffold/gox-apps/libs/capital/app/wallet"
	"github.com/thescaffold/gox-apps/libs/capital/migrations"
	"github.com/thescaffold/gox-packages/libs/core/module"
)

var Migrations = []sql.Migration{
	&migrations.CreateCapitalAccounts{},
	&migrations.CreateCapitalProviders{},
	&migrations.CreateCapitalRates{},
	&migrations.CreateCapitalPlanTypes{},
	&migrations.CreateCapitalPlans{},
	&migrations.CreateCapitalPayments{},
	&migrations.CreateCapitalPaymentLogs{},
	&migrations.CreateCapitalTransactions{},
	&migrations.CreateCapitalUsages{},
	&migrations.CreateCapitalVoucherTypes{},
	&migrations.CreateCapitalVouchers{},
}

type AppModule struct{}

func (m *AppModule) Imports() []types.Module {
	return []types.Module{
		module.New(module.CoreConfig{}),
		sql.Child(&sql.Config{Migrations: Migrations}),
		&capitalaccount.AccountModule{},
		&capitalprovider.ProviderModule{},
		&capitalrate.RateModule{},
		&capitalplantype.PlanTypeModule{},
		&capitalplan.PlanModule{},
		&capitalpayment.PaymentModule{},
		&capitalpaymentlog.PaymentLogModule{},
		&capitaltransaction.TransactionModule{},
		&capitalusage.UsageModule{},
		&capitalvouchertype.VoucherTypeModule{},
		&capitalvoucher.VoucherModule{},
		&capitalwallet.WalletModule{},
		ROUTES,
	}
}

func (m *AppModule) Exports() []any { return []any{&AppService{}} }

func (m *AppModule) Declarations() []any {
	svc := &AppService{}
	capitalAppSvc = svc
	return []any{&AppController{}, svc}
}
