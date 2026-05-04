package app

import (
	"github.com/thescaffold/gox-apps-statics/app/list"
	"github.com/thescaffold/gox-packages-core/events"
)

const Name = "statics"

// re-export for host mono repo
var AllMigrations = Migrations
var AllSeeders = Seeders

// DefaultPermissions mirrors the TS defaultRolesNPermissions.
var DefaultPermissions = map[string][]string{
	"guest":         {"guest:read:apps:statics:filter:*:*"},
	"member":        {"member:create:apps:statics:*:*:workspace", "member:update:apps:statics:*:*:self", "member:read:apps:statics:*:*:workspace", "member:delete:apps:statics:*:*:self"},
	"admin":         {"admin:*:apps:statics:*:*:workspace"},
	"global-member": {"global-member:*:apps:statics:*:*:global"},
	"global-admin":  {"global-admin:*:apps:statics:*:*:global"},
}

// Subscriptions lists events this app subscribes to (none for statics).
// TODO: i18n — translation path: translations/en/ntx/apps/statics.yaml
var Subscriptions = map[string]events.EventHandler{}

// ListKeyType aliases exported for host usage.
type ListKeyType = list.ListKeyType
