package app

import "github.com/thescaffold/gox-packages-core/events"

const Name = "notification"

var AllMigrations = Migrations

var DefaultPermissions = map[string][]string{
	"member":       {"member:read:apps:notification:*:*:workspace"},
	"admin":        {"admin:*:apps:notification:*:*:workspace"},
	"global-admin": {"global-admin:*:apps:notification:*:*:global"},
}

var Subscriptions = map[string]events.EventHandler{
	"apps.cron.heartbeat.hourly": func(_ string, _ any) {},
}
