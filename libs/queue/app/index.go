package app

import (
	"os"
	"strconv"
	"time"

	goqueues "github.com/awesome-goose/goose/modules/queues"
	"github.com/awesome-goose/goose/modules/sql"
	queuejob "github.com/thescaffold/gox-apps/libs/queue/app/job"
	queuelog "github.com/thescaffold/gox-apps/libs/queue/app/log"
	queuequeue "github.com/thescaffold/gox-apps/libs/queue/app/queue"
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

// queueAppSvc is set by AppModule.Declarations() so the subscription handler
// can drive typed AppService methods after DI.
var queueAppSvc *AppService

// handleHourlyHeartbeat mirrors TS subscription 'apps.cron.heartbeat.hourly':
// delete successful Queue/Job and Queue/Log rows older than
// LOG_RETENTION_THRESHOLD hours (default 24).
func handleHourlyHeartbeat(_ string, _ any) {
	if queueAppSvc == nil {
		return
	}
	threshold := 24
	if v := os.Getenv("LOG_RETENTION_THRESHOLD"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			threshold = n
		}
	}
	cutoff := time.Now().UTC().Add(-time.Duration(threshold) * time.Hour)
	queueAppSvc.Housekeep(cutoff)
}

var Subscriptions = map[string]events.EventHandler{
	"apps.cron.heartbeat.hourly": handleHourlyHeartbeat,
}

// Top-level exports mirroring ntx-apps/libs/queue/src/index.ts.
var (
	Entities        = []any{queuequeue.Queue{}, queuejob.Job{}, queuelog.Log{}}
	Messages        = map[string]any{}
	UnsafeEventList = []string{}
	Paths           = []string{"translations/en/ntx/apps/queue.yaml"}
	Crons           = []any{}
)
