package app

import (
	"encoding/json"
	"time"

	goqueues "github.com/awesome-goose/goose/modules/queues"
	bridgelicense "github.com/thescaffold/gox-apps/libs/bridge/app/license"
	bridgelicensetype "github.com/thescaffold/gox-apps/libs/bridge/app/licensetype"
	bridgewebhook "github.com/thescaffold/gox-apps/libs/bridge/app/webhook"
	capitalusage "github.com/thescaffold/gox-apps/libs/capital/app/usage"
)

// AppService is the public surface for bridge event handlers.
// Mirrors ntx-apps/libs/bridge/src/app.service.ts + the subscription bodies
// in app.controller.ts.
type AppService struct {
	licenseEntity     *bridgelicense.LicenseEntity         `inject:""`
	licenseTypeEntity *bridgelicensetype.LicenseTypeEntity `inject:""`
	webhookEntity     *bridgewebhook.WebhookEntity         `inject:""`
	usageService      *capitalusage.UsageService           `inject:""`
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

// PaymentPayload bridges capital payment events.
type PaymentPayload struct {
	UserID      string  `json:"userId,omitempty"`
	ClientID    string  `json:"clientId,omitempty"`
	WorkspaceID string  `json:"workspaceId,omitempty"`
	Amount      float64 `json:"amount,omitempty"`
	Currency    string  `json:"currency,omitempty"`
	Provider    string  `json:"provider,omitempty"`
	Reference   string  `json:"reference,omitempty"`
}

// OnPaymentPay marks the user's most-recent license as renewed.
func (s *AppService) OnPaymentPay(p PaymentPayload) error {
	if s.licenseEntity == nil || p.UserID == "" {
		return nil
	}
	lic, _ := s.licenseEntity.First(`user_id = ? AND workspace_id = ?`, p.UserID, p.WorkspaceID)
	if lic == nil {
		return nil
	}
	now := time.Now().UTC()
	lic.RenewedAt = &now
	status := "active"
	lic.Status = &status
	_, err := s.licenseEntity.Update(lic, `id = ?`, lic.Id)
	return err
}

// OnPaymentDebt marks the user's most-recent license as suspended.
func (s *AppService) OnPaymentDebt(p PaymentPayload) error {
	if s.licenseEntity == nil || p.UserID == "" {
		return nil
	}
	lic, _ := s.licenseEntity.First(`user_id = ? AND workspace_id = ?`, p.UserID, p.WorkspaceID)
	if lic == nil {
		return nil
	}
	status := "suspended"
	lic.Status = &status
	_, err := s.licenseEntity.Update(lic, `id = ?`, lic.Id)
	return err
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
