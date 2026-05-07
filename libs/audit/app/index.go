package app

import (
	auditlog "github.com/thescaffold/gox-apps-audit/app/log"
	auditbatch "github.com/thescaffold/gox-apps-audit/pkg/batch"
	"github.com/thescaffold/gox-packages-core/events"
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

// Subscriptions maps event patterns to handlers for this app.
var Subscriptions = map[string]events.EventHandler{
	"*.*.*.after-insert":           handleEvent,
	"*.*.*.after-update":           handleEvent,
	"*.*.*.after-delete":           handleEvent,
	"apps.cron.heartbeat.hourly":   auditbatch.Trigger,
	"apps.cron.heartbeat.weekly":   handleEvent,
	"apps.cron.heartbeat.monthly":  handleEvent,
	"apps.cron.heartbeat.yearly":   handleEvent,
}

// LogActionType aliases exported for host usage.
type LogActionType = auditlog.LogActionType
