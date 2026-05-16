package app

import (
	"encoding/json"

	healthlog "github.com/thescaffold/gox-apps/libs/health/app/log"
	healthservice "github.com/thescaffold/gox-apps/libs/health/app/service"
	healthsummary "github.com/thescaffold/gox-apps/libs/health/app/summary"
	"github.com/thescaffold/gox-packages/libs/core/events"
)

const Name = "health"

var AllMigrations = Migrations

// DefaultPermissions mirrors ntx-apps/libs/health/src/index.ts defaultRolesNPermissions.
var DefaultPermissions = map[string][]string{
	"guest":         {},
	"member":        {"member:create:apps:health:*:*:workspace", "member:update:apps:health:*:*:self", "member:read:apps:health:*:*:workspace", "member:delete:apps:health:*:*:self"},
	"admin":         {"admin:*:apps:health:*:*:workspace"},
	"global-member": {"global-member:*:apps:health:*:*:global"},
	"global-admin":  {"global-admin:*:apps:health:*:*:global"},
}

// healthSvc holds the DI-managed AppService so the top-level subscription
// handlers below can drive it without taking part in goose's struct injection.
// AppModule.Declarations() assigns this before the kernel boots.
var healthSvc *AppService

// stringOpt extracts an optional string property from a generic event payload.
// Returns (nil, true) if the key is present-but-null, (&value, true) for a
// string, or (nil, false) when the key is missing or wrong-typed.
func stringOpt(m map[string]any, key string) (*string, bool) {
	v, ok := m[key]
	if !ok {
		return nil, false
	}
	if v == nil {
		return nil, true
	}
	if s, ok := v.(string); ok {
		return &s, true
	}
	return nil, false
}

// payloadMap drills into `item.payload` — TS subscriptions receive `Item`s
// whose `payload` field carries the actual event data. gox publishes either
// the inner payload directly OR a wrapping {payload: …} envelope; we accept
// both shapes so producers stay free to pick the simplest form.
func payloadMap(raw any) map[string]any {
	m, ok := raw.(map[string]any)
	if !ok {
		return nil
	}
	if inner, ok := m["payload"].(map[string]any); ok {
		return inner
	}
	return m
}

// handleServiceRegister mirrors the TS subscription for
// `apps.health.service.register`: upserts a Service row by name.
func handleServiceRegister(_ string, payload any) {
	if healthSvc == nil {
		return
	}
	body := payloadMap(payload)
	if body == nil {
		return
	}
	name, _ := body["name"].(string)
	if name == "" {
		return
	}
	desc, _ := stringOpt(body, "desc")
	status, _ := stringOpt(body, "status")
	_, _ = healthSvc.Register(name, desc, status)
}

// handleServicePing mirrors the TS subscription for `apps.health.service.ping`:
// finds the service by name, writes a Log entry, then updates Service.state.
func handleServicePing(_ string, payload any) {
	if healthSvc == nil {
		return
	}
	body := payloadMap(payload)
	if body == nil {
		return
	}
	name, _ := body["name"].(string)
	state, _ := body["state"].(string)
	if name == "" || state == "" {
		return
	}
	var meta json.RawMessage
	if rawMeta, ok := body["meta"]; ok && rawMeta != nil {
		if b, err := json.Marshal(rawMeta); err == nil {
			meta = b
		}
	}
	_ = healthSvc.Ping(name, state, meta)
}

// handleHeartbeat mirrors the TS hourly subscription that purges old Log /
// Summary rows. Errors are swallowed — best-effort housekeeping.
func handleHeartbeat(_ string, _ any) {
	if healthSvc == nil {
		return
	}
	_ = healthSvc.Housekeep()
}

// Subscriptions mirrors ntx-apps/libs/health/src/index.ts subscriptions exactly:
// upsert on register, write log + bump state on ping, hourly housekeeping.
var Subscriptions = map[string]events.EventHandler{
	"apps.health.service.register": handleServiceRegister,
	"apps.health.service.ping":     handleServicePing,
	"apps.cron.heartbeat.hourly":   handleHeartbeat,
}

// Top-level exports mirroring ntx-apps/libs/health/src/index.ts.
var (
	Entities        = []any{healthlog.Log{}, healthservice.Service{}, healthsummary.Summary{}}
	Messages        = map[string]any{}
	UnsafeEventList = []string{}
	Paths           = []string{"translations/en/ntx/apps/health.yaml"}
	Jobs            = []any{}
	Crons           = []any{}
)
