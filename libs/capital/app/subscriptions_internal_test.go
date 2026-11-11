package app

import (
	"os"
	"testing"
	"time"

	"github.com/awesome-goose/goose/modules/sql"
	"github.com/google/uuid"
	capitalaccount "github.com/thescaffold/gox-apps/libs/capital/app/account"
	capitalpayment "github.com/thescaffold/gox-apps/libs/capital/app/payment"
	capitalplan "github.com/thescaffold/gox-apps/libs/capital/app/plan"
	capitalplantype "github.com/thescaffold/gox-apps/libs/capital/app/plantype"
	capitalrate "github.com/thescaffold/gox-apps/libs/capital/app/rate"
	capitalusage "github.com/thescaffold/gox-apps/libs/capital/app/usage"
	capitalwallet "github.com/thescaffold/gox-apps/libs/capital/app/wallet"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// The bus handlers a host turns on with events.RegisterSubscriptions("capital", Subscriptions)
// (Yan keeps them off): what each one writes when its event fires. Needs CAPITAL_TEST_DB_URL
// (a scratch Postgres database; the app's migrations run on it).
func subscriptionsFixture(t *testing.T) (*AppService, *gorm.DB) {
	t.Helper()
	dsn := os.Getenv("CAPITAL_TEST_DB_URL")
	if dsn == "" {
		t.Skip("CAPITAL_TEST_DB_URL not set; skipping integration test")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(2)
	t.Cleanup(func() { _ = sqlDB.Close() })
	q := (&sql.Query{}).With(&sql.Db{DB: db})
	for _, m := range AllMigrations {
		if err := m.Run(q); err != nil {
			t.Fatalf("migrate %T: %v", m, err)
		}
	}
	accounts := &capitalaccount.AccountEntity{Entity: (&sql.Entity[capitalaccount.Account]{}).With(q)}
	svc := &AppService{
		usageEntity:    &capitalusage.UsageEntity{Entity: (&sql.Entity[capitalusage.Usage]{}).With(q)},
		rateEntity:     &capitalrate.RateEntity{Entity: (&sql.Entity[capitalrate.Rate]{}).With(q)},
		planEntity:     &capitalplan.PlanEntity{Entity: (&sql.Entity[capitalplan.Plan]{}).With(q)},
		planTypeEntity: &capitalplantype.PlanTypeEntity{Entity: (&sql.Entity[capitalplantype.PlanType]{}).With(q)},
		paymentEntity:  &capitalpayment.PaymentEntity{Entity: (&sql.Entity[capitalpayment.Payment]{}).With(q)},
		walletService:  capitalwallet.NewWalletServiceForTest(accounts),
	}
	accounts.OnRegister()
	svc.usageEntity.OnRegister()
	svc.rateEntity.OnRegister()
	svc.planEntity.OnRegister()
	svc.planTypeEntity.OnRegister()
	svc.paymentEntity.OnRegister()
	return svc, db
}

func countPayments(t *testing.T, db *gorm.DB, where string, args ...any) int64 {
	t.Helper()
	var n int64
	if err := db.Model(&capitalpayment.Payment{}).Where(where, args...).Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	return n
}

// apps.capital.payment.debt announced the total a plan owes and the handler saved that total as
// one more unpaid row, so every debt run doubled the debt. It now records nothing.
func TestDebtAnnouncementDoesNotAddDebt(t *testing.T) {
	svc, db := subscriptionsFixture(t)
	user := "u-" + uuid.NewString()[:8]
	Subscriptions["apps.capital.payment.debt"]("apps.capital.payment.debt", map[string]any{"userId": user, "clientId": "c", "workspaceId": "w", "amount": 500.0})
	capitalAppSvc = svc
	defer func() { capitalAppSvc = nil }()
	Subscriptions["apps.capital.payment.debt"]("apps.capital.payment.debt", map[string]any{"userId": user, "clientId": "c", "workspaceId": "w", "amount": 500.0})
	if n := countPayments(t, db, `"user_id" = ?`, user); n != 0 {
		t.Fatalf("a debt announcement wrote %d payment rows", n)
	}
}

// PaymentService announces a payment it has already saved; recording it again duplicated it.
// A payment another service reports is recorded once per reference.
func TestPaymentPayIsRecordedOnce(t *testing.T) {
	svc, db := subscriptionsFixture(t)
	capitalAppSvc = svc
	defer func() { capitalAppSvc = nil }()
	user, ref := "u-"+uuid.NewString()[:8], "ref-"+uuid.NewString()[:8]
	pay := Subscriptions["apps.capital.payment.pay"]
	pay("apps.capital.payment.pay", map[string]any{"status": true, "type": "charge", "plan": map[string]any{"userId": user}, "payment": map[string]any{"id": "p1"}, "userId": user, "reference": ref})
	if n := countPayments(t, db, `"user_id" = ?`, user); n != 0 {
		t.Fatalf("PaymentService's own announcement was recorded again (%d rows)", n)
	}
	report := map[string]any{"userId": user, "clientId": "c", "workspaceId": "w", "amount": 300.0, "reference": ref}
	pay("apps.capital.payment.pay", report)
	pay("apps.capital.payment.pay", report)
	if n := countPayments(t, db, `"user_id" = ? AND "reference" = ?`, user, ref); n != 1 {
		t.Fatalf("a reported payment was recorded %d times, want once", n)
	}
}

// UsageService sends {entityName, entityId, quantity} with no rateId, which the handler required,
// so usage events never recorded anything.
func TestUsageEventsFromUsageServiceAreRecorded(t *testing.T) {
	svc, db := subscriptionsFixture(t)
	capitalAppSvc = svc
	defer func() { capitalAppSvc = nil }()
	user, entity, kind := "u-"+uuid.NewString()[:8], uuid.NewString(), "seat-"+uuid.NewString()[:6]
	if err := svc.rateEntity.Insert(&capitalrate.Rate{Type: kind, Currency: "USD"}); err != nil {
		t.Fatal(err)
	}
	Subscriptions["apps.capital.usage.start"]("apps.capital.usage.start", map[string]any{"userId": user, "clientId": "c", "workspaceId": "w", "entityName": kind, "entityId": entity, "quantity": 3.0})
	var rows []capitalusage.Usage
	if err := db.Where(`"user_id" = ?`, user).Find(&rows).Error; err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Quantity != 3 || rows[0].RateId == "" || rows[0].EntityName != kind {
		t.Fatalf("usage.start recorded %+v", rows)
	}
	Subscriptions["apps.capital.usage.stop"]("apps.capital.usage.stop", map[string]any{"userId": user, "entityName": kind, "entityId": entity})
	var stopped capitalusage.Usage
	if err := db.Where(`"user_id" = ?`, user).First(&stopped).Error; err != nil || stopped.StopAt == nil {
		t.Fatalf("usage.stop did not close the row: %+v %v", stopped, err)
	}
}

// Heartbeats without a queue accrue each plan once per period; a re-run must not bill twice.
func TestHeartbeatAccruesOncePerPeriod(t *testing.T) {
	svc, db := subscriptionsFixture(t)
	pt := &capitalplantype.PlanType{ClientId: "c-" + uuid.NewString()[:6], Key: "k-" + uuid.NewString()[:6], Name: "Pro", Currency: "USD", Monthly: 900}
	if err := svc.planTypeEntity.Insert(pt); err != nil {
		t.Fatal(err)
	}
	period := "weekly"
	plan := &capitalplan.Plan{UserId: "u-" + uuid.NewString()[:8], ClientId: pt.ClientId, WorkspaceId: "w", TypeId: pt.Id, PeriodType: &period}
	if err := svc.planEntity.Insert(plan); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		if err := svc.OnWeeklyHeartbeat(); err != nil {
			t.Fatal(err)
		}
	}
	if n := countPayments(t, db, `"plan_id" = ?`, plan.Id); n != 1 {
		t.Fatalf("three weekly heartbeats in one week wrote %d accruals, want 1", n)
	}
	if k := periodKey("weekly", time.Date(2026, 11, 11, 0, 0, 0, 0, time.UTC)); k != "2026-W46" {
		t.Fatalf("periodKey weekly = %s", k)
	}
}

// identity's apps.capital.wallet.init and the membership row's apps.db event arrive together at
// sign-up; the person must end up with one account.
func TestConcurrentWalletInitMakesOneAccount(t *testing.T) {
	svc, db := subscriptionsFixture(t)
	user := "u-" + uuid.NewString()[:8]
	done := make(chan error, 2)
	for i := 0; i < 2; i++ {
		go func() {
			done <- svc.OnWalletInit(WalletInitPayload{UserID: user, ClientID: "c", WorkspaceID: "w", Currency: "USD"})
		}()
	}
	for i := 0; i < 2; i++ {
		if err := <-done; err != nil {
			t.Fatal(err)
		}
	}
	var n int64
	db.Model(&capitalaccount.Account{}).Where(`"user_id" = ?`, user).Count(&n)
	if n != 1 {
		t.Fatalf("two concurrent wallet inits made %d accounts", n)
	}
}
