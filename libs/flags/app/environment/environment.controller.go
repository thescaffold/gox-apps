package environment

import (
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// EnvironmentController mirrors ntx-apps/libs/flags/src/api/environment/environment.controller.ts.
type EnvironmentController struct {
	crud.CrudResource[Environment, CreateEnvironmentDto, UpdateEnvironmentDto]
	entity *EnvironmentEntity `inject:""`
	lang   *i18n.Service      `inject:""`
}

func (c *EnvironmentController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Environment, CreateEnvironmentDto, UpdateEnvironmentDto]{
		// Mirrors TS environment.controller.ts:19 name = 'environment'.
		Name: "environment",
		// Mirrors TS environment.controller.ts:20 searchable = ['name','desc'].
		Searchable: []string{"name", "desc"},
		// Mirrors TS environment.controller.ts:22-24:
		//   unique = ({workspaceId, name}) => [{workspaceId, name}]
		Unique: func(d *CreateEnvironmentDto) []map[string]any {
			return []map[string]any{{"workspace_id": d.WorkspaceId, "name": d.Name}}
		},
		// TS environment.controller.ts registers this beforeCreate spread as a
		// `hooks` entry — strictly that's a TS bug (hooks are side-effect only;
		// mutating the payload via the return value relies on `morphs` semantics).
		// gox wires it as Morphs.BeforeCreate so the spread actually lands on
		// the persisted row, matching the TS intent.
		Morphs: map[string]crud.MorphFn{
			crud.BeforeCreate: func(payload any, ctx ntxctx.NTXContext) (any, error) {
				p, ok := payload.(*CreateEnvironmentDto)
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
