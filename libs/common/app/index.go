package app

import "github.com/thescaffold/gox-packages-core/events"

const Name = "common"

var AllMigrations = Migrations

var DefaultPermissions = map[string][]string{
	"guest":         {"guest:read:apps:common:location:*:*", "guest:read:apps:common:manifest:*:*"},
	"member":        {"member:create:apps:common:*:*:workspace", "member:update:apps:common:*:*:self", "member:read:apps:common:*:*:workspace", "member:delete:apps:common:*:*:self"},
	"admin":         {"admin:*:apps:common:*:*:workspace"},
	"global-member": {"global-member:*:apps:common:*:*:global"},
	"global-admin":  {"global-admin:*:apps:common:*:*:global"},
}

var Subscriptions = map[string]events.EventHandler{
	"apps.cron.heartbeat.weekly": func(_ string, _ any) {},
}
