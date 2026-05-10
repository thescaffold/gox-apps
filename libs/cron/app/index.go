package app

import (
	"context"

	gocron "github.com/awesome-goose/goose/modules/cron"
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/thescaffold/gox-packages-core/events"
)

const Name = "cron"

var Migrations = []sql.Migration{}

var AllMigrations = Migrations

var DefaultPermissions = map[string][]string{
	"guest":         {},
	"member":        {"member:create:apps:cron:*:*:workspace", "member:update:apps:cron:*:*:self", "member:read:apps:cron:*:*:workspace", "member:delete:apps:cron:*:*:self"},
	"admin":         {"admin:*:apps:cron:*:*:workspace"},
	"global-member": {"global-member:*:apps:cron:*:*:global"},
	"global-admin":  {"global-admin:*:apps:cron:*:*:global"},
}

var cronAppSvc *AppService

func heartbeatHandler(eventType string) gocron.CronHandlerFn {
	return func(_ context.Context, job *gocron.CronJob) (any, error) {
		if cronAppSvc != nil {
			cronAppSvc.PublishHeartbeat(eventType, job)
		}
		return nil, nil
	}
}

var Jobs = []*gocron.CronHandler{
	gocron.NewHandler("apps.cron", "heartbeat.minute", "0 * * * * *", heartbeatHandler("apps.cron.heartbeat.minute")),
	gocron.NewHandler("apps.cron", "heartbeat.hourly", "0 0 * * * *", heartbeatHandler("apps.cron.heartbeat.hourly")),
	gocron.NewHandler("apps.cron", "heartbeat.daily", "0 0 0 * * *", heartbeatHandler("apps.cron.heartbeat.daily")),
	gocron.NewHandler("apps.cron", "heartbeat.weekly", "0 0 0 * * 0", heartbeatHandler("apps.cron.heartbeat.weekly")),
	gocron.NewHandler("apps.cron", "heartbeat.monthly", "0 0 0 1 * *", heartbeatHandler("apps.cron.heartbeat.monthly")),
	gocron.NewHandler("apps.cron", "heartbeat.yearly", "0 0 0 1 1 *", heartbeatHandler("apps.cron.heartbeat.yearly")),
}

func handleHourlyHeartbeat(_ string, _ any) {
	// housekeeping: the cron module's own CleanupInterval handles log retention
}

var Subscriptions = map[string]events.EventHandler{
	"apps.cron.heartbeat.hourly": handleHourlyHeartbeat,
}

// Top-level exports mirroring ntx-apps/libs/cron/src/index.ts.
var (
	Messages        = map[string]any{}
	UnsafeEventList = []string{}
	Paths           = []string{"translations/en/ntx/apps/cron.yaml"}
	Crons           = []any{}
)
