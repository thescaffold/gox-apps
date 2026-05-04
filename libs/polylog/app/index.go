package app

import "github.com/thescaffold/gox-packages-core/events"

const Name = "polylog"

var AllMigrations = Migrations

var DefaultPermissions = map[string][]string{
	"member":       {"member:read:apps:polylog:*:*:workspace"},
	"admin":        {"admin:*:apps:polylog:*:*:workspace"},
	"global-admin": {"global-admin:*:apps:polylog:*:*:global"},
}

var Subscriptions = map[string]events.EventHandler{
	"apps.cron.heartbeat.hourly": func(_ string, _ any) {},
	"*.*.*.after-insert":         func(_ string, _ any) {},
}
