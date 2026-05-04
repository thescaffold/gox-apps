package app

import "github.com/thescaffold/gox-packages-core/events"

const Name = "controller"

var AllMigrations = Migrations

var DefaultPermissions = map[string][]string{
	"guest":         {"member:read:apps:controller:now:*:*"},
	"member":        {"member:create:apps:controller:*:*:workspace", "member:update:apps:controller:*:*:self", "member:read:apps:controller:*:*:workspace", "member:delete:apps:controller:*:*:self"},
	"admin":         {"admin:*:apps:controller:*:*:workspace"},
	"global-member": {"global-member:*:apps:controller:*:*:global"},
	"global-admin":  {"global-admin:*:apps:controller:*:*:global"},
}

var Subscriptions = map[string]events.EventHandler{
	"apps.controller.route.register": func(_ string, _ any) {},
	"apps.cron.heartbeat.hourly":     func(_ string, _ any) {},
}
