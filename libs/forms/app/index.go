package app

import "github.com/thescaffold/gox-packages-core/events"

const Name = "forms"

var AllMigrations = Migrations

var DefaultPermissions = map[string][]string{
	"member":       {"member:read:apps:forms:*:*:workspace"},
	"admin":        {"admin:*:apps:forms:*:*:workspace"},
	"global-admin": {"global-admin:*:apps:forms:*:*:global"},
}

var Subscriptions = map[string]events.EventHandler{
	"apps.cron.heartbeat.hourly": func(_ string, _ any) {},
}
