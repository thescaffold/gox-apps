package app

import "github.com/thescaffold/gox-packages-core/events"

const Name = "notification"

var AllMigrations = Migrations

// DefaultPermissions mirrors ntx-apps/libs/notification/src/index.ts defaultRolesNPermissions.
var DefaultPermissions = map[string][]string{
	"guest": {},
	"member": {
		"member:create:apps:notification:*:*:workspace",
		"member:update:apps:notification:*:*:self",
		"member:read:apps:notification:*:*:workspace",
		"member:delete:apps:notification:*:*:self",
	},
	"admin":         {"admin:*:apps:notification:*:*:workspace"},
	"global-member": {"global-member:*:apps:notification:*:*:global"},
	"global-admin":  {"global-admin:*:apps:notification:*:*:global"},
}

// Subscriptions mirrors TS subscriptions list. Handlers are placeholders;
// the TS controller persists messages and dispatches via providers.
var Subscriptions = map[string]events.EventHandler{
	"apps.notification.message.new":       func(_ string, _ any) {},
	"apps.identity.attribute.type.update": func(_ string, _ any) {},
}

// Top-level exports mirroring ntx-apps/libs/notification/src/index.ts.
var (
	Messages        = map[string]any{}
	UnsafeEventList = []string{}
	Paths           = []string{"translations/en/ntx/apps/notification.yaml"}
	Jobs            = []any{}
	Crons           = []any{}
)
