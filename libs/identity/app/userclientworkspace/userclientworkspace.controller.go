package userclientworkspace

import (
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// UserClientWorkspaceController mirrors ntx-apps/libs/identity/src/api/user-client-workspace/user-client-workspace.controller.ts.
type UserClientWorkspaceController struct {
	crud.CrudResource[UserClientWorkspace, CreateUserClientWorkspaceDto, UpdateUserClientWorkspaceDto]

	entity *UserClientWorkspaceEntity `inject:""`
	lang   *i18n.Service              `inject:""`
}

func (c *UserClientWorkspaceController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[UserClientWorkspace, CreateUserClientWorkspaceDto, UpdateUserClientWorkspaceDto]{
		// Mirrors TS user-client-workspace.controller.ts name = 'user client workspace'.
		Name: "user client workspace",
		// Mirrors TS user-client-workspace.controller.ts searchable = [] (empty).
		Searchable: []string{},
		// Mirrors TS user-client-workspace.controller.ts unique = ({userId,clientId,
		// workspaceId}) => [{userId,clientId,workspaceId}].
		Unique: func(d *CreateUserClientWorkspaceDto) []map[string]any {
			return []map[string]any{{
				"user_id":      d.UserId,
				"client_id":    d.ClientId,
				"workspace_id": d.WorkspaceId,
			}}
		},
		// Mirrors TS user-client-workspace.controller.ts morphs.beforeCreate.
		Morphs: map[string]crud.MorphFn{
			crud.BeforeCreate: func(payload any, ctx ntxctx.NTXContext) (any, error) {
				p, ok := payload.(*CreateUserClientWorkspaceDto)
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
