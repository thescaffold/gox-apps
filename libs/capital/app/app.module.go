package app

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
	capitalaccount "github.com/thescaffold/gox-apps-capital/app/account"
	capitalpayment "github.com/thescaffold/gox-apps-capital/app/payment"
	capitalpaymentlog "github.com/thescaffold/gox-apps-capital/app/paymentlog"
	capitalplan "github.com/thescaffold/gox-apps-capital/app/plan"
	capitalplantype "github.com/thescaffold/gox-apps-capital/app/plantype"
	capitalprovider "github.com/thescaffold/gox-apps-capital/app/provider"
	capitalrate "github.com/thescaffold/gox-apps-capital/app/rate"
	capitaltransaction "github.com/thescaffold/gox-apps-capital/app/transaction"
	capitalusage "github.com/thescaffold/gox-apps-capital/app/usage"
	capitalvoucher "github.com/thescaffold/gox-apps-capital/app/voucher"
	capitalvouchertype "github.com/thescaffold/gox-apps-capital/app/vouchertype"
	capitalwallet "github.com/thescaffold/gox-apps-capital/app/wallet"
	"github.com/thescaffold/gox-apps-capital/migrations"
	"github.com/thescaffold/gox-packages-core/module"
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
