package project

import (
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// ProjectController mirrors ntx-apps/libs/common/src/api/project/project.controller.ts.
type ProjectController struct {
	crud.CrudResource[Project, CreateProjectDto, UpdateProjectDto]

	entity *ProjectEntity `inject:""`
	lang   *i18n.Service  `inject:""`
}

func (c *ProjectController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Project, CreateProjectDto, UpdateProjectDto]{
		// Mirrors TS project.controller.ts:19 name = 'project'.
		Name: "project",
		// Mirrors TS project.controller.ts:20 searchable = [] (empty).
		Searchable: []string{},
		// Mirrors TS project.controller.ts:22-30 unique = ({workspaceId,
		// groupName,serviceName,entityName,entityId}) => [{...}].
		Unique: func(d *CreateProjectDto) []map[string]any {
			return []map[string]any{{
				"workspace_id": d.WorkspaceId,
				"group_name":   d.GroupName,
				"service_name": d.ServiceName,
				"entity_name":  d.EntityName,
				"entity_id":    d.EntityId,
			}}
		},
		// Mirrors TS project.controller.ts:32-43 morphs.beforeCreate.
		Morphs: map[string]crud.MorphFn{
			crud.BeforeCreate: func(payload any, ctx ntxctx.NTXContext) (any, error) {
				p, ok := payload.(*CreateProjectDto)
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
