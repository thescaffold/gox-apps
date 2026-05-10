package app

import (
	filepkg "github.com/thescaffold/gox-apps-blobs/app/file"
	pagepkg "github.com/thescaffold/gox-apps-blobs/app/page"
	"github.com/thescaffold/gox-packages-core/events"
)

const Name = "blobs"

var AllMigrations = Migrations

// DefaultPermissions mirrors ntx-apps/libs/blobs/src/index.ts defaultRolesNPermissions exactly.
var DefaultPermissions = map[string][]string{
	"guest": {
		"guest:read:apps:blobs:remote:*:self",
		// Added per TS index.ts:14 — Go was missing dynamic read.
		"guest:read:apps:blobs:dynamic:*:self",
		"guest:create:apps:blobs:file:*:self",
		"guest:update:apps:blobs:file:*:self",
		"guest:read:apps:blobs:file:*:self",
	},
	"member":        {"member:create:apps:blobs:*:*:workspace", "member:update:apps:blobs:*:*:self", "member:read:apps:blobs:*:*:workspace", "member:delete:apps:blobs:*:*:self"},
	"admin":         {"admin:*:apps:blobs:*:*:workspace"},
	"global-member": {"global-member:*:apps:blobs:*:*:global"},
	"global-admin":  {"global-admin:*:apps:blobs:*:*:global"},
}

var Subscriptions = map[string]events.EventHandler{}

// Top-level exports mirroring ntx-apps/libs/blobs/src/index.ts.

// Entities is the list of entity types this app exposes for migration/registry.
var Entities = []any{filepkg.File{}, pagepkg.Page{}}

// Messages declares cross-app message handlers (TS messages = {}).
var Messages = map[string]any{}

// UnsafeEventList lists event names the app must not emit on (TS = []).
var UnsafeEventList = []string{}

// Paths lists the i18n translation YAML files this app contributes.
var Paths = []string{"translations/en/ntx/apps/blobs.yaml"}

// Jobs declares background jobs (TS jobs = []).
var Jobs = []any{}

// Crons declares cron schedules (TS crons = []).
var Crons = []any{}
