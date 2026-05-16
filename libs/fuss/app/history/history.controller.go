package history

import (
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// HistoryController mirrors ntx-apps/libs/fuss/src/api/history/history.controller.ts.
type HistoryController struct {
	crud.CrudResource[History, CreateHistoryDto, UpdateHistoryDto]

	entity *HistoryEntity `inject:""`
	lang   *i18n.Service  `inject:""`
}

func (c *HistoryController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[History, CreateHistoryDto, UpdateHistoryDto]{
		// Mirrors TS history.controller.ts:19 name = 'history'.
		Name: "history",
		// Mirrors TS history.controller.ts:20 searchable = ['service'] verbatim.
		// The History entity has no `service` column — TS keeps this placeholder
		// and gox preserves it so the surface stays identical.
		Searchable: []string{"service"},
		// Mirrors TS history.controller.ts:22-24:
		//   unique = ({workspaceId, type, query}) => [{workspaceId, type, query}]
		Unique: func(d *CreateHistoryDto) []map[string]any {
			return []map[string]any{{
				"workspace_id": d.WorkspaceId,
				"type":         d.Type,
				"query":        d.Query,
			}}
		},
		// Mirrors TS morphs.beforeCreate: spreads { userId, clientId,
		// workspaceId } from the request context onto the payload.
		Morphs: map[string]crud.MorphFn{
			crud.BeforeCreate: func(payload any, ctx ntxctx.NTXContext) (any, error) {
				p, ok := payload.(*CreateHistoryDto)
				if !ok || p == nil {
					return nil, nil
				}
				if ctx.User != nil {
					if id, ok := ctx.User["id"].(string); ok {
						p.UserId = id
					}
				}
				if ctx.Client != nil {
					if id, ok := ctx.Client["id"].(string); ok {
						p.ClientId = id
					}
				}
				if ctx.Workspace != nil {
					if id, ok := ctx.Workspace["id"].(string); ok {
						p.WorkspaceId = id
					}
				}
				return p, nil
			},
		},
	})
	c.SetLang(c.lang)
}
