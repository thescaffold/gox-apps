package app

import "github.com/thescaffold/gox-packages-core/events"

const Name = "health"

var AllMigrations = Migrations

var DefaultPermissions = map[string][]string{
	"guest":         {},
	"member":        {"member:create:apps:health:*:*:workspace", "member:update:apps:health:*:*:self", "member:read:apps:health:*:*:workspace", "member:delete:apps:health:*:*:self"},
	"admin":         {"admin:*:apps:health:*:*:workspace"},
	"global-member": {"global-member:*:apps:health:*:*:global"},
	"global-admin":  {"global-admin:*:apps:health:*:*:global"},
}

var healthSvc *AppService

func handleHeartbeat(_ string, payload any) {
	if healthSvc == nil {
		return
	}
	_ = payload
}

var Subscriptions = map[string]events.EventHandler{
	"apps.health.service.register": func(_ string, _ any) {},
	"apps.health.service.ping":     func(_ string, _ any) {},
	"apps.cron.heartbeat.hourly":   handleHeartbeat,
}

// Top-level exports mirroring ntx-apps/libs/health/src/index.ts.
var (
	Messages        = map[string]any{}
	UnsafeEventList = []string{}
	Paths           = []string{"translations/en/ntx/apps/health.yaml"}
	Jobs            = []any{}
	Crons           = []any{}
)
