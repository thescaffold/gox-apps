package app

import "github.com/thescaffold/gox-packages-core/events"

const Name = "capital"

var AllMigrations = Migrations

// DefaultPermissions mirrors ntx-apps/libs/capital/src/index.ts defaultRolesNPermissions.
var DefaultPermissions = map[string][]string{
	"guest": {"guest:read:apps:capital:plan-type:*:*"},
	"member": {
		"member:create:apps:capital:*:*:workspace",
		"member:update:apps:capital:*:*:self",
		"member:read:apps:capital:*:*:workspace",
		"member:delete:apps:capital:*:*:self",
	},
	"admin":         {"admin:*:apps:capital:*:*:workspace"},
	"global-member": {"global-member:*:apps:capital:*:*:global"},
	"global-admin":  {"global-admin:*:apps:capital:*:*:global"},
}

// capitalAppSvc is set by AppModule.Declarations() so the subscription
// closures below can dispatch to the typed AppService methods after DI.
var capitalAppSvc *AppService

// payloadOf extracts the "payload" map from a polylog/eventbus envelope.
func payloadOf(v any) map[string]any {
	if m, ok := v.(map[string]any); ok {
		if p, ok := m["payload"].(map[string]any); ok {
			return p
		}
		return m
	}
	return nil
}

func decodeUCW(p map[string]any) IdentityUserClientWorkspacePayload {
	out := IdentityUserClientWorkspacePayload{}
	if meta, ok := p["meta"].(map[string]any); ok {
		out.UserID, _ = meta["userId"].(string)
		out.ClientID, _ = meta["clientId"].(string)
		out.WorkspaceID, _ = meta["workspaceId"].(string)
	}
	if ctx, ok := p["context"].(map[string]any); ok {
		if pref, ok := ctx["preference"].(map[string]any); ok {
			out.Currency, _ = pref["currency"].(string)
		}
	}
	return out
}

func decodeWallet(p map[string]any) WalletInitPayload {
	w := WalletInitPayload{}
	w.UserID, _ = p["userId"].(string)
	w.ClientID, _ = p["clientId"].(string)
	w.WorkspaceID, _ = p["workspaceId"].(string)
	w.Currency, _ = p["currency"].(string)
	w.Type, _ = p["type"].(string)
	w.Label, _ = p["label"].(string)
	return w
}

func decodeUsage(p map[string]any) UsageStartPayload {
	u := UsageStartPayload{}
	u.UserID, _ = p["userId"].(string)
	u.ClientID, _ = p["clientId"].(string)
	u.WorkspaceID, _ = p["workspaceId"].(string)
	u.Service, _ = p["service"].(string)
	u.Entity, _ = p["entity"].(string)
	u.EntityID, _ = p["entityId"].(string)
	u.Reference, _ = p["reference"].(string)
	return u
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
	pp.Plan, _ = p["plan"].(string)
	return pp
}

// Subscriptions mirrors TS subscriptions list. Each handler decodes the bus
// payload into a typed struct and forwards to a method on the resolved
// AppService singleton. The TS-equivalent bodies are stubbed inside
// AppService until the corresponding sub-services (Wallet/Usage/Payment/Plan)
// are ported.
var Subscriptions = map[string]events.EventHandler{
	"apps.db.identityuserclientworkspaces.after-insert": func(_ string, raw any) {
		if capitalAppSvc != nil {
			_ = capitalAppSvc.OnIdentityUserClientWorkspaceCreated(decodeUCW(payloadOf(raw)))
		}
	},
	"apps.capital.wallet.init": func(_ string, raw any) {
		if capitalAppSvc != nil {
			_ = capitalAppSvc.OnWalletInit(decodeWallet(payloadOf(raw)))
		}
	},
	"apps.capital.usage.start": func(_ string, raw any) {
		if capitalAppSvc != nil {
			_ = capitalAppSvc.OnUsageStart(decodeUsage(payloadOf(raw)))
		}
	},
	"apps.capital.usage.update": func(_ string, raw any) {
		if capitalAppSvc != nil {
			_ = capitalAppSvc.OnUsageUpdate(decodeUsage(payloadOf(raw)))
		}
	},
	"apps.capital.usage.stop": func(_ string, raw any) {
		if capitalAppSvc != nil {
			_ = capitalAppSvc.OnUsageStop(decodeUsage(payloadOf(raw)))
		}
	},
	"apps.cron.heartbeat.monthly": func(_ string, _ any) {
		if capitalAppSvc != nil {
			_ = capitalAppSvc.OnMonthlyHeartbeat()
		}
	},
	"apps.cron.heartbeat.yearly": func(_ string, _ any) {
		if capitalAppSvc != nil {
			_ = capitalAppSvc.OnYearlyHeartbeat()
		}
	},
	"apps.capital.rate.register": func(_ string, raw any) {
		if capitalAppSvc == nil {
			return
		}
		p := payloadOf(raw)
		r := RatePayload{}
		r.Type, _ = p["type"].(string)
		r.Currency, _ = p["currency"].(string)
		if v, ok := p["perUnit"].(float64); ok {
			n := int(v)
			r.PerUnit = &n
		}
		_ = capitalAppSvc.OnRateRegister(r)
	},
	"apps.capital.plan.register": func(_ string, raw any) {
		if capitalAppSvc == nil {
			return
		}
		p := payloadOf(raw)
		pl := PlanPayload{}
		pl.UserID, _ = p["userId"].(string)
		pl.ClientID, _ = p["clientId"].(string)
		pl.WorkspaceID, _ = p["workspaceId"].(string)
		pl.Key, _ = p["key"].(string)
		pl.Name, _ = p["name"].(string)
		pl.Description, _ = p["description"].(string)
		pl.Currency, _ = p["currency"].(string)
		if v, ok := p["monthly"].(float64); ok {
			pl.Monthly = int(v)
		}
		if v, ok := p["yearly"].(float64); ok {
			n := int(v)
			pl.Yearly = &n
		}
		pl.PeriodType, _ = p["periodType"].(string)
		pl.Type, _ = p["type"].(string)
		_ = capitalAppSvc.OnPlanRegister(pl)
	},
	"apps.capital.payment.pay": func(_ string, raw any) {
		if capitalAppSvc != nil {
			_ = capitalAppSvc.OnPaymentPay(decodePayment(payloadOf(raw)))
		}
	},
	"apps.capital.payment.debt": func(_ string, raw any) {
		if capitalAppSvc != nil {
			_ = capitalAppSvc.OnPaymentDebt(decodePayment(payloadOf(raw)))
		}
	},
}

// Top-level exports mirroring ntx-apps/libs/capital/src/index.ts.
var (
	Messages        = map[string]any{}
	UnsafeEventList = []string{}
	Paths           = []string{"translations/en/ntx/apps/capital.yaml"}
	Jobs            = []any{}
	Crons           = []any{}
)
