package tag

import (
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// TagController mirrors ntx-apps/libs/common/src/api/tag/tag.controller.ts.
type TagController struct {
	crud.CrudResource[Tag, CreateTagDto, UpdateTagDto]

	entity *TagEntity    `inject:""`
	lang   *i18n.Service `inject:""`
}

func (c *TagController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Tag, CreateTagDto, UpdateTagDto]{
		// Mirrors TS tag.controller.ts:19 name = 'tag'.
		Name: "tag",
		// Mirrors TS tag.controller.ts:20 searchable = [] (empty list).
		Searchable: []string{},
		// Mirrors TS tag.controller.ts:22-28 unique = ({workspaceId,groupName,
		// serviceName,entityName,entityId}) => [{...}].
		Unique: func(d *CreateTagDto) []map[string]any {
			return []map[string]any{{
				"workspace_id": d.WorkspaceId,
				"group_name":   d.GroupName,
				"service_name": d.ServiceName,
				"entity_name":  d.EntityName,
				"entity_id":    d.EntityId,
			}}
		},
		// Mirrors TS tag.controller.ts:30-40 morphs.beforeCreate: spreads
		// userId/clientId/workspaceId from context onto payload.
		Morphs: map[string]crud.MorphFn{
			crud.BeforeCreate: func(payload any, ctx ntxctx.NTXContext) (any, error) {
				p, ok := payload.(*CreateTagDto)
				if !ok || p == nil {
					return nil, nil
				}
				if ctx.User != nil {
					if id, ok := ctx.User["id"].(string); ok && id != "" {
						p.UserId = &id
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
