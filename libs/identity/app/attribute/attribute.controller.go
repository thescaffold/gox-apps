package attribute

import (
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// AttributeController mirrors ntx-apps/libs/identity/src/api/attribute/attribute.controller.ts.
type AttributeController struct {
	crud.CrudResource[Attribute, CreateAttributeDto, UpdateAttributeDto]

	entity *AttributeEntity `inject:""`
	lang   *i18n.Service    `inject:""`
}

func (c *AttributeController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Attribute, CreateAttributeDto, UpdateAttributeDto]{
		// Mirrors TS attribute.controller.ts name = 'attribute'.
		Name: "attribute",
		// Mirrors TS attribute.controller.ts searchable = ['value'].
		Searchable: []string{"value"},
		// Mirrors TS attribute.controller.ts unique = ({workspaceId,type,key}) =>
		// [{workspaceId,type,key}].
		Unique: func(d *CreateAttributeDto) []map[string]any {
			return []map[string]any{{
				"workspace_id": d.WorkspaceId,
				"type":         d.Type,
				"key":          d.Key,
			}}
		},
		// Mirrors TS attribute.controller.ts morphs.beforeCreate.
		Morphs: map[string]crud.MorphFn{
			crud.BeforeCreate: func(payload any, ctx ntxctx.NTXContext) (any, error) {
				p, ok := payload.(*CreateAttributeDto)
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
						p.ClientId = &id
					}
				}
				if ctx.Workspace != nil {
					if id, ok := ctx.Workspace["id"].(string); ok && id != "" {
						p.WorkspaceId = &id
					}
				}
				return p, nil
			},
		},
	})
	c.SetLang(c.lang)
}
