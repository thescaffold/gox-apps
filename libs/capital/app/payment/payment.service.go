// Package payment hosts the gox capital payment surface — start/complete/new/
// existing/debt/accrual handlers ported deeply from
// ntx-apps/libs/capital/src/api/payment/payment.service.ts.
//
// Provider HTTP integration (Stripe/Flutterwave/Paystack) plugs in via the
// Provider interface — host code wires real providers post-DI through
// `RegisterProvider` so the service runs without HTTP in test/dev.
package payment

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	capitalpaymentlog "github.com/thescaffold/gox-apps/libs/capital/app/paymentlog"
	capitalplan "github.com/thescaffold/gox-apps/libs/capital/app/plan"
	capitalplantype "github.com/thescaffold/gox-apps/libs/capital/app/plantype"
	capitalprovider "github.com/thescaffold/gox-apps/libs/capital/app/provider"
	capitalrate "github.com/thescaffold/gox-apps/libs/capital/app/rate"
	capitalusage "github.com/thescaffold/gox-apps/libs/capital/app/usage"
	capitalwallet "github.com/thescaffold/gox-apps/libs/capital/app/wallet"
	"github.com/thescaffold/gox-packages/libs/core/events"
	"github.com/thescaffold/gox-packages/libs/core/utils"
)

// Period type strings — gox uses the lowercased values directly while TS uses
// a PeriodType enum. Keeping them here as untyped constants avoids dragging
// in the full enum surface for one comparison.
const (
	PeriodDaily   = "daily"
	PeriodWeekly  = "weekly"
	PeriodMonthly = "monthly"
	PeriodYearly  = "yearly"
)

// PaymentLog type constants mirror TS PaymentLogType.
const (
	LogInit       = "init"
	LogVerify     = "verify"
	LogCharge     = "charge"
	LogWallet     = "wallet"
	LogRecordOnly = "record_only"
	LogInvoice    = "invoice"
	LogReceipt    = "receipt"
	LogLog        = "log"
)

// Payment type constants mirror TS PaymentType.
const (
	TypeCharge     = "charge"
	TypeVerify     = "verify"
	TypeWallet     = "wallet"
	TypeRecordOnly = "record_only"
)

// PaymentStatus constants mirror TS PaymentStatusType.
const (
	StatusPending = "pending"
	StatusSuccess = "success"
)

// PaymentMetaNonRecurrentTypes mirrors the TS constant — these payment-meta
// `type` values bypass debt aggregation and aren't deduplicated by
// (plan, period, currency, type).
var PaymentMetaNonRecurrentTypes = map[string]bool{
	"apps.capital.license.pay":             true,
	"apps.capital.instance.pay":            true,
	"apps.capital.card.link":               true,
	"apps.capital.gate.link":               true,
	"apps.capital.payment.link":            true,
	"apps.capital.dashboard.upgrade-plan":  true,
	"apps.capital.dashboard.pay-debt":      true,
}

// ProviderResponse is the rich return value providers emit. TS returns a
// `[status, title, message, data]` tuple; gox uses a struct for clarity.
type ProviderResponse struct {
	Status  bool
	Title   string
	Message string
	Data    any // map[string]any or []map[string]any
}

// CardLinkData is the shape providers should emit in ProviderResponse.Data
// after a successful Verify. Multiple cards per response are supported (TS
// flattens a single map into [map]).
type CardLinkData struct {
	Type      string         `json:"type,omitempty"`
	Signature string         `json:"signature,omitempty"`
	Email     string         `json:"email,omitempty"`
	Token     string         `json:"token,omitempty"`
	Country   string         `json:"country,omitempty"`
	Meta      map[string]any `json:"meta,omitempty"`
}

// Provider is the rich surface every payment provider satisfies. Mirrors the
// TS provider contract (Init/Verify/Charge) one-for-one.
type Provider interface {
	Name() string
	// Init mirrors TS provider.init — opens a provider session for the
	// (user, client, workspace, amount, currency) tuple.
	Init(firstName, lastName, email, userID, clientID, workspaceID string, amount float64, currency string) (map[string]any, error)
	// Verify confirms a previously-initialised reference. Returns the rich
	// response so the PaymentService can drive card-link side effects.
	Verify(userID, clientID, workspaceID, reference string, amount float64, currency string) (*ProviderResponse, error)
	// Charge runs a saved-card charge against the supplied Provider row.
	Charge(provider *capitalprovider.Provider, userID, clientID, workspaceID, reference string, amount float64, currency string) (*ProviderResponse, error)
}

// LegacyProvider is the previous (simpler) Provider surface kept for
// back-compat with the Phase 7.5 ProviderResponse-less callers. New providers
// should implement Provider; legacy ones can be adapted via wrapLegacy.
type LegacyProvider interface {
	Name() string
	Initialize(amount float64, currency, firstName, lastName, ref string, meta map[string]any) (map[string]any, error)
	Verify(reference string) (bool, error)
	Charge(providerId string, amount float64, currency string, meta map[string]any) (string, bool, error)
}

// PaymentService is goose-DI-managed. Provider registrations happen post-DI
// via RegisterProvider — host code wires real Stripe/Paystack/Flutterwave
// implementations after reading env credentials.
type PaymentService struct {
	entity            *PaymentEntity                       `inject:""`
	paymentLogEntity  *capitalpaymentlog.PaymentLogEntity  `inject:""`
	planTypeEntity    *capitalplantype.PlanTypeEntity      `inject:""`
	providerEntity    *capitalprovider.ProviderEntity      `inject:""`
	usageEntity       *capitalusage.UsageEntity            `inject:""`
	rateEntity        *capitalrate.RateEntity              `inject:""`
	walletService     *capitalwallet.WalletService         `inject:""`
	tracker           *events.TrackerService               `inject:""`

	mu        sync.RWMutex
	providers map[string]Provider
}

// RegisterProvider installs (or replaces) the provider for a given name.
// Safe to call from module OnRegister hooks.
func (s *PaymentService) RegisterProvider(p Provider) {
	if p == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.providers == nil {
		s.providers = map[string]Provider{}
	}
	s.providers[p.Name()] = p
}

// RegisterLegacyProvider adapts a LegacyProvider to the rich Provider
// surface and registers it. Used by code that hasn't migrated to the
// ProviderResponse return shape.
func (s *PaymentService) RegisterLegacyProvider(p LegacyProvider) {
	if p == nil {
		return
	}
	s.RegisterProvider(&legacyAdapter{p: p})
}

// providerFor returns the registered provider matching name. nil when missing.
func (s *PaymentService) providerFor(name string) Provider {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.providers == nil {
		return nil
	}
	return s.providers[name]
}

// Entity exposes the entity for callers needing low-level access (the
// AppService.OnPaymentPay subscription is one such caller).
func (s *PaymentService) Entity() *PaymentEntity { return s.entity }

// ---------- Public surface ----------

// Debt mirrors TS PaymentService.debt — aggregates unpaid recurrent Payments
// for the plan's owner triple and returns the open window + line items.
// Non-recurrent payment types (card/gate/payment links, dashboard upgrades)
// are excluded from the amount total per the TS skip list.
func (s *PaymentService) Debt(plan *capitalplan.Plan) (*DebtSummary, error) {
	if plan == nil {
		return &DebtSummary{}, nil
	}
	payments, err := s.entity.Find(0, 0,
		`"user_id" = ? AND "client_id" = ? AND "workspace_id" = ? AND ("paid" IS NULL OR "paid" = ?) ORDER BY "created_at" ASC`,
		plan.UserId, plan.ClientId, plan.WorkspaceId, false,
	)
	if err != nil {
		return nil, err
	}
	out := &DebtSummary{}
	if len(payments) == 0 {
		return out, nil
	}

	// Collect PaymentLog `Log`-typed rows for these payments (TS-equivalent
	// of the unioned `paymentLogRepository.find` lookup keyed by paymentId).
	ids := make([]string, 0, len(payments))
	for i := range payments {
		ids = append(ids, payments[i].Id)
	}
	logRows, _ := s.findLogsByPaymentIds(ids, LogLog)
	for _, l := range logRows {
		var req map[string]any
		_ = json.Unmarshal(l.Request, &req)
		out.Logs = append(out.Logs, req)
	}

	var owed []*Payment
	for i := range payments {
		p := payments[i]
		if p.Type != "" && PaymentMetaNonRecurrentTypes[p.Type] {
			continue
		}
		if p.Amount != nil {
			out.Amount += float64(*p.Amount)
		}
		owed = append(owed, &p)
	}
	if oldest := payments[0]; oldest.CreatedAt != nil {
		out.From = *oldest.CreatedAt
	}
	if newest := payments[len(payments)-1]; newest.CreatedAt != nil {
		out.To = *newest.CreatedAt
	}
	out.Payments = owed
	return out, nil
}

// Accrual mirrors TS PaymentService.accrual — sums in-progress + recently
// stopped Usage rows against their Rate.perUnit and the current plan period
// window. Computes hour-prorated cost: rate.perUnit * quantity * (hours/730.5).
func (s *PaymentService) Accrual(plan *capitalplan.Plan) (*AccrualSummary, error) {
	if plan == nil {
		return &AccrualSummary{}, nil
	}
	periodType := PeriodMonthly
	if plan.PeriodType != nil && *plan.PeriodType != "" {
		periodType = *plan.PeriodType
	}
	from, to, ok := getPeriod(0, periodType)
	if !ok {
		return &AccrualSummary{}, nil
	}
	out := &AccrualSummary{From: from, To: to}

	// Two unioned filters per TS: stopAt IS NULL (still running) OR stopAt >= from.
	usages, err := s.usageEntity.Find(0, 0,
		`"user_id" = ? AND "client_id" = ? AND "workspace_id" = ? AND "start_at" IS NOT NULL AND ("stop_at" IS NULL OR "stop_at" >= ?)`,
		plan.UserId, plan.ClientId, plan.WorkspaceId, from,
	)
	if err != nil {
		return nil, err
	}
	const totalHours = 730.5 // (365.25/12) * 24
	for i := range usages {
		u := usages[i]
		rate, _ := s.rateEntity.First(`"id" = ?`, u.RateId)
		if rate == nil || rate.PerUnit == nil {
			continue
		}
		startAt := from
		if u.StartAt != nil {
			startAt = *u.StartAt
		}
		stopAt := to
		if u.StopAt != nil {
			stopAt = *u.StopAt
		}
		effectiveStart := startAt
		if startAt.Before(from) {
			effectiveStart = from
		}
		effectiveStop := stopAt
		if stopAt.Before(to) {
			effectiveStop = stopAt
		}
		hours := effectiveStop.Sub(effectiveStart).Hours()
		if hours <= 0 {
			continue
		}
		quantity := float64(u.Quantity)
		amount := float64(*rate.PerUnit) * quantity * (hours / totalHours)
		out.Amount += amount
		out.Logs = append(out.Logs, AccrualLog{
			EntityName: u.EntityName, EntityId: u.EntityId,
			Rate: *rate.PerUnit, Quantity: u.Quantity,
			Hours: hours, Amount: amount,
			StartAt: effectiveStart, StopAt: effectiveStop,
		})
	}
	return out, nil
}

// Start mirrors TS PaymentService.start — initiates a payment session by
// finding-or-creating the Payment row (period-aware), kicking off the
// provider's init flow, persisting the PaymentLog, and patching the payment
// reference with the provider-supplied reference. Returns the provider's
// response map for the caller envelope.
func (s *PaymentService) Start(plan *capitalplan.Plan, amount float64, providerName, firstName, lastName, email string, meta map[string]any) (map[string]any, error) {
	if plan == nil {
		return nil, nil
	}
	if err := s.checkMetaForType(meta); err != nil {
		return nil, err
	}
	pay, err := s.getPayment(plan, amount, meta)
	if err != nil || pay == nil {
		return nil, err
	}

	resp, err := s.initLeg(plan, pay, providerName, firstName, lastName, email)
	if err != nil {
		return nil, err
	}

	// TS pulls `reference` from the provider response and updates Payment.
	if ref := stringFromMap(resp, "reference"); ref != "" {
		pay.Reference = ref
		_, _ = s.entity.Update(pay, `"id" = ?`, pay.Id)
		resp["reference"] = ref
	} else {
		resp["reference"] = pay.Reference
	}
	return resp, nil
}

// Complete mirrors TS PaymentService.complete — looks up the Payment by the
// provider's returned reference, then runs the `Verify` leg via process().
func (s *PaymentService) Complete(plan *capitalplan.Plan, providerName, reference string, context map[string]any) (bool, error) {
	if reference == "" {
		return false, errors.New("app.capital.service.payment.invalid-reference")
	}
	pay, _ := s.entity.First(`"reference" = ?`, reference)
	if pay == nil {
		return false, errors.New("app.capital.service.payment.invalid-reference")
	}
	paid, _, err := s.process(pay, []string{TypeVerify}, context, plan, providerName, "")
	return paid, err
}

// New mirrors TS PaymentService.new — getPayment then process() across the
// supplied PaymentType list.
func (s *PaymentService) New(types []string, plan *capitalplan.Plan, amount float64, meta map[string]any, providerName, providerId string, context map[string]any) (bool, string, error) {
	if plan == nil || len(types) == 0 {
		return false, "", nil
	}
	if err := s.checkMetaForType(meta); err != nil {
		return false, "", err
	}
	pay, err := s.getPayment(plan, amount, meta)
	if err != nil || pay == nil {
		return false, "", err
	}
	return s.process(pay, types, context, plan, providerName, providerId)
}

// Existing mirrors TS PaymentService.existing — iterates unpaid recurrent
// Payments for the plan owner and runs process() for each. Returns true only
// when every attempt succeeded.
func (s *PaymentService) Existing(types []string, plan *capitalplan.Plan, providerName, providerId string, context map[string]any) (bool, error) {
	if plan == nil || len(types) == 0 {
		return false, nil
	}
	payments, err := s.entity.Find(0, 0,
		`"user_id" = ? AND "client_id" = ? AND "workspace_id" = ? AND ("paid" IS NULL OR "paid" = ?) ORDER BY "created_at" ASC`,
		plan.UserId, plan.ClientId, plan.WorkspaceId, false,
	)
	if err != nil {
		return false, err
	}
	if len(payments) == 0 {
		return false, nil
	}
	allPaid := true
	any := false
	for i := range payments {
		p := payments[i]
		if PaymentMetaNonRecurrentTypes[p.Type] {
			continue
		}
		any = true
		paid, _, err := s.process(&p, types, context, plan, providerName, providerId)
		if err != nil {
			return false, err
		}
		if !paid {
			allPaid = false
		}
	}
	return any && allPaid, nil
}

// ---------- Process orchestration ----------

// process is the unified flow used by Start (verify-only after init), Complete
// (verify), New (caller-chosen types), and Existing (retry across types).
// Pulls open PaymentLog rows for invoice context, attempts each PaymentType
// in order, then issues invoice/receipt + tracker events.
func (s *PaymentService) process(pay *Payment, types []string, context map[string]any, plan *capitalplan.Plan, providerName, providerId string) (bool, string, error) {
	currency := ""
	if pay.Currency != nil {
		currency = *pay.Currency
	}
	amount := 0.0
	if pay.Amount != nil {
		amount = float64(*pay.Amount)
	}
	var from, to time.Time
	if pay.StartAt != nil {
		from = *pay.StartAt
	}
	if pay.EndAt != nil {
		to = *pay.EndAt
	}

	// Issue invoice when we have request context + no existing invoice URL.
	if context != nil && (pay.InvoiceUrl == nil || *pay.InvoiceUrl == "") {
		// gather LOG-typed logs to embed in the invoice.
		logRows, _ := s.findLogsByPaymentIds([]string{pay.Id}, LogLog)
		logs := make([]map[string]any, 0, len(logRows))
		for _, l := range logRows {
			var m map[string]any
			_ = json.Unmarshal(l.Request, &m)
			logs = append(logs, m)
		}
		url := s.invoice(plan, pay, from, to, amount, currency, pay.Reference, logs, context)
		pay.InvoiceUrl = &url
	}

	paid := pay.Paid != nil && *pay.Paid
	var usedType string
	if !paid {
		for _, t := range types {
			ok := false
			switch t {
			case TypeCharge:
				ok = s.chargeLeg(plan, pay, from, to, amount, currency, providerId)
			case TypeVerify:
				ok = s.verifyLeg(plan, pay, from, to, amount, currency, providerName)
			case TypeWallet:
				ok = s.walletLeg(plan, pay, from, to, amount, currency)
			case TypeRecordOnly:
				ok = s.recordOnlyLeg(plan, pay, from, to, amount, currency)
			}
			if ok {
				paid = true
				usedType = t
				break
			}
		}
	}

	// Receipt when paid + context available.
	if paid && context != nil && (pay.ReceiptUrl == nil || *pay.ReceiptUrl == "") {
		url := s.receipt(plan, pay, from, to, amount, currency, pay.Reference, context)
		pay.ReceiptUrl = &url
	}

	// Persist the payment row's terminal state.
	status := StatusPending
	if paid {
		status = StatusSuccess
	}
	pay.Status = &status
	pay.Paid = &paid
	_, _ = s.entity.Update(pay, `"id" = ?`, pay.Id)

	// Tracker dispatch.
	if s.tracker != nil {
		if paid {
			s.tracker.Message("apps.capital.payment.pay", map[string]any{
				"status":  true,
				"type":    usedType,
				"plan":    plan,
				"payment": pay,
			})
		} else {
			s.tracker.Message("apps.notification.message.new", map[string]any{
				"reference":   pay.Reference,
				"userId":      plan.UserId,
				"clientId":    plan.ClientId,
				"workspaceId": plan.WorkspaceId,
				"key":         "failed-payment",
				"channels":    []string{"email"},
				"data": map[string]any{
					"plan": plan, "amount": amount, "reference": pay.Reference,
					"payment": pay, "type": strings.Join(types, ", "),
				},
				"subject":  "Payment failed - Verify your billing information",
				"type":     "system",
				"priority": "medium",
				"theme":    "info",
				"scope":    "workspace",
			})
		}
	}
	return paid, pay.Reference, nil
}

// ---------- Per-type legs ----------

func (s *PaymentService) initLeg(plan *capitalplan.Plan, pay *Payment, providerName, firstName, lastName, email string) (map[string]any, error) {
	currency := ""
	if pay.Currency != nil {
		currency = *pay.Currency
	}
	amount := 0.0
	if pay.Amount != nil {
		amount = float64(*pay.Amount)
	}
	p := s.providerFor(providerName)
	var data map[string]any
	if p != nil {
		out, err := p.Init(firstName, lastName, email, plan.UserId, plan.ClientId, plan.WorkspaceId, amount, currency)
		if err != nil {
			return nil, err
		}
		data = out
	}
	if data == nil {
		// no provider wired — return a minimal session shape so callers don't
		// have to special-case the dev/test path.
		data = map[string]any{
			"reference": pay.Reference,
			"amount":    amount,
			"currency":  currency,
		}
	}
	s.logRequestResponse(pay.Id, LogInit, map[string]any{
		"plan": plan, "amount": amount, "currency": currency,
		"paymentId": pay.Id, "providerName": providerName,
	}, data)
	return data, nil
}

func (s *PaymentService) verifyLeg(plan *capitalplan.Plan, pay *Payment, _, _ time.Time, amount float64, currency, providerName string) bool {
	p := s.providerFor(providerName)
	if p == nil {
		s.logRequestResponse(pay.Id, LogVerify, map[string]any{
			"plan": plan, "amount": amount, "currency": currency,
			"paymentId": pay.Id, "providerName": providerName,
		}, nil)
		return false
	}
	resp, err := p.Verify(plan.UserId, plan.ClientId, plan.WorkspaceId, pay.Reference, amount, currency)
	if err != nil || resp == nil {
		s.logRequestResponse(pay.Id, LogVerify, map[string]any{
			"plan": plan, "amount": amount, "currency": currency,
			"paymentId": pay.Id, "providerName": providerName,
		}, errorMap(err))
		return false
	}
	s.logRequestResponse(pay.Id, LogVerify, map[string]any{
		"plan": plan, "amount": amount, "currency": currency,
		"paymentId": pay.Id, "providerName": providerName,
	}, resp.Data)
	if !resp.Status {
		return false
	}

	// Card-link side effects: persist Provider rows for each card the verify
	// response surfaced. Multiple cards per response are supported (TS
	// normalises single-map → [single-map]).
	cards := s.cardsFromResponse(resp)
	for _, card := range cards {
		s.linkCard(plan, providerName, currency, card)
	}

	// Fund the wallet with the verified amount.
	if s.walletService != nil {
		_, _ = s.walletService.Fund(plan.UserId, pay.Reference, amount, currency, map[string]any{
			"narration": "chisq|card|charge|new",
		})
	}
	return true
}

func (s *PaymentService) chargeLeg(plan *capitalplan.Plan, pay *Payment, _, _ time.Time, amount float64, currency, providerId string) bool {
	providers, err := s.providersForCharge(plan, providerId)
	if err != nil || len(providers) == 0 {
		return false
	}
	paid := false
	for _, prov := range providers {
		p := s.providerFor(prov.Name)
		if p == nil {
			continue
		}
		row := prov
		resp, err := p.Charge(&row, plan.UserId, plan.ClientId, plan.WorkspaceId, pay.Reference, amount, prov.Currency)
		var data any
		if resp != nil {
			data = resp.Data
		}
		s.logRequestResponse(pay.Id, LogCharge, map[string]any{
			"plan": plan, "amount": amount, "currency": currency,
			"paymentId": pay.Id, "providerName": prov.Name,
		}, data)
		if err == nil && resp != nil && resp.Status {
			paid = true
			break
		}
	}
	if !paid {
		return false
	}
	if s.walletService != nil {
		_, _ = s.walletService.Fund(plan.UserId, pay.Reference, amount, currency, map[string]any{
			"narration": "chisq|card|charge",
		})
	}
	return true
}

func (s *PaymentService) walletLeg(plan *capitalplan.Plan, pay *Payment, _, _ time.Time, amount float64, currency string) bool {
	s.logRequestResponse(pay.Id, LogWallet, map[string]any{
		"plan": plan, "amount": amount, "currency": currency,
		"paymentId": pay.Id,
	}, nil)
	if s.walletService == nil {
		return true
	}
	// Mirrors TS walletService.withdraw.default(...) — debits the user's
	// wallet; on insufficient funds the leg returns false so the caller
	// falls through to the next PaymentType (TS-equivalent error path).
	if _, err := s.walletService.Withdraw(plan.UserId, pay.Reference, amount, currency, map[string]any{
		"narration": "chisq|wallet|charge",
	}); err != nil {
		return false
	}
	return true
}

func (s *PaymentService) recordOnlyLeg(plan *capitalplan.Plan, pay *Payment, _, _ time.Time, amount float64, currency string) bool {
	s.logRequestResponse(pay.Id, LogRecordOnly, map[string]any{
		"plan": plan, "amount": amount, "currency": currency,
		"paymentId": pay.Id,
	}, nil)
	if s.walletService != nil {
		_, _ = s.walletService.Fund(plan.UserId, pay.Reference, amount, currency, map[string]any{
			"narration": "chisq|card|charge",
		})
	}
	return true
}

// ---------- Invoice + Receipt dispatch ----------

func (s *PaymentService) invoice(plan *capitalplan.Plan, pay *Payment, from, to time.Time, amount float64, currency, ref string, logs []map[string]any, context map[string]any) string {
	now := time.Now().UTC()
	name := fmt.Sprintf("INV-%s-%s-%s", from.Format("20060102"), to.Format("20060102"), utils.Reference("", 12))
	narration := fmt.Sprintf("Invoice for %s to %s", from.Format("Jan 2 2006"), to.Format("Jan 2 2006"))
	url := figsURL(name)

	if s.tracker != nil {
		s.tracker.Message("apps.figs.message.new", map[string]any{
			"input": map[string]any{
				"type": "html",
				"data": map[string]any{
					"plan": plan, "from": from, "to": to, "amount": amount, "logs": logs,
					"tax": 0.0, "total": amount, "reference": ref, "narration": narration,
					"context": context,
				},
				"templateUrl": fmt.Sprintf("pdf/en/%s/invoice.html", plan.ClientId),
			},
			"output": map[string]any{"type": "pdf"},
			"meta":   map[string]any{"name": name, "store": "local"},
		})
	}
	s.logRequestResponse(pay.Id, LogInvoice, map[string]any{
		"plan": plan, "from": from, "to": to, "amount": amount, "reference": ref,
	}, map[string]any{"url": url})

	if s.tracker != nil {
		expire := now.Add(7 * 24 * time.Hour)
		s.tracker.Message("apps.notification.message.new", map[string]any{
			"reference": ref, "userId": plan.UserId, "clientId": plan.ClientId, "workspaceId": plan.WorkspaceId,
			"key": "invoice", "channels": []string{"email"},
			"data": map[string]any{
				"plan": plan, "from": from, "to": to, "amount": amount, "logs": logs,
				"tax": 0.0, "total": amount, "reference": ref, "invoiceUrl": url, "narration": narration,
			},
			"subject": "Payment Invoice", "type": "system", "priority": "medium",
			"theme": "info", "scope": "workspace", "position": "bottom right",
			"publishAt": now, "expireAt": expire,
		})
	}
	return url
}

func (s *PaymentService) receipt(plan *capitalplan.Plan, pay *Payment, from, to time.Time, amount float64, currency, ref string, context map[string]any) string {
	now := time.Now().UTC()
	name := fmt.Sprintf("RCP-%s-%s-%s", from.Format("20060102"), to.Format("20060102"), utils.Reference("", 12))
	narration := fmt.Sprintf("Payment for %s to %s", from.Format("Jan 2 2006"), to.Format("Jan 2 2006"))
	url := figsURL(name)

	if s.tracker != nil {
		s.tracker.Message("apps.figs.message.new", map[string]any{
			"input": map[string]any{
				"type": "html",
				"data": map[string]any{
					"plan": plan, "from": from, "to": to, "amount": amount,
					"reference": ref, "time": now, "narration": narration, "context": context,
				},
				"templateUrl": fmt.Sprintf("pdf/en/%s/receipt.html", plan.ClientId),
			},
			"output": map[string]any{"type": "pdf"},
			"meta":   map[string]any{"name": name, "store": "local"},
		})
	}
	s.logRequestResponse(pay.Id, LogReceipt, map[string]any{
		"plan": plan, "from": from, "to": to, "amount": amount,
		"currency": currency, "paymentId": pay.Id, "reference": ref,
	}, map[string]any{"url": url})

	if s.tracker != nil {
		expire := now.Add(7 * 24 * time.Hour)
		s.tracker.Message("apps.notification.message.new", map[string]any{
			"reference": ref, "userId": plan.UserId, "clientId": plan.ClientId, "workspaceId": plan.WorkspaceId,
			"key": "receipt", "channels": []string{"email"},
			"data": map[string]any{
				"plan": plan, "from": from, "to": to, "amount": amount,
				"reference": ref, "time": now, "receiptUrl": url, "narration": narration,
			},
			"subject": "Payment Receipt", "type": "system", "priority": "medium",
			"theme": "info", "scope": "workspace", "position": "bottom right",
			"publishAt": now, "expireAt": expire,
		})
	}
	return url
}

// ---------- Internal helpers ----------

// getPayment mirrors TS PaymentService.getPayment — for recurrent payment-meta
// types, it updateOrCreates by (userId, clientId, workspaceId, planId, periodType,
// period, currency, type); for non-recurrent types it inserts a fresh row.
// Aborts with an error when the period's existing row already paid.
func (s *PaymentService) getPayment(plan *capitalplan.Plan, amount float64, meta map[string]any) (*Payment, error) {
	currency, err := s.getCurrency(plan)
	if err != nil {
		return nil, err
	}
	metaType, _ := meta["type"].(string)
	nonRecurrent := PaymentMetaNonRecurrentTypes[metaType]
	periodType := PeriodMonthly
	if plan.PeriodType != nil && *plan.PeriodType != "" {
		periodType = *plan.PeriodType
	}

	var existingCount int64
	if !nonRecurrent {
		existingCount, _ = s.entity.Count(
			`"user_id" = ? AND "client_id" = ? AND "workspace_id" = ? AND "plan_id" = ? AND "period_type" = ? AND "currency" = ? AND "type" = ?`,
			plan.UserId, plan.ClientId, plan.WorkspaceId, plan.Id, periodType, currency, metaType,
		)
	}
	offset := 0
	if !nonRecurrent {
		if existingCount > 0 {
			offset = -1
		} else {
			offset = -2
		}
	}
	from, to, ok := getPeriod(offset, periodType)
	if !ok {
		return nil, errors.New("app.capital.payment.service.error.invalid-payment-type")
	}
	period := from.Format(periodFormat(periodType))
	amt := int(amount)
	ref := utils.Reference("TEM", 36)
	metaJSON, _ := json.Marshal(meta)
	status := StatusPending

	if nonRecurrent {
		row := &Payment{
			UserId: plan.UserId, ClientId: plan.ClientId, WorkspaceId: plan.WorkspaceId,
			PlanId: plan.Id, PeriodType: &periodType, Period: &period,
			Currency: &currency, Type: metaType, Reference: ref, Amount: &amt,
			StartAt: &from, EndAt: &to, Status: &status, Meta: metaJSON,
		}
		if err := s.entity.Insert(row); err != nil {
			return nil, err
		}
		return row, nil
	}

	// Recurrent: updateOrCreate by the (plan, period, currency, type) tuple.
	existing, _ := s.entity.First(
		`"user_id" = ? AND "client_id" = ? AND "workspace_id" = ? AND "plan_id" = ? AND "period_type" = ? AND "period" = ? AND "currency" = ? AND "type" = ?`,
		plan.UserId, plan.ClientId, plan.WorkspaceId, plan.Id, periodType, period, currency, metaType,
	)
	if existing != nil {
		if existing.Status != nil && *existing.Status == StatusSuccess {
			return nil, errors.New("app.capital.service.payment.already-paid")
		}
		existing.Reference = ref
		existing.Amount = &amt
		existing.StartAt = &from
		existing.EndAt = &to
		existing.Status = &status
		existing.Meta = metaJSON
		if _, err := s.entity.Update(existing, `"id" = ?`, existing.Id); err != nil {
			return nil, err
		}
		return existing, nil
	}
	row := &Payment{
		UserId: plan.UserId, ClientId: plan.ClientId, WorkspaceId: plan.WorkspaceId,
		PlanId: plan.Id, PeriodType: &periodType, Period: &period,
		Currency: &currency, Type: metaType, Reference: ref, Amount: &amt,
		StartAt: &from, EndAt: &to, Status: &status, Meta: metaJSON,
	}
	if err := s.entity.Insert(row); err != nil {
		return nil, err
	}
	return row, nil
}

func (s *PaymentService) getCurrency(plan *capitalplan.Plan) (string, error) {
	if plan == nil {
		return "", errors.New("app.capital.payment.service.error.currency-not-found")
	}
	pt, _ := s.planTypeEntity.First(`"id" = ?`, plan.TypeId)
	if pt != nil && pt.Currency != "" {
		return pt.Currency, nil
	}
	return "", errors.New("app.capital.payment.service.error.currency-not-found")
}

func (s *PaymentService) checkMetaForType(meta map[string]any) error {
	if meta == nil {
		return errors.New("app.capital.payment.service.error.invalid-payment-meta-type")
	}
	t, _ := meta["type"].(string)
	if t == "" || len(t) == 1 {
		return errors.New("app.capital.payment.service.error.invalid-payment-meta-type")
	}
	return nil
}

func (s *PaymentService) findLogsByPaymentIds(ids []string, logType string) ([]capitalpaymentlog.PaymentLog, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	// Build an IN clause; gox-go ent doesn't have one-shot IN, so concatenate.
	placeholders := make([]string, len(ids))
	args := make([]any, 0, len(ids)+1)
	for i, id := range ids {
		placeholders[i] = "?"
		args = append(args, id)
	}
	args = append(args, logType)
	where := `"payment_id" IN (` + strings.Join(placeholders, ",") + `) AND "type" = ? ORDER BY "created_at" ASC`
	return s.paymentLogEntity.Find(0, 0, where, args...)
}

func (s *PaymentService) logRequestResponse(paymentId, logType string, request, response any) {
	if s.paymentLogEntity == nil {
		return
	}
	reqJSON, _ := json.Marshal(request)
	respJSON, _ := json.Marshal(response)
	_ = s.paymentLogEntity.Insert(&capitalpaymentlog.PaymentLog{
		PaymentId: paymentId, Type: logType,
		Request: reqJSON, Response: respJSON,
	})
}

func (s *PaymentService) cardsFromResponse(resp *ProviderResponse) []CardLinkData {
	if resp == nil || resp.Data == nil {
		return nil
	}
	switch d := resp.Data.(type) {
	case []CardLinkData:
		return d
	case CardLinkData:
		return []CardLinkData{d}
	case map[string]any:
		return []CardLinkData{cardFromMap(d)}
	case []map[string]any:
		out := make([]CardLinkData, 0, len(d))
		for _, m := range d {
			out = append(out, cardFromMap(m))
		}
		return out
	case []any:
		out := make([]CardLinkData, 0, len(d))
		for _, v := range d {
			if m, ok := v.(map[string]any); ok {
				out = append(out, cardFromMap(m))
			}
		}
		return out
	}
	return nil
}

func (s *PaymentService) linkCard(plan *capitalplan.Plan, providerName, currency string, card CardLinkData) {
	if s.providerEntity == nil || s.walletService == nil {
		return
	}
	acct, err := s.walletService.Init(plan.UserId, plan.ClientId, plan.WorkspaceId, currency, "default", "default")
	if err != nil || acct == nil {
		return
	}
	existing, _ := s.providerEntity.First(
		`"account_id" = ? AND "name" = ? AND "type" = ? AND "signature" = ? AND "email" = ?`,
		acct.Id, providerName, "card", card.Signature, card.Email,
	)
	if existing != nil {
		return
	}
	primary := true
	accId := acct.Id
	email := card.Email
	tok := card.Token
	sig := card.Signature
	metaJSON, _ := json.Marshal(card.Meta)
	prov := &capitalprovider.Provider{
		AccountId: &accId, Name: providerName, Type: "card",
		Signature: &sig, Email: &email, Token: &tok,
		Primary: &primary, Currency: currency, Country: card.Country,
		Meta: metaJSON,
	}
	if err := s.providerEntity.Insert(prov); err != nil {
		return
	}
	// Reset primary on sibling rows.
	notPrimary := false
	_, _ = s.providerEntity.Update(
		&capitalprovider.Provider{Primary: &notPrimary},
		`"account_id" = ? AND "id" <> ?`, acct.Id, prov.Id,
	)
}

func (s *PaymentService) providersForCharge(plan *capitalplan.Plan, providerId string) ([]capitalprovider.Provider, error) {
	if s.providerEntity == nil {
		return nil, nil
	}
	if providerId != "" {
		row, _ := s.providerEntity.First(`"id" = ?`, providerId)
		if row == nil {
			return nil, nil
		}
		return []capitalprovider.Provider{*row}, nil
	}
	// Resolve via walletService.Detail → accountId → providers.
	if s.walletService == nil {
		return nil, nil
	}
	acct, err := s.walletService.Detail(plan.UserId)
	if err != nil || acct == nil {
		return nil, err
	}
	rows, err := s.providerEntity.Find(0, 0, `"account_id" = ?`, acct.Id)
	if err != nil {
		return nil, err
	}
	// primary-first sort, matching TS.
	sortProvidersPrimaryFirst(rows)
	return rows, nil
}

// ---------- DebtSummary / AccrualSummary shapes ----------

// DebtSummary is the response shape of Debt and part of Status.
type DebtSummary struct {
	Amount   float64          `json:"amount"`
	From     time.Time        `json:"from"`
	To       time.Time        `json:"to"`
	Logs     []map[string]any `json:"logs"`
	Payments []*Payment       `json:"payments"`
}

// AccrualSummary is the response shape of Accrual.
type AccrualSummary struct {
	Amount float64      `json:"amount"`
	From   time.Time    `json:"from"`
	To     time.Time    `json:"to"`
	Logs   []AccrualLog `json:"logs"`
}

// AccrualLog is one line item of AccrualSummary.Logs.
type AccrualLog struct {
	EntityName string    `json:"entityName"`
	EntityId   string    `json:"entityId"`
	Rate       int       `json:"rate"`
	Quantity   int       `json:"quantity"`
	Hours      float64   `json:"hours"`
	Amount     float64   `json:"amount"`
	StartAt    time.Time `json:"startAt"`
	StopAt     time.Time `json:"stopAt"`
}

// ---------- Package-level helpers ----------

// getPeriod mirrors TS getPeriod — given an offset (0=now, -1=previous,
// -2=next) and a period type, returns the [from, to] window.
func getPeriod(offset int, periodType string) (time.Time, time.Time, bool) {
	now := time.Now().UTC()
	var anchor time.Time
	switch offset {
	case 0:
		anchor = now
	case -1:
		anchor = subtractPeriod(now, periodType)
	case -2:
		anchor = addPeriod(now, periodType)
	default:
		return time.Time{}, time.Time{}, false
	}
	from := startOfPeriod(anchor, periodType)
	to := endOfPeriod(anchor, periodType)
	return from, to, true
}

func subtractPeriod(t time.Time, periodType string) time.Time {
	switch periodType {
	case PeriodDaily:
		return t.AddDate(0, 0, -1)
	case PeriodWeekly:
		return t.AddDate(0, 0, -7)
	case PeriodMonthly:
		return t.AddDate(0, -1, 0)
	case PeriodYearly:
		return t.AddDate(-1, 0, 0)
	}
	return t
}

func addPeriod(t time.Time, periodType string) time.Time {
	switch periodType {
	case PeriodDaily:
		return t.AddDate(0, 0, 1)
	case PeriodWeekly:
		return t.AddDate(0, 0, 7)
	case PeriodMonthly:
		return t.AddDate(0, 1, 0)
	case PeriodYearly:
		return t.AddDate(1, 0, 0)
	}
	return t
}

func startOfPeriod(t time.Time, periodType string) time.Time {
	switch periodType {
	case PeriodDaily:
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	case PeriodWeekly:
		// ISO week: Monday start.
		offset := int(t.Weekday()) - 1
		if offset < 0 {
			offset = 6
		}
		base := t.AddDate(0, 0, -offset)
		return time.Date(base.Year(), base.Month(), base.Day(), 0, 0, 0, 0, time.UTC)
	case PeriodMonthly:
		return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	case PeriodYearly:
		return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
	}
	return t
}

func endOfPeriod(t time.Time, periodType string) time.Time {
	switch periodType {
	case PeriodDaily:
		return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999_000_000, time.UTC)
	case PeriodWeekly:
		start := startOfPeriod(t, PeriodWeekly)
		return start.AddDate(0, 0, 7).Add(-time.Nanosecond)
	case PeriodMonthly:
		return startOfPeriod(t, PeriodMonthly).AddDate(0, 1, 0).Add(-time.Nanosecond)
	case PeriodYearly:
		return startOfPeriod(t, PeriodYearly).AddDate(1, 0, 0).Add(-time.Nanosecond)
	}
	return t
}

func periodFormat(periodType string) string {
	switch periodType {
	case PeriodDaily:
		return "02-01-2006"
	case PeriodWeekly:
		return "01-2006" // gox approximation; TS uses ISO WW-YYYY.
	case PeriodMonthly:
		return "01-2006"
	case PeriodYearly:
		return "2006"
	}
	return "01-2006"
}

func sortProvidersPrimaryFirst(rows []capitalprovider.Provider) {
	// Stable sort: primary=true first. Avoids importing sort by doing a
	// simple in-place shuffle (lists here are O(low single digits) in
	// practice — one card per user is the common case).
	for i := 0; i < len(rows); i++ {
		if rows[i].Primary != nil && *rows[i].Primary {
			rows[0], rows[i] = rows[i], rows[0]
			break
		}
	}
}

func cardFromMap(m map[string]any) CardLinkData {
	c := CardLinkData{}
	c.Type, _ = m["type"].(string)
	c.Signature, _ = m["signature"].(string)
	c.Email, _ = m["email"].(string)
	c.Token, _ = m["token"].(string)
	c.Country, _ = m["country"].(string)
	if v, ok := m["meta"].(map[string]any); ok {
		c.Meta = v
	}
	return c
}

func stringFromMap(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func errorMap(err error) map[string]any {
	if err == nil {
		return nil
	}
	return map[string]any{"error": err.Error()}
}

// figsURL builds the public URL for a generated figs document. Matches TS
// `${BLOBS_BASE_URL}/${name}`; empty BLOBS_BASE_URL falls back to a relative
// path so tests don't depend on env.
func figsURL(name string) string {
	base := envFigsBase()
	if base == "" {
		return "/figs/" + name
	}
	return base + "/" + name
}

func envFigsBase() string { return os.Getenv("BLOBS_BASE_URL") }

// ---------- LegacyProvider adapter ----------

type legacyAdapter struct{ p LegacyProvider }

func (a *legacyAdapter) Name() string { return a.p.Name() }
func (a *legacyAdapter) Init(firstName, lastName, email, userID, clientID, workspaceID string, amount float64, currency string) (map[string]any, error) {
	return a.p.Initialize(amount, currency, firstName, lastName, "", map[string]any{
		"userId": userID, "clientId": clientID, "workspaceId": workspaceID, "email": email,
	})
}
func (a *legacyAdapter) Verify(userID, clientID, workspaceID, reference string, amount float64, currency string) (*ProviderResponse, error) {
	ok, err := a.p.Verify(reference)
	if err != nil {
		return nil, err
	}
	return &ProviderResponse{Status: ok}, nil
}
func (a *legacyAdapter) Charge(provider *capitalprovider.Provider, userID, clientID, workspaceID, reference string, amount float64, currency string) (*ProviderResponse, error) {
	providerId := ""
	if provider != nil {
		providerId = provider.Id
	}
	_, ok, err := a.p.Charge(providerId, amount, currency, nil)
	if err != nil {
		return nil, err
	}
	return &ProviderResponse{Status: ok}, nil
}
