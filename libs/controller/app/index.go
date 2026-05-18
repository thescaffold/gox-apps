package app

import (
	"os"
	"strconv"
	"time"

	ctrlreq "github.com/thescaffold/gox-apps/libs/controller/app/request"
	ctrlroute "github.com/thescaffold/gox-apps/libs/controller/app/route"
	"github.com/thescaffold/gox-packages/libs/core/events"
)

const Name = "controller"

var AllMigrations = Migrations

var DefaultPermissions = map[string][]string{
	"guest":         {"member:read:apps:controller:now:*:*"},
	"member":        {"member:create:apps:controller:*:*:workspace", "member:update:apps:controller:*:*:self", "member:read:apps:controller:*:*:workspace", "member:delete:apps:controller:*:*:self"},
	"admin":         {"admin:*:apps:controller:*:*:workspace"},
	"global-member": {"global-member:*:apps:controller:*:*:global"},
	"global-admin":  {"global-admin:*:apps:controller:*:*:global"},
}

// controllerAppSvc is set by AppModule.Declarations() so subscription handlers
// can drive typed AppService methods after DI.
var controllerAppSvc *AppService

// handleRouteRegister mirrors TS subscription 'apps.controller.route.register':
// upsert one Http + one Ws Route entry by (group, service, type, name) using
// the event's payload map.
func handleRouteRegister(_ string, payload any) {
	if controllerAppSvc == nil {
		return
	}
	// goose events wrap payload in an envelope; unwrap to the raw payload map.
	switch p := payload.(type) {
	case map[string]any:
		if inner, ok := p["payload"]; ok {
			_ = controllerAppSvc.RegisterRoutePayload(inner)
			return
		}
		_ = controllerAppSvc.RegisterRoutePayload(p)
	default:
		_ = controllerAppSvc.RegisterRoutePayload(payload)
	}
}

// handleHourlyHeartbeat mirrors TS subscription 'apps.cron.heartbeat.hourly':
// delete ControllerRequests rows older than LOG_RETENTION_THRESHOLD hours
// (default 24). No status filter — matches TS.
func handleHourlyHeartbeat(_ string, _ any) {
	if controllerAppSvc == nil {
		return
	}
	threshold := 24
	if v := os.Getenv("LOG_RETENTION_THRESHOLD"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			threshold = n
		}
	}
	cutoff := time.Now().UTC().Add(-time.Duration(threshold) * time.Hour)
	controllerAppSvc.Housekeep(cutoff)
}

var Subscriptions = map[string]events.EventHandler{
	"apps.controller.route.register": handleRouteRegister,
	"apps.cron.heartbeat.hourly":     handleHourlyHeartbeat,
}

// Top-level exports mirroring ntx-apps/libs/controller/src/index.ts.
var (
	Entities        = []any{ctrlroute.Route{}, ctrlreq.Request{}}
	Messages        = map[string]any{}
	UnsafeEventList = []string{}
	Paths           = []string{"translations/en/ntx/apps/controller.yaml"}
	Jobs            = []any{}
	Crons           = []any{}
)
