package app

import (
	"github.com/thescaffold/gox-apps-identity/pkg"
	"github.com/thescaffold/gox-packages-core/events"
)

const Name = "identity"

var AllMigrations = Migrations

// AuthModule and AuthMiddleware are re-exported here so host apps only need to
// import this package rather than the pkg sub-package directly.
var AuthModule = &pkg.AuthModule{}
var AuthMiddleware = pkg.AuthMiddleware

var DefaultPermissions = map[string][]string{
	"member":       {"member:read:apps:identity:*:*:workspace"},
	"admin":        {"admin:*:apps:identity:*:*:workspace"},
	"global-admin": {"global-admin:*:apps:identity:*:*:global"},
}

var Subscriptions = map[string]events.EventHandler{
	"apps.cron.heartbeat.hourly": func(_ string, _ any) {},
	"*.*.*.after-insert":         func(_ string, _ any) {},
}
