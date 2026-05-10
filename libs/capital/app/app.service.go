package app

import (
	"time"

	capitalpayment "github.com/thescaffold/gox-apps/libs/capital/app/payment"
	capitalpaymentlog "github.com/thescaffold/gox-apps/libs/capital/app/paymentlog"
	capitalplan "github.com/thescaffold/gox-apps/libs/capital/app/plan"
	capitalplantype "github.com/thescaffold/gox-apps/libs/capital/app/plantype"
	capitalrate "github.com/thescaffold/gox-apps/libs/capital/app/rate"
	capitalusage "github.com/thescaffold/gox-apps/libs/capital/app/usage"
	capitalwallet "github.com/thescaffold/gox-apps/libs/capital/app/wallet"
)

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

// OnMonthlyHeartbeat iterates active plans and creates accrual Payment rows.
// Mirrors TS PaymentBatch which loops plans and calls paymentService.accrual().
// Inserts an unpaid "accrual" Payment row per active plan with the plan-type's
// monthly amount; downstream payment.pay events come from the actual payment
// rails.
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
	plans, err := s.planEntity.Find(0, 0, ``)
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
