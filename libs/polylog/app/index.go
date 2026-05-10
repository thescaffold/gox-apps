package app

import "github.com/thescaffold/gox-packages/libs/core/events"

const Name = "polylog"

var AllMigrations = Migrations

// DefaultPermissions mirrors ntx-apps/libs/polylog/src/index.ts defaultRolesNPermissions.
var DefaultPermissions = map[string][]string{
	"guest": {"guest:create:apps:polylog:root-source:*:workspace"},
	"member": {
		"member:create:apps:polylog:*:*:workspace",
		"member:update:apps:polylog:*:*:self",
		"member:read:apps:polylog:*:*:workspace",
		"member:delete:apps:polylog:*:*:self",
	},
	"admin":         {"admin:*:apps:polylog:*:*:workspace"},
	"global-member": {"global-member:*:apps:polylog:*:*:global"},
	"global-admin":  {"global-admin:*:apps:polylog:*:*:global"},
}

// Subscriptions mirrors TS subscriptions list. The after-insert wildcard is
// kept for now even though TS only declares the cron heartbeat — Go's database
// listener layer publishes through the bus and downstream pipelines depend on
// it. Both are fired into the bus.
var Subscriptions = map[string]events.EventHandler{
	"apps.cron.heartbeat.hourly": func(_ string, _ any) {},
	"*.*.*.after-insert":         func(_ string, _ any) {},
}

// Top-level exports mirroring ntx-apps/libs/polylog/src/index.ts.
var (
	Messages        = map[string]any{}
	UnsafeEventList = []string{}
	Paths           = []string{"translations/en/ntx/apps/polylog.yaml"}
	Jobs            = []any{}
	Crons           = []any{}
)
