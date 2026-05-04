package app

import "github.com/thescaffold/gox-packages-core/events"

const Name = "capital"

var AllMigrations = Migrations

var DefaultPermissions = map[string][]string{
	"member":       {"member:read:apps:capital:*:*:workspace"},
	"admin":        {"admin:*:apps:capital:*:*:workspace"},
	"global-admin": {"global-admin:*:apps:capital:*:*:global"},
}

var Subscriptions = map[string]events.EventHandler{
	"apps.cron.heartbeat.daily": func(_ string, _ any) {},
}
