package app

import (
	"encoding/json"
	"time"

	goqueues "github.com/awesome-goose/goose/modules/queues"
	bridgelicense "github.com/thescaffold/gox-apps/libs/bridge/app/license"
	bridgelicensetype "github.com/thescaffold/gox-apps/libs/bridge/app/licensetype"
	bridgewebhook "github.com/thescaffold/gox-apps/libs/bridge/app/webhook"
	capitalusage "github.com/thescaffold/gox-apps/libs/capital/app/usage"
	"github.com/thescaffold/gox-packages/libs/core/events"
	"github.com/thescaffold/gox-packages/libs/core/utils"
)

// AppService is the public surface for bridge event handlers.
// Mirrors ntx-apps/libs/bridge/src/app.service.ts + the subscription bodies
// in app.controller.ts.
type AppService struct {
	licenseEntity     *bridgelicense.LicenseEntity         `inject:""`
	licenseTypeEntity *bridgelicensetype.LicenseTypeEntity `inject:""`
	webhookEntity     *bridgewebhook.WebhookEntity         `inject:""`
	usageService      *capitalusage.UsageService           `inject:""`
	tracker           *events.TrackerService               `inject:""`
}

func (s *AppService) GetHello() string { return "Hello World!" }

// LicensePayload is the apps.bridge.license.register event payload.
type LicensePayload struct {
	UserID      string         `json:"userId,omitempty"`
	ClientID    string         `json:"clientId,omitempty"`
	WorkspaceID string         `json:"workspaceId,omitempty"`
	TypeKey     string         `json:"typeKey,omitempty"`
	TypeName    string         `json:"typeName,omitempty"`
	Currency    string         `json:"currency,omitempty"`
	Monthly     float64        `json:"monthly,omitempty"`
	PeriodType  string         `json:"periodType,omitempty"`
	Meta        map[string]any `json:"meta,omitempty"`
}

// OnLicenseRegister upserts a LicenseType row, then inserts a License for the
// supplied owner (if any). Mirrors TS updateOrCreate(LicenseType) then create(License).
func (s *AppService) OnLicenseRegister(p LicensePayload) error {
	if s.licenseTypeEntity == nil || s.licenseEntity == nil || p.TypeKey == "" {
		return nil
	}
	currency := p.Currency
	if currency == "" {
		currency = "USD"
	}
	lt, _ := s.licenseTypeEntity.First(`"key" = ?`, p.TypeKey)
	if lt == nil {
		name := p.TypeName
		if name == "" {
			name = p.TypeKey
		}
		lt = &bridgelicensetype.LicenseType{
			Key: p.TypeKey, Name: name, Currency: currency, Monthly: p.Monthly,
		}
		if err := s.licenseTypeEntity.Insert(lt); err != nil {
			return err
		}
	}
	if p.UserID == "" {
		return nil
	}
	// One license per (userId, workspaceId, typeId); skip if exists.
	existing, _ := s.licenseEntity.First(`user_id = ? AND workspace_id = ? AND type_id = ?`, p.UserID, p.WorkspaceID, lt.Id)
	if existing != nil {
		return nil
	}
	now := time.Now().UTC()
	lic := &bridgelicense.License{
		UserId:      p.UserID,
		ClientId:    p.ClientID,
		WorkspaceId: p.WorkspaceID,
		TypeId:      lt.Id,
		StartAt:     &now,
	}
	if p.PeriodType != "" {
		lic.PeriodType = &p.PeriodType
	}
	return s.licenseEntity.Insert(lic)
}

// OnPaymentPay ports the `apps.capital.payment.pay` subscription
// (ntx-apps/libs/bridge/src/app.controller.ts). It branches on the payment's
// meta.type and links/renews the matching license(s):
//   - apps.capital.license.pay     → the single license by meta.licenseId
//   - apps.capital.payment.renewal → each meta.logs[].meta.licenseId
//   - apps.capital.payment.record  → every license for (user, client, workspace)
//
// In each case it verifies the licenseType amount for the license's period
// matches the paid amount (else emits apps.capital.pay.mismatch and aborts),
// links the license (period→monthly, fresh token, start/renewed=now,
// expiredAt=+1 month, status=Success), and emits apps.identity.attribute.update.
//
// NB: TS does a redundant `status: Failed` write before the Success write when
// status!=true; that write is immediately overwritten, so (per the idiomatic-Go
// directive) we skip the dead write — the final status is always Success.
func (s *AppService) OnPaymentPay(payload map[string]any) error {
	if s.licenseEntity == nil || payload == nil {
		return nil
	}
	payment, _ := payload["payment"].(map[string]any)
	if payment == nil {
		return nil
	}
	userID, _ := payment["userId"].(string)
	clientID, _ := payment["clientId"].(string)
	workspaceID, _ := payment["workspaceId"].(string)
	amount := toFloat(payment["amount"])
	meta, _ := payment["meta"].(map[string]any)
	if meta == nil {
		return nil
	}
	typ, _ := meta["type"].(string)

	switch typ {
	case "apps.capital.license.pay":
		licenseID, _ := meta["licenseId"].(string)
		lic, _ := s.licenseEntity.First(`id = ?`, licenseID)
		if lic == nil {
			return nil
		}
		if !s.amountMatches(lic, amount) {
			s.payMismatch(payment, lic, userID, clientID, workspaceID)
			return nil
		}
		s.linkLicense(lic)
		s.attributeUpdate(userID, clientID, workspaceID)

	case "apps.capital.payment.renewal":
		renewalAmount := toFloat(meta["amount"]) // TS re-binds amount from meta
		logs, _ := meta["logs"].([]any)
		for _, lraw := range logs {
			log, _ := lraw.(map[string]any)
			lmeta, _ := log["meta"].(map[string]any)
			if lmeta == nil || lmeta["licenseId"] == nil {
				continue
			}
			licenseID, _ := lmeta["licenseId"].(string)
			lic, _ := s.licenseEntity.First(`id = ?`, licenseID)
			if lic == nil {
				continue
			}
			if !s.amountMatches(lic, renewalAmount) {
				s.payMismatch(payment, lic, userID, clientID, workspaceID)
				return nil // TS returns from the whole handler on a mismatch
			}
			s.linkLicense(lic)
			s.attributeUpdate(userID, clientID, workspaceID)
		}

	case "apps.capital.payment.record":
		licenses, _ := s.licenseEntity.Find(0, 0,
			`user_id = ? AND client_id = ? AND workspace_id = ?`, userID, clientID, workspaceID)
		for i := range licenses {
			lic := &licenses[i]
			if !s.amountMatches(lic, amount) {
				s.payMismatch(payment, lic, userID, clientID, workspaceID)
				return nil
			}
			s.linkLicense(lic)
			s.attributeUpdate(userID, clientID, workspaceID)
		}
	}
	return nil
}

// OnPaymentDebt ports the `apps.capital.payment.debt` subscription: for each
// payment whose meta.type is apps.capital.license.pay, mark its license Failed.
// Mirrors TS exactly: a payment whose type is NOT license.pay aborts the whole
// handler (return, not continue).
func (s *AppService) OnPaymentDebt(payload map[string]any) error {
	if s.licenseEntity == nil || payload == nil {
		return nil
	}
	payments, _ := payload["payments"].([]any)
	failed := "failed"
	for _, praw := range payments {
		payment, _ := praw.(map[string]any)
		meta, _ := payment["meta"].(map[string]any)
		if meta == nil {
			return nil
		}
		typ, _ := meta["type"].(string)
		if typ != "apps.capital.license.pay" {
			return nil
		}
		licenseID, _ := meta["licenseId"].(string)
		_, _ = s.licenseEntity.Update(&bridgelicense.License{Status: &failed}, `id = ?`, licenseID)
	}
	return nil
}

// amountMatches reports whether the license's type amount for its period equals
// the paid amount — TS `license.type[license.periodType] === amount`. A missing
// type, unknown/absent period amount counts as a mismatch (TS undefined !== n).
func (s *AppService) amountMatches(lic *bridgelicense.License, amount float64) bool {
	if s.licenseTypeEntity == nil {
		return false
	}
	lt, _ := s.licenseTypeEntity.First(`id = ?`, lic.TypeId)
	if lt == nil {
		return false
	}
	amt, ok := licenseTypeAmount(lt, strPtrVal(lic.PeriodType))
	return ok && amt == amount
}

func licenseTypeAmount(lt *bridgelicensetype.LicenseType, period string) (float64, bool) {
	switch period {
	case "daily":
		if lt.Daily != nil {
			return *lt.Daily, true
		}
	case "weekly":
		if lt.Weekly != nil {
			return *lt.Weekly, true
		}
	case "monthly":
		return lt.Monthly, true
	case "yearly":
		if lt.Yearly != nil {
			return *lt.Yearly, true
		}
	}
	return 0, false
}

// linkLicense applies the TS "link license" update: period→monthly, fresh token,
// start/renewed=now, expiredAt=+1 month, status=Success.
func (s *AppService) linkLicense(lic *bridgelicense.License) {
	now := time.Now().UTC()
	exp := now.AddDate(0, 1, 0)
	monthly := "monthly"
	tkn := utils.Reference("TKN", 36)
	success := "success"
	lic.PeriodType = &monthly
	lic.Token = &tkn
	lic.StartAt = &now
	lic.RenewedAt = &now
	lic.ExpiredAt = &exp
	lic.Status = &success
	_, _ = s.licenseEntity.Update(lic, `id = ?`, lic.Id)
}

func (s *AppService) payMismatch(payment map[string]any, lic *bridgelicense.License, userID, clientID, workspaceID string) {
	if s.tracker == nil {
		return
	}
	s.tracker.Message("apps.capital.pay.mismatch", map[string]any{
		"payment":     payment,
		"license":     lic,
		"userId":      userID,
		"clientId":    clientID,
		"workspaceId": workspaceID,
	})
}

func (s *AppService) attributeUpdate(userID, clientID, workspaceID string) {
	if s.tracker == nil {
		return
	}
	s.tracker.Message("apps.identity.attribute.update", map[string]any{
		"type":        "activity",
		"key":         "licenses",
		"value":       "1",
		"action":      "add",
		"userId":      userID,
		"clientId":    clientID,
		"workspaceId": workspaceID,
	})
}

func toFloat(v any) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	case int64:
		return float64(n)
	}
	return 0
}

func strPtrVal(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// runHeartbeat is the shared body for daily/weekly/monthly/yearly heartbeats.
// It scans licenses whose periodType matches and marks them as renewed when
// the period boundary has elapsed since last RenewedAt. Mirrors TS LicenseBatch
// keyed by PeriodType.
func (s *AppService) runHeartbeat(period string) error {
	if s.licenseEntity == nil {
		return nil
	}
	licenses, err := s.licenseEntity.Find(0, 0, `period_type = ?`, period)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	cutoff := cutoffFor(period, now)
	for _, lic := range licenses {
		lic := lic
		if lic.RenewedAt != nil && lic.RenewedAt.After(cutoff) {
			continue
		}
		lic.RenewedAt = &now
		_, _ = s.licenseEntity.Update(&lic, `id = ?`, lic.Id)
	}
	return nil
}

// cutoffFor returns the "before" instant a license must have been renewed
// after to be considered current for the given period.
func cutoffFor(period string, now time.Time) time.Time {
	switch period {
	case "daily":
		return now.Add(-24 * time.Hour)
	case "weekly":
		return now.Add(-7 * 24 * time.Hour)
	case "monthly":
		return now.AddDate(0, -1, 0)
	case "yearly":
		return now.AddDate(-1, 0, 0)
	}
	return now
}

func (s *AppService) OnDailyHeartbeat() error   { return s.runHeartbeat("daily") }
func (s *AppService) OnWeeklyHeartbeat() error  { return s.runHeartbeat("weekly") }
func (s *AppService) OnMonthlyHeartbeat() error { return s.runHeartbeat("monthly") }
func (s *AppService) OnYearlyHeartbeat() error  { return s.runHeartbeat("yearly") }

// OnLicenseJob handles a job from the `queue/apps/bridge/license` queue.
// Mirrors TS handler in ntx-apps/libs/bridge/src/index.ts jobs[0]:
//
//	const { plan } = job.data;
//	usageService.updateUsage(
//	  plan.userId, plan.clientId, plan.workspaceId,
//	  `apps/bridge/plan/${plan.type.key}/${plan.periodType}`,
//	  plan.type.id, { plan }, 1,
//	);
//
// The plan envelope on job.data carries the nested type with key+id, matching
// TS `relations: ['type']` from the queue producer.
func (s *AppService) OnLicenseJob(job *goqueues.QueueJob) (any, error) {
	if job == nil {
		return nil, nil
	}
	var data map[string]any
	if len(job.Data) > 0 {
		_ = json.Unmarshal(job.Data, &data)
	}
	plan, _ := data["plan"].(map[string]any)
	if plan == nil {
		return map[string]any{"jobId": job.Id, "status": "skipped"}, nil
	}
	userID, _ := plan["userId"].(string)
	clientID, _ := plan["clientId"].(string)
	workspaceID, _ := plan["workspaceId"].(string)
	periodType, _ := plan["periodType"].(string)
	planType, _ := plan["type"].(map[string]any)
	typeKey := ""
	typeID := ""
	if planType != nil {
		typeKey, _ = planType["key"].(string)
		typeID, _ = planType["id"].(string)
	}
	if userID == "" || typeID == "" {
		return map[string]any{"jobId": job.Id, "status": "skipped"}, nil
	}
	name := "apps/bridge/plan/" + typeKey + "/" + periodType
	if s.usageService != nil {
		s.usageService.UpdateUsage(userID, clientID, workspaceID, name, typeID, map[string]any{"plan": plan}, 1)
	}
	return map[string]any{"jobId": job.Id, "status": "dispatched", "name": name}, nil
}
