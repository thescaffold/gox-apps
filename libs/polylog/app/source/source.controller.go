package source

import (
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// SourceController mirrors ntx-apps/libs/polylog/src/api/source/source.controller.ts.
type SourceController struct {
	crud.CrudResource[Source, CreateSourceDto, UpdateSourceDto]

	entity *SourceEntity `inject:""`
	lang   *i18n.Service `inject:""`
}

func (c *SourceController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Source, CreateSourceDto, UpdateSourceDto]{
		// Mirrors TS source.controller.ts:19 name = 'source'.
		Name: "source",
		// Mirrors TS source.controller.ts:20 searchable = ['name','desc'].
		Searchable: []string{"name", "desc"},
		// Mirrors TS source.controller.ts:22 unique = ({workspaceId,name}) =>
		// [{workspaceId,name}].
		Unique: func(d *CreateSourceDto) []map[string]any {
			return []map[string]any{{
				"workspace_id": d.WorkspaceId,
				"name":         d.Name,
			}}
		},
		// Mirrors TS source.controller.ts:24-36 hooks.beforeCreate — TS spreads
		// {userId,clientId,workspaceId} onto the payload before save. gox crud
		// only distinguishes morphs from hooks for distinct timing, but the
		// effect is identical here, so we wire it through Morphs.BeforeCreate.
		Morphs: map[string]crud.MorphFn{
			crud.BeforeCreate: func(payload any, ctx ntxctx.NTXContext) (any, error) {
				p, ok := payload.(*CreateSourceDto)
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
