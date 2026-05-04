package app

import "github.com/thescaffold/gox-packages-core/events"

const Name = "cache"

var AllMigrations = Migrations

var DefaultPermissions = map[string][]string{
	"guest":         {},
	"member":        {"member:create:apps:cache:*:*:workspace", "member:update:apps:cache:*:*:self", "member:read:apps:cache:*:*:workspace", "member:delete:apps:cache:*:*:self"},
	"admin":         {"admin:*:apps:cache:*:*:workspace"},
	"global-member": {"global-member:*:apps:cache:*:*:global"},
	"global-admin":  {"global-admin:*:apps:cache:*:*:global"},
}

var Subscriptions = map[string]events.EventHandler{}
