package app

import "github.com/thescaffold/gox-packages-core/events"

const Name = "flags"

var AllMigrations = Migrations

// DefaultPermissions mirrors ntx-apps/libs/flags/src/index.ts defaultRolesNPermissions.
var DefaultPermissions = map[string][]string{
	"guest": {
		"guest:read:apps:flags:remote:*:self",
		"guest:read:apps:flags:dynamic:*:self",
		"guest:create:apps:flags:file:*:self",
		"guest:update:apps:flags:file:*:self",
		"guest:read:apps:flags:file:*:self",
	},
	"member": {
		"member:create:apps:flags:*:*:workspace",
		"member:update:apps:flags:*:*:self",
		"member:read:apps:flags:*:*:workspace",
		"member:delete:apps:flags:*:*:self",
	},
	"admin":         {"admin:*:apps:flags:*:*:workspace"},
	"global-member": {"global-member:*:apps:flags:*:*:global"},
	"global-admin":  {"global-admin:*:apps:flags:*:*:global"},
}

// Subscriptions maps event patterns to handlers, mirroring TS subscriptions list.
// Handlers are stubs in Go; the TS controller implements register/log persistence.
var Subscriptions = map[string]events.EventHandler{
	"apps.flags.flag.register": func(_ string, _ any) {},
	"apps.flags.flag.log":      func(_ string, _ any) {},
}

// Top-level exports mirroring ntx-apps/libs/flags/src/index.ts.
var (
	Messages        = map[string]any{}
	UnsafeEventList = []string{}
	Paths           = []string{"translations/en/ntx/apps/flags.yaml"}
	Jobs            = []any{}
	Crons           = []any{}
)
