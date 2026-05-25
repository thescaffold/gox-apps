package app

import (
	goqueues "github.com/awesome-goose/goose/modules/queues"
	bridgelicense "github.com/thescaffold/gox-apps/libs/bridge/app/license"
	bridgelicensetype "github.com/thescaffold/gox-apps/libs/bridge/app/licensetype"
	bridgeplantype "github.com/thescaffold/gox-apps/libs/bridge/app/plantype"
	bridgepreference "github.com/thescaffold/gox-apps/libs/bridge/app/preference"
	bridgewebhook "github.com/thescaffold/gox-apps/libs/bridge/app/webhook"
	bridgewebhooklog "github.com/thescaffold/gox-apps/libs/bridge/app/webhooklog"
	"github.com/thescaffold/gox-packages/libs/core/events"
)

const Name = "bridge"

var AllMigrations = Migrations

// DefaultPermissions mirrors ntx-apps/libs/bridge/src/index.ts defaultRolesNPermissions.
var DefaultPermissions = map[string][]string{
	"guest": {"guest:read:apps:bridge:license-type:*:self"},
	"member": {
		"member:create:apps:bridge:*:*:workspace",
		"member:update:apps:bridge:*:*:self",
		"member:read:apps:bridge:*:*:workspace",
		"member:delete:apps:bridge:*:*:self",
	},
	"admin":         {"admin:*:apps:bridge:*:*:workspace"},
	"global-member": {"global-member:*:apps:bridge:*:*:global"},
	"global-admin":  {"global-admin:*:apps:bridge:*:*:global"},
}

// bridgeAppSvc is set by AppModule.Declarations() so the subscription
// closures below can dispatch to typed AppService methods after DI.
var bridgeAppSvc *AppService

func payloadOf(v any) map[string]any {
	if m, ok := v.(map[string]any); ok {
		if p, ok := m["payload"].(map[string]any); ok {
			return p
		}
		return m
	}
	return nil
}

func decodeLicense(p map[string]any) LicensePayload {
	l := LicensePayload{}
	l.UserID, _ = p["userId"].(string)
	l.ClientID, _ = p["clientId"].(string)
	l.WorkspaceID, _ = p["workspaceId"].(string)
	l.TypeKey, _ = p["typeKey"].(string)
	l.TypeName, _ = p["typeName"].(string)
	l.Currency, _ = p["currency"].(string)
	if v, ok := p["monthly"].(float64); ok {
		l.Monthly = v
	}
	l.PeriodType, _ = p["periodType"].(string)
	l.Meta, _ = p["meta"].(map[string]any)
	return l
}

// Subscriptions mirrors TS subscriptions list. Each handler forwards the bus
// payload to a method on the resolved AppService singleton.
var Subscriptions = map[string]events.EventHandler{
	"apps.bridge.license.register": func(_ string, raw any) {
		if bridgeAppSvc != nil {
			_ = bridgeAppSvc.OnLicenseRegister(decodeLicense(payloadOf(raw)))
		}
	},
	"apps.capital.payment.pay": func(_ string, raw any) {
		if bridgeAppSvc != nil {
			_ = bridgeAppSvc.OnPaymentPay(payloadOf(raw))
		}
	},
	"apps.capital.payment.debt": func(_ string, raw any) {
		if bridgeAppSvc != nil {
			_ = bridgeAppSvc.OnPaymentDebt(payloadOf(raw))
		}
	},
	"apps.cron.heartbeat.daily": func(_ string, _ any) {
		if bridgeAppSvc != nil {
			_ = bridgeAppSvc.OnDailyHeartbeat()
		}
	},
	"apps.cron.heartbeat.weekly": func(_ string, _ any) {
		if bridgeAppSvc != nil {
			_ = bridgeAppSvc.OnWeeklyHeartbeat()
		}
	},
	"apps.cron.heartbeat.monthly": func(_ string, _ any) {
		if bridgeAppSvc != nil {
			_ = bridgeAppSvc.OnMonthlyHeartbeat()
		}
	},
	"apps.cron.heartbeat.yearly": func(_ string, _ any) {
		if bridgeAppSvc != nil {
			_ = bridgeAppSvc.OnYearlyHeartbeat()
		}
	},
}

// Jobs mirrors ntx-apps/libs/bridge/src/index.ts jobs array. The `license`
// queue handler logs the licence for the next payment cycle via the capital
// app's UsageService. Capital's UsageService is wired in Phase 7.5; until
// then the handler logs the job and returns success so the queue worker
// (Phase 3) treats it as drained rather than retrying indefinitely.
//
// TS handler:
//
//	queue: 'queue/apps/bridge', job: 'license', fn: async (job) => {
//	  ...syncService.once(`apps:bridge:payment:queue:${plan.id}`, async () => {
//	    usageService.updateUsage(userId, clientId, workspaceId,
//	      `apps/bridge/plan/${plan.type.key}/${plan.periodType}`,
//	      plan.type.id, { plan }, 1);
//	  });
//	}
//
// The handler is registered against the queue.AppService at boot — that
// happens in the host application after both bridge and queue modules load.
// Host code: `queueApp.RegisterJob(bridge.Jobs[0])` (or RegisterJobs).
var Jobs = []*goqueues.JobHandler{
	goqueues.NewSimpleHandler("queue/apps/bridge", "license", func(job *goqueues.QueueJob) (any, error) {
		if bridgeAppSvc == nil {
			return nil, nil
		}
		return bridgeAppSvc.OnLicenseJob(job)
	}),
}

// Top-level exports mirroring ntx-apps/libs/bridge/src/index.ts.
var (
	Entities = []any{
		bridgelicense.License{},
		bridgelicensetype.LicenseType{},
		bridgeplantype.PlanType{},
		bridgepreference.Preference{},
		bridgewebhook.Webhook{},
		bridgewebhooklog.WebhookLog{},
	}
	Messages        = map[string]any{}
	UnsafeEventList = []string{}
	Paths           = []string{"translations/en/ntx/apps/bridge.yaml"}
	Crons           = []any{}
)
