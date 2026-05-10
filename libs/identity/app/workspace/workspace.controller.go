package workspace

import "github.com/thescaffold/gox-packages/libs/core/crud"

type WorkspaceController struct {
	crud.CrudResource[Workspace, CreateWorkspaceDto, UpdateWorkspaceDto]
	entity *WorkspaceEntity `inject:""`
}

func (c *WorkspaceController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Workspace, CreateWorkspaceDto, UpdateWorkspaceDto]{
		Name:       "IdentityWorkspace",
		Searchable: []string{"user_id", "client_id", "slug", "name"},
	})
}
