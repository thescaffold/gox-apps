package event

import (
	"github.com/awesome-goose/goose/types"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
	"github.com/thescaffold/gox-packages/libs/core/response"
)

// Status constants mirroring ntx-apps/libs/polylog/src/app.config.ts
// EventStatusType. Kept package-local; only Completed/InProgress are used by
// the dashboard count.
const (
	StatusCompleted  = "completed"
	StatusInProgress = "inprogress"
)

// EventController mirrors ntx-apps/libs/polylog/src/api/event/event.controller.ts.
// Exposes the CRUD surface from CrudResource plus GET /event/dashboard.
type EventController struct {
	crud.CrudResource[Event, CreateEventDto, UpdateEventDto]

	entity *EventEntity  `inject:""`
	lang   *i18n.Service `inject:""`
}

func (c *EventController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Event, CreateEventDto, UpdateEventDto]{
		// Mirrors TS event.controller.ts:22 name = 'event'.
		Name: "event",
		// Mirrors TS event.controller.ts:23 searchable = [] (empty).
		Searchable: []string{},
		// Mirrors TS event.controller.ts:25 unique = ({reference}) => [{reference}].
		Unique: func(d *CreateEventDto) []map[string]any {
			return []map[string]any{{"reference": d.Reference}}
		},
		// Mirrors TS event.controller.ts:27-38 morphs.beforeCreate.
		Morphs: map[string]crud.MorphFn{
			crud.BeforeCreate: func(payload any, ctx ntxctx.NTXContext) (any, error) {
				p, ok := payload.(*CreateEventDto)
				if !ok || p == nil {
					return nil, nil
				}
				if ctx.User != nil {
					if id, ok := ctx.User["id"].(string); ok && id != "" {
						p.UserId = id
					}
				}
				if ctx.Client != nil {
					if id, ok := ctx.Client["id"].(string); ok && id != "" {
						p.ClientId = id
					}
				}
				if ctx.Workspace != nil {
					if id, ok := ctx.Workspace["id"].(string); ok && id != "" {
						p.WorkspaceId = id
					}
				}
				return p, nil
			},
		},
	})
	c.SetLang(c.lang)
}

// Dashboard mirrors TS @Get('dashboard') — returns {total, completed, inProgress}
// counts, optionally filtered by `?type=`.
func (c *EventController) Dashboard(dto *DashboardDto) types.Output {
	pref := dto.NTX.Preference
	title := c.lang.Translate("apps.polylog.event.title", nil, pref)
	var completed, inProgress int64
	if dto.Type == "" {
		completed, _ = c.entity.Count(`"status" = ?`, StatusCompleted)
		inProgress, _ = c.entity.Count(`"status" = ?`, StatusInProgress)
	} else {
		completed, _ = c.entity.Count(`"type" = ? AND "status" = ?`, dto.Type, StatusCompleted)
		inProgress, _ = c.entity.Count(`"type" = ? AND "status" = ?`, dto.Type, StatusInProgress)
	}
	return response.Success(map[string]any{
		"total":      completed + inProgress,
		"completed":  completed,
		"inProgress": inProgress,
	},
		title,
		c.lang.Translate("apps.polylog.event.get.dashboard.success", nil, pref),
		nil)
}
