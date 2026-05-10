package app

import "github.com/thescaffold/gox-packages-core/events"

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

func decodePayment(p map[string]any) PaymentPayload {
	pp := PaymentPayload{}
	pp.UserID, _ = p["userId"].(string)
	pp.ClientID, _ = p["clientId"].(string)
	pp.WorkspaceID, _ = p["workspaceId"].(string)
	if v, ok := p["amount"].(float64); ok {
		pp.Amount = v
	}
	pp.Currency, _ = p["currency"].(string)
	pp.Provider, _ = p["provider"].(string)
	pp.Reference, _ = p["reference"].(string)
	return pp
}

// Subscriptions mirrors TS subscriptions list. Each handler decodes the bus
// payload into a typed struct and forwards to a method on the resolved
// AppService singleton. TS-equivalent bodies are stubbed in AppService.
var Subscriptions = map[string]events.EventHandler{
	"apps.bridge.license.register": func(_ string, raw any) {
		if bridgeAppSvc != nil {
			_ = bridgeAppSvc.OnLicenseRegister(decodeLicense(payloadOf(raw)))
		}
	},
	"apps.capital.payment.pay": func(_ string, raw any) {
		if bridgeAppSvc != nil {
			_ = bridgeAppSvc.OnPaymentPay(decodePayment(payloadOf(raw)))
		}
	},
	"apps.capital.payment.debt": func(_ string, raw any) {
		if bridgeAppSvc != nil {
			_ = bridgeAppSvc.OnPaymentDebt(decodePayment(payloadOf(raw)))
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

// Top-level exports mirroring ntx-apps/libs/bridge/src/index.ts.
var (
	Messages        = map[string]any{}
	UnsafeEventList = []string{}
	Paths           = []string{"translations/en/ntx/apps/bridge.yaml"}
	Jobs            = []any{}
	Crons           = []any{}
)
