package app

import (
	commonip "github.com/thescaffold/gox-apps/libs/common/app/ip"
	commonproject "github.com/thescaffold/gox-apps/libs/common/app/project"
	commonprojecttype "github.com/thescaffold/gox-apps/libs/common/app/projecttype"
	commonrate "github.com/thescaffold/gox-apps/libs/common/app/rate"
	commonratelog "github.com/thescaffold/gox-apps/libs/common/app/ratelog"
	commontag "github.com/thescaffold/gox-apps/libs/common/app/tag"
	commontagtype "github.com/thescaffold/gox-apps/libs/common/app/tagtype"
	"github.com/thescaffold/gox-packages/libs/core/events"
)

const Name = "common"

var AllMigrations = Migrations

var DefaultPermissions = map[string][]string{
	"guest":         {"guest:read:apps:common:location:*:*", "guest:read:apps:common:manifest:*:*"},
	"member":        {"member:create:apps:common:*:*:workspace", "member:update:apps:common:*:*:self", "member:read:apps:common:*:*:workspace", "member:delete:apps:common:*:*:self"},
	"admin":         {"admin:*:apps:common:*:*:workspace"},
	"global-member": {"global-member:*:apps:common:*:*:global"},
	"global-admin":  {"global-admin:*:apps:common:*:*:global"},
}

// commonAppSvc is set by AppModule.Declarations so the package-level weekly
// heartbeat handler can call into AppService after DI has wired everything.
var commonAppSvc *AppService

// handleWeeklyHeartbeat mirrors TS app.controller.ts 'apps.cron.heartbeat.weekly':
// upserts a Rate row + appends a RateLog entry per currency returned by the
// exchangeratesapi provider. Errors are swallowed (cron runs again next week).
func handleWeeklyHeartbeat(_ string, _ any) {
	if commonAppSvc == nil {
		return
	}
	_ = commonAppSvc.WeeklyHeartbeat()
}

var Subscriptions = map[string]events.EventHandler{
	"apps.cron.heartbeat.weekly": handleWeeklyHeartbeat,
}

// Top-level exports mirroring ntx-apps/libs/common/src/index.ts.
var (
	Entities = []any{
		commontag.Tag{},
		commontagtype.TagType{},
		commonip.Ip{},
		commonrate.Rate{},
		commonratelog.RateLog{},
		commonprojecttype.ProjectType{},
		commonproject.Project{},
	}
	Messages        = map[string]any{}
	UnsafeEventList = []string{}
	Paths           = []string{"translations/en/ntx/apps/common.yaml"}
	Jobs            = []any{}
	Crons           = []any{}
)
