package app

import (
	auditlog "github.com/thescaffold/gox-apps/libs/audit/app/log"
	"github.com/thescaffold/gox-packages/libs/core/events"
)

const Name = "audit"

var AllMigrations = Migrations

var DefaultPermissions = map[string][]string{
	"guest":         {},
	"member":        {"member:create:apps:audit:*:*:workspace", "member:update:apps:audit:*:*:self", "member:read:apps:audit:*:*:workspace", "member:delete:apps:audit:*:*:self"},
	"admin":         {"admin:*:apps:audit:*:*:workspace"},
	"global-member": {"global-member:*:apps:audit:*:*:global"},
	"global-admin":  {"global-admin:*:apps:audit:*:*:global"},
}

// appSvc is set by AppModule.Declarations so Subscriptions handlers can call it after DI.
var appSvc *AppService

func handleEvent(_ string, payload any) {
	if appSvc == nil {
		return
	}
	if m, ok := payload.(map[string]any); ok {
		appSvc.HandleEvent(m) //nolint:errcheck
	}
}

// summaryHandler returns a heartbeat handler that runs the periodic summary for
// `period`. Mirrors TS app.controller.ts which dispatches a SummaryBatch via
// batchService.run (async); we run it in a goroutine so the synchronous event
// bus is not blocked while every user's summary is aggregated.
func summaryHandler(period string) events.EventHandler {
	return func(_ string, _ any) {
		if appSvc == nil {
			return
		}
		go func() { _ = appSvc.RunSummary(period) }()
	}
}

// Subscriptions maps event patterns to handlers for this app.
// Mirrors ntx-apps/libs/audit/src/index.ts subscriptions exactly (6 entries):
// the after-* events create audit logs; the heartbeats run the period summary.
var Subscriptions = map[string]events.EventHandler{
	"*.*.*.after-insert":          handleEvent,
	"*.*.*.after-update":          handleEvent,
	"*.*.*.after-delete":          handleEvent,
	"apps.cron.heartbeat.weekly":  summaryHandler("weekly"),
	"apps.cron.heartbeat.monthly": summaryHandler("monthly"),
	"apps.cron.heartbeat.yearly":  summaryHandler("yearly"),
}

// LogActionType aliases exported for host usage.
type LogActionType = auditlog.LogActionType

// Top-level exports mirroring ntx-apps/libs/audit/src/index.ts.

// Entities is the list of entity types this app exposes for migration/registry.
var Entities = []any{auditlog.Log{}}

// Messages declares cross-app message handlers (TS messages = {}).
var Messages = map[string]any{}

// UnsafeEventList lists event names the app must not emit on (parity with TS).
var UnsafeEventList = []string{}

// Paths lists the i18n translation YAML files this app contributes.
var Paths = []string{"translations/en/ntx/apps/audit.yaml"}

// Jobs declares background jobs (TS jobs = []).
var Jobs = []any{}

// Crons declares cron schedules (TS crons = []).
var Crons = []any{}
