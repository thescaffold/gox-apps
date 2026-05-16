package eventlog

import (
	"github.com/awesome-goose/goose/types"
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
	"github.com/thescaffold/gox-packages/libs/core/response"
)

// Status constants mirroring ntx-apps/libs/polylog/src/app.config.ts
// EventLogStatusType. Only Completed/InProgress are needed for the dashboard.
const (
	StatusCompleted  = "completed"
	StatusInProgress = "inprogress"
)

// EventLogController mirrors ntx-apps/libs/polylog/src/api/event-log/event-log.controller.ts.
// Exposes the CRUD surface from CrudResource plus GET /event-log/dashboard.
type EventLogController struct {
	crud.CrudResource[EventLog, CreateEventLogDto, UpdateEventLogDto]

	entity *EventLogEntity `inject:""`
	lang   *i18n.Service   `inject:""`
}

func (c *EventLogController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[EventLog, CreateEventLogDto, UpdateEventLogDto]{
		// Mirrors TS event-log.controller.ts:21 name = 'event log'.
		Name: "event log",
		// Mirrors TS event-log.controller.ts:22 searchable = [] (empty).
		Searchable: []string{},
		// TS unique = () => [] — no uniqueness constraints; left unset.
	})
	c.SetLang(c.lang)
}

// Dashboard mirrors TS @Get('dashboard') on event-log — counts EventLog rows
// whose linked Event has the given type, partitioned by EventLog status. The
// TS `where: [{ event: { type }, status }]` relation becomes a subquery in SQL.
func (c *EventLogController) Dashboard(dto *DashboardDto) types.Output {
	pref := dto.NTX.Preference
	title := c.lang.Translate("apps.polylog.event-log.title", nil, pref)
	completed, _ := c.entity.Count(
		`"status" = ? AND "event_id" IN (SELECT "id" FROM "PolylogEvents" WHERE "type" = ?)`,
		StatusCompleted, dto.Type,
	)
	inProgress, _ := c.entity.Count(
		`"status" = ? AND "event_id" IN (SELECT "id" FROM "PolylogEvents" WHERE "type" = ?)`,
		StatusInProgress, dto.Type,
	)
	return response.Success(map[string]any{
		"total":      completed + inProgress,
		"completed":  completed,
		"inProgress": inProgress,
	},
		title,
		c.lang.Translate("apps.polylog.event-log.get.dashboard.success", nil, pref),
		nil)
}
