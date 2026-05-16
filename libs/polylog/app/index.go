package app

import (
	"encoding/json"
	"os"
	"strconv"
	"time"

	goqueues "github.com/awesome-goose/goose/modules/queues"
	polylogchannel "github.com/thescaffold/gox-apps/libs/polylog/app/channel"
	polylogevent "github.com/thescaffold/gox-apps/libs/polylog/app/event"
	polylogeventlog "github.com/thescaffold/gox-apps/libs/polylog/app/eventlog"
	polylogsink "github.com/thescaffold/gox-apps/libs/polylog/app/sink"
	polylogsinktype "github.com/thescaffold/gox-apps/libs/polylog/app/sinktype"
	polylogsource "github.com/thescaffold/gox-apps/libs/polylog/app/source"
	polylogsourcetype "github.com/thescaffold/gox-apps/libs/polylog/app/sourcetype"
	"github.com/thescaffold/gox-packages/libs/core/events"
)

const Name = "polylog"

var AllMigrations = Migrations

// DefaultPermissions mirrors ntx-apps/libs/polylog/src/index.ts defaultRolesNPermissions.
var DefaultPermissions = map[string][]string{
	"guest": {"guest:create:apps:polylog:root-source:*:workspace"},
	"member": {
		"member:create:apps:polylog:*:*:workspace",
		"member:update:apps:polylog:*:*:self",
		"member:read:apps:polylog:*:*:workspace",
		"member:delete:apps:polylog:*:*:self",
	},
	"admin":         {"admin:*:apps:polylog:*:*:workspace"},
	"global-member": {"global-member:*:apps:polylog:*:*:global"},
	"global-admin":  {"global-admin:*:apps:polylog:*:*:global"},
}

// polylogAppSvc is set by AppModule.Declarations() so subscription handlers
// can drive typed AppService methods after DI.
var polylogAppSvc *AppService

// handleHourlyHeartbeat mirrors TS subscription 'apps.cron.heartbeat.hourly':
// delete Event/EventLog rows older than LOG_RETENTION_THRESHOLD hours
// (default 24). TS swallows errors via console.info; gox does the same.
func handleHourlyHeartbeat(_ string, _ any) {
	if polylogAppSvc == nil {
		return
	}
	threshold := 24
	if v := os.Getenv("LOG_RETENTION_THRESHOLD"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			threshold = n
		}
	}
	cutoff := time.Now().UTC().Add(-time.Duration(threshold) * time.Hour)
	polylogAppSvc.Housekeep(cutoff)
}

// Subscriptions mirrors TS subscriptions list. TS comments out the
// source.register / sink.register subscriptions, so only the hourly
// heartbeat is wired here.
var Subscriptions = map[string]events.EventHandler{
	"apps.cron.heartbeat.hourly": handleHourlyHeartbeat,
}

// Top-level exports mirroring ntx-apps/libs/polylog/src/index.ts.
var (
	Entities = []any{
		polylogchannel.Channel{},
		polylogsink.Sink{},
		polylogsource.Source{},
		polylogsinktype.SinkType{},
		polylogsourcetype.SourceType{},
		polylogevent.Event{},
		polylogeventlog.EventLog{},
	}
	Messages        = map[string]any{}
	UnsafeEventList = []string{}
	Paths           = []string{"translations/en/ntx/apps/polylog.yaml"}
	Crons           = []any{}
)

// Jobs mirrors ntx-apps/libs/polylog/src/index.ts jobs array. The
// `queue/apps/polylog/ingest` consumer hydrates a Event from job.data and
// dispatches it through PipelineService.Process.
// Host code: queueApp.RegisterJob(polylog.Jobs[0]).
var Jobs = []*goqueues.JobHandler{
	goqueues.NewSimpleHandler("queue/apps/polylog", "ingest", func(job *goqueues.QueueJob) (any, error) {
		if polylogAppSvc == nil {
			return nil, nil
		}
		var ev polylogevent.Event
		if len(job.Data) > 0 {
			_ = json.Unmarshal(job.Data, &ev)
		}
		if ev.Id == "" {
			return map[string]any{"jobId": job.Id, "status": "skipped"}, nil
		}
		if err := polylogAppSvc.ProcessEvent(&ev); err != nil {
			return nil, err
		}
		return map[string]any{"jobId": job.Id, "status": "processed", "eventId": ev.Id}, nil
	}),
}
