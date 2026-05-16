package sink

import (
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// SinkController mirrors ntx-apps/libs/polylog/src/api/sink/sink.controller.ts.
type SinkController struct {
	crud.CrudResource[Sink, CreateSinkDto, UpdateSinkDto]

	entity *SinkEntity   `inject:""`
	lang   *i18n.Service `inject:""`
}

func (c *SinkController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Sink, CreateSinkDto, UpdateSinkDto]{
		// Mirrors TS sink.controller.ts:19 name = 'sink'.
		Name: "sink",
		// Mirrors TS sink.controller.ts:20 searchable = ['name','desc'].
		Searchable: []string{"name", "desc"},
		// Mirrors TS sink.controller.ts:22 unique = ({workspaceId,name}) =>
		// [{workspaceId,name}].
		Unique: func(d *CreateSinkDto) []map[string]any {
			return []map[string]any{{
				"workspace_id": d.WorkspaceId,
				"name":         d.Name,
			}}
		},
		// Mirrors TS sink.controller.ts:24-34 morphs.beforeCreate.
		Morphs: map[string]crud.MorphFn{
			crud.BeforeCreate: func(payload any, ctx ntxctx.NTXContext) (any, error) {
				p, ok := payload.(*CreateSinkDto)
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
