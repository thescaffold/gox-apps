package app

import (
	"github.com/thescaffold/gox-apps/libs/assets/app/file"
	"github.com/thescaffold/gox-packages/libs/core/events"
)

const Name = "assets"

var AllMigrations = Migrations

var DefaultPermissions = map[string][]string{
	"guest":         {"guest:read:apps:assets:remote:*:self", "guest:read:apps:assets:dynamic:*:self", "guest:create:apps:assets:file:*:self", "guest:update:apps:assets:file:*:self", "guest:read:apps:assets:file:*:self"},
	"member":        {"member:create:apps:assets:*:*:workspace", "member:update:apps:assets:*:*:self", "member:read:apps:assets:*:*:workspace", "member:delete:apps:assets:*:*:self"},
	"admin":         {"admin:*:apps:assets:*:*:workspace"},
	"global-member": {"global-member:*:apps:assets:*:*:global"},
	"global-admin":  {"global-admin:*:apps:assets:*:*:global"},
}

// Subscriptions — assets has no event subscriptions.
var Subscriptions = map[string]events.EventHandler{}

// Top-level exports mirroring ntx-apps/libs/assets/src/index.ts.
var (
	Entities        = []any{file.File{}}
	Messages        = map[string]any{}
	UnsafeEventList = []string{}
	Paths           = []string{"translations/en/ntx/apps/assets.yaml"}
	Jobs            = []any{}
	Crons           = []any{}
)
