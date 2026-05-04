package app

import "github.com/thescaffold/gox-packages-core/events"

const Name = "flags"

var AllMigrations = Migrations

var DefaultPermissions = map[string][]string{
	"member":        {"member:read:apps:flags:*:*:workspace"},
	"admin":         {"admin:*:apps:flags:*:*:workspace"},
	"global-admin":  {"global-admin:*:apps:flags:*:*:global"},
}

var Subscriptions = map[string]events.EventHandler{
	"apps.cron.heartbeat.hourly": func(_ string, _ any) {},
}
