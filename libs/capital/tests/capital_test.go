package tests

import (
	"testing"

	goosetest "github.com/awesome-goose/goose/testing"
	"github.com/thescaffold/gox-apps-capital/app"
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
)

func TestAccountEntity(t *testing.T) {
	goosetest.NewSuiteRunner(t, &AccountEntitySuite{}).Run()
}

type AccountEntitySuite struct{ goosetest.Suite }

func (s *AccountEntitySuite) TestTableName() {
	s.T.Expect(capitalaccount.Account{}.TableName()).ToEqual("CapitalAccounts")
}

func TestProviderEntity(t *testing.T) {
	goosetest.NewSuiteRunner(t, &ProviderEntitySuite{}).Run()
}

type ProviderEntitySuite struct{ goosetest.Suite }

func (s *ProviderEntitySuite) TestTableName() {
	s.T.Expect(capitalprovider.Provider{}.TableName()).ToEqual("CapitalProviders")
}

func TestRateEntity(t *testing.T) {
	goosetest.NewSuiteRunner(t, &RateEntitySuite{}).Run()
}

type RateEntitySuite struct{ goosetest.Suite }

func (s *RateEntitySuite) TestTableName() {
	s.T.Expect(capitalrate.Rate{}.TableName()).ToEqual("CapitalRates")
}

func TestPlanTypeEntity(t *testing.T) {
	goosetest.NewSuiteRunner(t, &PlanTypeEntitySuite{}).Run()
}

type PlanTypeEntitySuite struct{ goosetest.Suite }

func (s *PlanTypeEntitySuite) TestTableName() {
	s.T.Expect(capitalplantype.PlanType{}.TableName()).ToEqual("CapitalPlanTypes")
}

func TestPlanEntity(t *testing.T) {
	goosetest.NewSuiteRunner(t, &PlanEntitySuite{}).Run()
}

type PlanEntitySuite struct{ goosetest.Suite }

func (s *PlanEntitySuite) TestTableName() {
	s.T.Expect(capitalplan.Plan{}.TableName()).ToEqual("CapitalPlans")
}

func TestPaymentEntity(t *testing.T) {
	goosetest.NewSuiteRunner(t, &PaymentEntitySuite{}).Run()
}

type PaymentEntitySuite struct{ goosetest.Suite }

func (s *PaymentEntitySuite) TestTableName() {
	s.T.Expect(capitalpayment.Payment{}.TableName()).ToEqual("CapitalPayments")
}

func TestPaymentLogEntity(t *testing.T) {
	goosetest.NewSuiteRunner(t, &PaymentLogEntitySuite{}).Run()
}

type PaymentLogEntitySuite struct{ goosetest.Suite }

func (s *PaymentLogEntitySuite) TestTableName() {
	s.T.Expect(capitalpaymentlog.PaymentLog{}.TableName()).ToEqual("CapitalPaymentLogs")
}

func TestTransactionEntity(t *testing.T) {
	goosetest.NewSuiteRunner(t, &TransactionEntitySuite{}).Run()
}

type TransactionEntitySuite struct{ goosetest.Suite }

func (s *TransactionEntitySuite) TestTableName() {
	s.T.Expect(capitaltransaction.Transaction{}.TableName()).ToEqual("CapitalTransactions")
}

func TestUsageEntity(t *testing.T) {
	goosetest.NewSuiteRunner(t, &UsageEntitySuite{}).Run()
}

type UsageEntitySuite struct{ goosetest.Suite }

func (s *UsageEntitySuite) TestTableName() {
	s.T.Expect(capitalusage.Usage{}.TableName()).ToEqual("CapitalUsages")
}

func TestVoucherTypeEntity(t *testing.T) {
	goosetest.NewSuiteRunner(t, &VoucherTypeEntitySuite{}).Run()
}

type VoucherTypeEntitySuite struct{ goosetest.Suite }

func (s *VoucherTypeEntitySuite) TestTableName() {
	s.T.Expect(capitalvouchertype.VoucherType{}.TableName()).ToEqual("CapitalVoucherTypes")
}

func TestVoucherEntity(t *testing.T) {
	goosetest.NewSuiteRunner(t, &VoucherEntitySuite{}).Run()
}

type VoucherEntitySuite struct{ goosetest.Suite }

func (s *VoucherEntitySuite) TestTableName() {
	s.T.Expect(capitalvoucher.Voucher{}.TableName()).ToEqual("CapitalVouchers")
}

func TestAppModule(t *testing.T) {
	goosetest.NewSuiteRunner(t, &AppModuleSuite{}).Run()
}

type AppModuleSuite struct{ goosetest.Suite }

func (s *AppModuleSuite) TestMigrations_Count() {
	s.T.Expect(len(app.Migrations)).ToEqual(11)
}

func (s *AppModuleSuite) TestName() {
	s.T.Expect(app.Name).ToEqual("capital")
}
