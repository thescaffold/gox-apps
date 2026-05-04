package app

import "github.com/thescaffold/gox-packages-core/events"

const Name = "bridge"

var AllMigrations = Migrations

var DefaultPermissions = map[string][]string{
	"member":       {"member:read:apps:bridge:*:*:workspace"},
	"admin":        {"admin:*:apps:bridge:*:*:workspace"},
	"global-admin": {"global-admin:*:apps:bridge:*:*:global"},
}

var Subscriptions = map[string]events.EventHandler{
	"apps.cron.heartbeat.daily": func(_ string, _ any) {},
}
