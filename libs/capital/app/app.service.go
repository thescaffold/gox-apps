package app

import (
	"encoding/json"
	"time"

	goqueues "github.com/awesome-goose/goose/modules/queues"

	capitalpayment "github.com/thescaffold/gox-apps/libs/capital/app/payment"
	capitalpaymentlog "github.com/thescaffold/gox-apps/libs/capital/app/paymentlog"
	capitalplan "github.com/thescaffold/gox-apps/libs/capital/app/plan"
	capitalplantype "github.com/thescaffold/gox-apps/libs/capital/app/plantype"
	capitalrate "github.com/thescaffold/gox-apps/libs/capital/app/rate"
	capitalusage "github.com/thescaffold/gox-apps/libs/capital/app/usage"
	capitalwallet "github.com/thescaffold/gox-apps/libs/capital/app/wallet"
	flagsapp "github.com/thescaffold/gox-apps/libs/flags/app/flag"
	queueapp "github.com/thescaffold/gox-apps/libs/queue/app"
	"github.com/thescaffold/gox-packages/libs/core/events"
	"github.com/thescaffold/gox-packages/libs/core/services"
)

// FindOrCreateUserPlan mirrors TS AppService.findOrCreateUserPlan — returns an
// existing Plan for the (userId, clientId, workspaceId) tuple or creates one
// keyed to the default plan-type ("lite") with monthly period. After insert,
// registers the planType.meta.flags map via FlagService when available.
func (s *AppService) FindOrCreateUserPlan(userID, clientID, workspaceID string) (*capitalplan.Plan, error) {
	if s.planEntity == nil {
		return nil, nil
	}
	existing, _ := s.planEntity.First(
		`"user_id" = ? AND "client_id" = ? AND "workspace_id" = ?`,
		userID, clientID, workspaceID,
	)
	if existing != nil {
		return existing, nil
	}
	if s.planTypeEntity == nil {
		return nil, nil
	}
	pt, _ := s.planTypeEntity.First(`"key" = ?`, "lite")
	if pt == nil {
		return nil, nil
	}
	period := "monthly"
	plan := &capitalplan.Plan{
		UserId:      userID,
		ClientId:    clientID,
		WorkspaceId: workspaceID,
		TypeId:      pt.Id,
		PeriodType:  &period,
	}
	if err := s.planEntity.Insert(plan); err != nil {
		return nil, err
	}
	s.registerPlanTypeFlags(pt, userID, clientID, workspaceID)
	return plan, nil
}

// registerPlanTypeFlags decodes planType.meta.flags and seeds the user's
// per-flag rows via FlagService. Mirrors TS object-entries iteration over
// `defaultPlanType?.meta?.flags`.
func (s *AppService) registerPlanTypeFlags(pt *capitalplantype.PlanType, userID, clientID, workspaceID string) {
	if pt == nil || s.flagService == nil || len(pt.Meta) == 0 {
		return
	}
	var meta map[string]any
	if err := json.Unmarshal(pt.Meta, &meta); err != nil {
		return
	}
	flagsMap, _ := meta["flags"].(map[string]any)
	if len(flagsMap) == 0 {
		return
	}
	dtos := flagsapp.FromMeta(userID, clientID, workspaceID, flagsMap)
	_, _ = s.flagService.Register(dtos)
}

// PlanType returns the PlanType linked to a Plan, or nil when missing.
func (s *AppService) PlanType(plan *capitalplan.Plan) (*capitalplantype.PlanType, error) {
	if plan == nil || s.planTypeEntity == nil {
		return nil, nil
	}
	return s.planTypeEntity.First(`"id" = ?`, plan.TypeId)
}

// ChangeUserPlan mirrors TS AppService.changeUserPlan — switches the user's
// Plan to the (currency, key) PlanType. Returns nil when no such PlanType
// exists, when periods would downgrade yearly→monthly, or when amount would
// downgrade (TS-equivalent guards).
func (s *AppService) ChangeUserPlan(userID, clientID, workspaceID, currency, periodType, key string) (any, error) {
	if s.planTypeEntity == nil || s.planEntity == nil {
		return nil, nil
	}
	// Lock the user's plan-change slot for 10s; abort if another caller is
	// mid-flight. Mirrors TS cacheService.lock('change-user-plan', userId, 10).
	release, lockErr := services.AcquireLock(s.cache, "change-user-plan", userID, 10*time.Second, true)
	if lockErr != nil {
		return nil, nil
	}
	defer release()
	pt, _ := s.planTypeEntity.First(`"currency" = ? AND "key" = ?`, currency, key)
	if pt == nil {
		return nil, nil
	}
	plan, err := s.FindOrCreateUserPlan(userID, clientID, workspaceID)
	if err != nil || plan == nil {
		return nil, err
	}
	existingType, _ := s.PlanType(plan)
	if existingType != nil && existingType.Id == pt.Id {
		return nil, nil
	}
	if plan.PeriodType != nil && *plan.PeriodType == "yearly" && periodType == "monthly" {
		return nil, nil
	}
	if existingType != nil && periodAmount(existingType, periodType) > periodAmount(pt, periodType) {
		return nil, nil
	}
	plan.TypeId = pt.Id
	plan.PeriodType = &periodType
	if _, err := s.planEntity.Update(plan, `"id" = ?`, plan.Id); err != nil {
		return nil, err
	}
	// Re-register flags from the *previous* plan-type's meta — TS does this on
	// existingPlanType (not the new one) so the user keeps their old caps until
	// the new period kicks in.
	s.registerPlanTypeFlags(existingType, userID, clientID, workspaceID)
	return plan, nil
}

// Subscribe mirrors TS AppService.subscribe — records a record_only Payment
// for the existing plan if the supplied amount matches plan.type[periodType].
func (s *AppService) Subscribe(userID, clientID, workspaceID, currency string, amount float64, paymentSvc *capitalpayment.PaymentService) (any, error) {
	// Same change-user-plan slot as ChangeUserPlan — TS shares the lock so
	// a user can't subscribe and switch plans at the same time.
	release, lockErr := services.AcquireLock(s.cache, "change-user-plan", userID, 10*time.Second, true)
	if lockErr != nil {
		return nil, nil
	}
	defer release()
	plan, err := s.FindOrCreateUserPlan(userID, clientID, workspaceID)
	if err != nil || plan == nil {
		return nil, err
	}
	pt, _ := s.PlanType(plan)
	if pt == nil || plan.PeriodType == nil {
		return nil, nil
	}
	if amount != float64(periodAmount(pt, *plan.PeriodType)) {
		return nil, nil
	}
	if paymentSvc == nil {
		return map[string]any{"paid": false, "reference": ""}, nil
	}
	meta := map[string]any{
		"type": "apps.capital.record.pay",
		"plan": plan,
	}
	paid, ref, err := paymentSvc.New([]string{string(PaymentTypeRecordOnly)}, plan, amount, meta, "", "", nil)
	if err != nil {
		return nil, err
	}
	return map[string]any{"paid": paid, "reference": ref}, nil
}

// ComputePlanUpgrade mirrors TS AppController.computePlanUpgrade — returns
// (deduction, newAmount, diff) prorated against the current period.
func (s *AppService) ComputePlanUpgrade(periodType string, existing, target *capitalplantype.PlanType) (float64, float64, float64, bool) {
	totalDays := 0
	daysElapsed := 0
	now := time.Now().UTC()
	switch periodType {
	case "monthly":
		totalDays = 28
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		daysElapsed = int(now.Sub(start).Hours() / 24)
	case "yearly":
		totalDays = 365
		start := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
		daysElapsed = int(now.Sub(start).Hours() / 24)
	default:
		return 0, 0, 0, false
	}
	if daysElapsed > totalDays {
		daysElapsed = totalDays
	}
	deduction := 0.0
	if existing != nil {
		deduction = (float64(periodAmount(existing, periodType)) / float64(totalDays)) * float64(daysElapsed)
	}
	newAmount := float64(periodAmount(target, periodType))
	diff := newAmount - deduction
	return deduction, newAmount, diff, true
}

func periodAmount(pt *capitalplantype.PlanType, periodType string) int {
	if pt == nil {
		return 0
	}
	switch periodType {
	case "daily":
		if pt.Daily != nil {
			return *pt.Daily
		}
	case "weekly":
		if pt.Weekly != nil {
			return *pt.Weekly
		}
	case "monthly":
		return pt.Monthly
	case "yearly":
		if pt.Yearly != nil {
			return *pt.Yearly
		}
	}
	return 0
}

// AppService surfaces capital event handlers. Mirrors
// ntx-apps/libs/capital/src/app.service.ts + the subscription bodies in
// app.controller.ts. Sub-services are injected by goose; tests can construct
// AppService without DI and handlers will no-op gracefully.
type AppService struct {
	walletService    *capitalwallet.WalletService        `inject:""`
	usageEntity      *capitalusage.UsageEntity           `inject:""`
	rateEntity       *capitalrate.RateEntity             `inject:""`
	planEntity       *capitalplan.PlanEntity             `inject:""`
	planTypeEntity   *capitalplantype.PlanTypeEntity     `inject:""`
	paymentEntity    *capitalpayment.PaymentEntity       `inject:""`
	paymentLogEntity *capitalpaymentlog.PaymentLogEntity `inject:""`
	queueService     *queueapp.AppService                `inject:""`
	flagService      *flagsapp.FlagService               `inject:""`
	cache            services.CacheBackend               `inject:""`
	paymentService   *capitalpayment.PaymentService      `inject:""`
	tracker          *events.TrackerService              `inject:""`
}

func (s *AppService) GetHello() string { return "Hello World!" }

// IdentityUserClientWorkspacePayload is the shape of
// apps.db.identityuserclientworkspaces.after-insert event payloads.
type IdentityUserClientWorkspacePayload struct {
	UserID      string `json:"userId"`
	ClientID    string `json:"clientId"`
	WorkspaceID string `json:"workspaceId"`
	Currency    string `json:"currency"`
}

// OnIdentityUserClientWorkspaceCreated initializes the user's wallet for the
// workspace currency. Mirrors TS subscription body which calls walletService.init().
func (s *AppService) OnIdentityUserClientWorkspaceCreated(p IdentityUserClientWorkspacePayload) error {
	if s.walletService == nil || p.UserID == "" {
		return nil
	}
	_, err := s.walletService.Init(p.UserID, p.ClientID, p.WorkspaceID, p.Currency, "default", "default")
	return err
}

// WalletInitPayload is the apps.capital.wallet.init event payload.
type WalletInitPayload struct {
	UserID      string `json:"userId"`
	ClientID    string `json:"clientId"`
	WorkspaceID string `json:"workspaceId"`
	Currency    string `json:"currency"`
	Type        string `json:"type,omitempty"`
	Label       string `json:"label,omitempty"`
}

// OnWalletInit ensures a wallet exists for the (user, client, workspace, currency) tuple.
func (s *AppService) OnWalletInit(p WalletInitPayload) error {
	if s.walletService == nil || p.UserID == "" {
		return nil
	}
	typ := p.Type
	if typ == "" {
		typ = "default"
	}
	label := p.Label
	if label == "" {
		label = "default"
	}
	_, err := s.walletService.Init(p.UserID, p.ClientID, p.WorkspaceID, p.Currency, typ, label)
	return err
}

// UsageStartPayload is the apps.capital.usage.start/update/stop payload.
type UsageStartPayload struct {
	UserID      string `json:"userId"`
	ClientID    string `json:"clientId"`
	WorkspaceID string `json:"workspaceId"`
	Service     string `json:"service"`
	Entity      string `json:"entity"`
	EntityID    string `json:"entityId"`
	Reference   string `json:"reference,omitempty"`
	RateID      string `json:"rateId,omitempty"`
	Quantity    int    `json:"quantity,omitempty"`
}

// OnUsageStart opens a usage record. Inserts a fresh row keyed by
// (userId, rateId, entityId).
func (s *AppService) OnUsageStart(p UsageStartPayload) error {
	if s.usageEntity == nil || p.UserID == "" || p.RateID == "" {
		return nil
	}
	now := time.Now().UTC()
	q := p.Quantity
	if q <= 0 {
		q = 1
	}
	return s.usageEntity.Insert(&capitalusage.Usage{
		UserId:      p.UserID,
		ClientId:    p.ClientID,
		WorkspaceId: p.WorkspaceID,
		RateId:      p.RateID,
		EntityId:    p.EntityID,
		EntityName:  p.Entity,
		Quantity:    q,
		StartAt:     &now,
	})
}

// OnUsageUpdate increments quantity on the latest open usage row.
func (s *AppService) OnUsageUpdate(p UsageStartPayload) error {
	if s.usageEntity == nil || p.UserID == "" {
		return nil
	}
	row, _ := s.usageEntity.First(
		`user_id = ? AND entity_id = ? AND stop_at IS NULL`,
		p.UserID, p.EntityID,
	)
	if row == nil {
		return nil
	}
	q := p.Quantity
	if q <= 0 {
		q = 1
	}
	row.Quantity += q
	_, err := s.usageEntity.Update(row, `id = ?`, row.Id)
	return err
}

// OnUsageStop closes the latest open usage row.
func (s *AppService) OnUsageStop(p UsageStartPayload) error {
	if s.usageEntity == nil || p.UserID == "" {
		return nil
	}
	row, _ := s.usageEntity.First(
		`user_id = ? AND entity_id = ? AND stop_at IS NULL`,
		p.UserID, p.EntityID,
	)
	if row == nil {
		return nil
	}
	now := time.Now().UTC()
	row.StopAt = &now
	_, err := s.usageEntity.Update(row, `id = ?`, row.Id)
	return err
}

// RatePayload is the apps.capital.rate.register event payload.
type RatePayload struct {
	Type     string `json:"type,omitempty"`
	Currency string `json:"currency"`
	PerUnit  *int   `json:"perUnit,omitempty"`
}

// OnRateRegister upserts a rate row keyed by (type, currency).
func (s *AppService) OnRateRegister(p RatePayload) error {
	if s.rateEntity == nil || p.Currency == "" {
		return nil
	}
	typ := p.Type
	if typ == "" {
		typ = "default"
	}
	existing, _ := s.rateEntity.First(`type = ? AND currency = ?`, typ, p.Currency)
	if existing != nil {
		existing.PerUnit = p.PerUnit
		_, err := s.rateEntity.Update(existing, `id = ?`, existing.Id)
		return err
	}
	return s.rateEntity.Insert(&capitalrate.Rate{
		Type: typ, Currency: p.Currency, PerUnit: p.PerUnit,
	})
}

// PlanPayload is the apps.capital.plan.register event payload.
type PlanPayload struct {
	UserID      string `json:"userId,omitempty"`
	ClientID    string `json:"clientId"`
	WorkspaceID string `json:"workspaceId,omitempty"`
	Key         string `json:"key,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Currency    string `json:"currency,omitempty"`
	Monthly     int    `json:"monthly,omitempty"`
	Yearly      *int   `json:"yearly,omitempty"`
	PeriodType  string `json:"periodType,omitempty"`
	Type        string `json:"type,omitempty"`
}

// OnPlanRegister upserts the plan-type, then a plan row for the supplied
// owner. Mirrors TS body that does updateOrCreate on PlanType then Plan.
func (s *AppService) OnPlanRegister(p PlanPayload) error {
	if s.planTypeEntity == nil || s.planEntity == nil || p.ClientID == "" || p.Name == "" {
		return nil
	}
	key := p.Key
	if key == "" {
		key = p.Name
	}
	currency := p.Currency
	if currency == "" {
		currency = "USD"
	}
	pt, _ := s.planTypeEntity.First(`client_id = ? AND "key" = ?`, p.ClientID, key)
	if pt == nil {
		pt = &capitalplantype.PlanType{
			ClientId: p.ClientID,
			Key:      key,
			Name:     p.Name,
			Currency: currency,
			Monthly:  p.Monthly,
			Yearly:   p.Yearly,
		}
		if p.Description != "" {
			pt.Desc = &p.Description
		}
		if p.Type != "" {
			pt.Type = &p.Type
		}
		if err := s.planTypeEntity.Insert(pt); err != nil {
			return err
		}
	} else {
		pt.Name = p.Name
		pt.Currency = currency
		pt.Monthly = p.Monthly
		pt.Yearly = p.Yearly
		if p.Description != "" {
			pt.Desc = &p.Description
		}
		if p.Type != "" {
			pt.Type = &p.Type
		}
		if _, err := s.planTypeEntity.Update(pt, `id = ?`, pt.Id); err != nil {
			return err
		}
	}

	// Insert (or skip) Plan row for the owner if userId is provided.
	if p.UserID == "" {
		return nil
	}
	existing, _ := s.planEntity.First(`user_id = ? AND workspace_id = ? AND type_id = ?`, p.UserID, p.WorkspaceID, pt.Id)
	if existing != nil {
		return nil
	}
	plan := &capitalplan.Plan{
		UserId: p.UserID, ClientId: p.ClientID, WorkspaceId: p.WorkspaceID, TypeId: pt.Id,
	}
	if p.PeriodType != "" {
		plan.PeriodType = &p.PeriodType
	}
	return s.planEntity.Insert(plan)
}

// PaymentPayload is the apps.capital.payment.pay / .debt event payload.
type PaymentPayload struct {
	UserID      string  `json:"userId"`
	ClientID    string  `json:"clientId"`
	WorkspaceID string  `json:"workspaceId"`
	PlanID      string  `json:"planId,omitempty"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency,omitempty"`
	Provider    string  `json:"provider,omitempty"`
	Reference   string  `json:"reference,omitempty"`
	Plan        string  `json:"plan,omitempty"`
}

// OnPaymentPay records a successful payment. Inserts a Payment row marked paid.
func (s *AppService) OnPaymentPay(p PaymentPayload) error {
	return s.recordPayment(p, "wallet", true)
}

// OnPaymentDebt records a failed/owed payment as type "debt".
func (s *AppService) OnPaymentDebt(p PaymentPayload) error {
	return s.recordPayment(p, "debt", false)
}

func (s *AppService) recordPayment(p PaymentPayload, typ string, paid bool) error {
	if s.paymentEntity == nil || p.UserID == "" {
		return nil
	}
	amount := int(p.Amount)
	pay := &capitalpayment.Payment{
		UserId:      p.UserID,
		ClientId:    p.ClientID,
		WorkspaceId: p.WorkspaceID,
		PlanId:      p.PlanID,
		Amount:      &amount,
		Type:        typ,
		Reference:   p.Reference,
		Paid:        &paid,
	}
	if p.Currency != "" {
		pay.Currency = &p.Currency
	}
	return s.paymentEntity.Insert(pay)
}

// OnPaymentJob handles `queue/apps/capital/payment` — TS-equivalent dispatch:
// extract plan from job.data, look up its planType.monthly/yearly amount,
// invoke PaymentService.New(Charge) which drives provider charge + wallet
// + receipt. Returns the (paid, reference) outcome so the queue worker
// records it on the JobLog row.
func (s *AppService) OnPaymentJob(job *goqueues.QueueJob) (any, error) {
	if job == nil {
		return nil, nil
	}
	plan, err := s.planFromJob(job)
	if err != nil || plan == nil {
		return map[string]any{"jobId": job.Id, "status": "skipped", "queue": "payment"}, nil
	}
	periodType := "monthly"
	if plan.PeriodType != nil && *plan.PeriodType != "" {
		periodType = *plan.PeriodType
	}
	pt, _ := s.planTypeEntity.First(`"id" = ?`, plan.TypeId)
	amount := 0.0
	if pt != nil {
		amount = float64(periodAmount(pt, periodType))
	}
	if amount <= 0 || s.paymentService == nil {
		return map[string]any{"jobId": job.Id, "status": "skipped", "queue": "payment", "reason": "no amount"}, nil
	}
	meta := map[string]any{
		"type":       "apps.capital.payment.renewal",
		"plan":       plan,
		"planTypeId": plan.TypeId,
		"periodType": periodType,
	}
	paid, ref, err := s.paymentService.New(
		[]string{capitalpayment.TypeCharge}, plan, amount, meta, "", "", nil,
	)
	if err != nil {
		return nil, err
	}
	status := "failed"
	if paid {
		status = "success"
	}
	return map[string]any{
		"jobId":     job.Id,
		"queue":     "payment",
		"status":    status,
		"reference": ref,
		"amount":    amount,
	}, nil
}

// OnDebtJob handles `queue/apps/capital/debt` — TS-equivalent dispatch:
// extract plan from job.data, run PaymentService.Debt to total the user's
// outstanding amount, then publish `apps.capital.payment.debt` on the bus
// so downstream listeners (notification, etc) can dispatch dunning emails.
func (s *AppService) OnDebtJob(job *goqueues.QueueJob) (any, error) {
	if job == nil {
		return nil, nil
	}
	plan, err := s.planFromJob(job)
	if err != nil || plan == nil {
		return map[string]any{"jobId": job.Id, "status": "skipped", "queue": "debt"}, nil
	}
	if s.paymentService == nil {
		return map[string]any{"jobId": job.Id, "status": "skipped", "queue": "debt", "reason": "no payment service"}, nil
	}
	debt, err := s.paymentService.Debt(plan)
	if err != nil {
		return nil, err
	}
	if s.tracker != nil && debt != nil && debt.Amount > 0 {
		s.tracker.Message("apps.capital.payment.debt", map[string]any{
			"userId":      plan.UserId,
			"clientId":    plan.ClientId,
			"workspaceId": plan.WorkspaceId,
			"planId":      plan.Id,
			"amount":      debt.Amount,
			"from":        debt.From,
			"to":          debt.To,
		})
	}
	amount := 0.0
	if debt != nil {
		amount = debt.Amount
	}
	return map[string]any{
		"jobId":  job.Id,
		"queue":  "debt",
		"status": "dispatched",
		"amount": amount,
	}, nil
}

// planFromJob decodes the plan envelope from a queue job payload — TS
// producers stuff `{name, page, index, plan}` so plan lives under .plan.
// Falls back to a top-level plan_id when only the bare id is supplied.
func (s *AppService) planFromJob(job *goqueues.QueueJob) (*capitalplan.Plan, error) {
	if job == nil || len(job.Data) == 0 || s.planEntity == nil {
		return nil, nil
	}
	var data map[string]any
	if err := json.Unmarshal(job.Data, &data); err != nil {
		return nil, err
	}
	planId := ""
	if nested, ok := data["plan"].(map[string]any); ok {
		planId, _ = nested["id"].(string)
	}
	if planId == "" {
		planId, _ = data["planId"].(string)
	}
	if planId == "" {
		return nil, nil
	}
	return s.planEntity.First(`"id" = ?`, planId)
}

// OnDailyHeartbeat mirrors TS BatchService.run(PaymentBatch + DebtBatch, DAILY).
// Now uses the real PaymentBatch/DebtBatch fan-out (Phase C) — each plan in
// the daily period gets a `queue/apps/capital/payment` + `debt` job pushed.
// Falls back to the per-plan accrual insert when the queue isn't wired.
func (s *AppService) OnDailyHeartbeat() error {
	return s.runHeartbeatBatch("daily")
}

// OnWeeklyHeartbeat mirrors TS BatchService.run(..., WEEKLY).
func (s *AppService) OnWeeklyHeartbeat() error {
	return s.runHeartbeatBatch("weekly")
}

// OnMonthlyHeartbeat fans out payment + debt jobs for monthly plans.
func (s *AppService) OnMonthlyHeartbeat() error {
	return s.runHeartbeatBatch("monthly")
}

// OnYearlyHeartbeat is the annual counterpart.
func (s *AppService) OnYearlyHeartbeat() error {
	return s.runHeartbeatBatch("yearly")
}

func (s *AppService) runHeartbeatBatch(period string) error {
	if s.planEntity == nil || s.planTypeEntity == nil || s.paymentEntity == nil {
		return nil
	}
	// Preferred path: queue-backed fan-out via PaymentBatch + DebtBatch. Mirrors
	// TS BatchService.run(new PaymentBatch(...)) + run(new DebtBatch(...)).
	if s.queueService != nil {
		_, _ = capitalpayment.RunFromQueueApp(s.queueService, s.planEntity, period)
		_, _ = capitalpayment.RunDebtFromQueueApp(s.queueService, s.planEntity, s.paymentEntity, period)
		return nil
	}
	// Fallback: synchronous accrual insert when the queue isn't wired (tests/dev).
	plans, err := s.planEntity.Find(0, 0, `"period_type" = ?`, period)
	if err != nil {
		return err
	}
	for _, plan := range plans {
		pt, _ := s.planTypeEntity.First(`id = ?`, plan.TypeId)
		if pt == nil {
			continue
		}
		amount := pt.Monthly
		if period == "yearly" && pt.Yearly != nil {
			amount = *pt.Yearly
		}
		paid := false
		_ = s.paymentEntity.Insert(&capitalpayment.Payment{
			UserId:      plan.UserId,
			ClientId:    plan.ClientId,
			WorkspaceId: plan.WorkspaceId,
			PlanId:      plan.Id,
			Amount:      &amount,
			Currency:    &pt.Currency,
			Type:        "accrual",
			Reference:   plan.Id + "-" + period + "-" + time.Now().UTC().Format("2006-01"),
			Paid:        &paid,
			PeriodType:  &period,
		})
	}
	return nil
}
