package app

import (
	goqueues "github.com/awesome-goose/goose/modules/queues"
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/thescaffold/gox-packages/libs/core/events"
)

const Name = "queue"

var Migrations = []sql.Migration{}

var AllMigrations = Migrations

var DefaultPermissions = map[string][]string{
	"guest":         {},
	"member":        {"member:create:apps:queue:*:*:workspace", "member:update:apps:queue:*:*:self", "member:read:apps:queue:*:*:workspace", "member:delete:apps:queue:*:*:self"},
	"admin":         {"admin:*:apps:queue:*:*:workspace"},
	"global-member": {"global-member:*:apps:queue:*:*:global"},
	"global-admin":  {"global-admin:*:apps:queue:*:*:global"},
}

// Jobs holds pre-registered job handlers for the host to wire in.
// Populate at startup if custom queue processors are needed.
var Jobs = []*goqueues.JobHandler{}

func handleHourlyHeartbeat(_ string, _ any) {
	// housekeeping: goqueues.NewModule CleanupInterval handles retention
}

var Subscriptions = map[string]events.EventHandler{
	"apps.cron.heartbeat.hourly": handleHourlyHeartbeat,
}

// Top-level exports mirroring ntx-apps/libs/queue/src/index.ts.
var (
	Messages        = map[string]any{}
	UnsafeEventList = []string{}
	Paths           = []string{"translations/en/ntx/apps/queue.yaml"}
	Crons           = []any{}
)
