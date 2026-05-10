package userclientworkspace

import "github.com/thescaffold/gox-packages/libs/core/crud"

type UserClientWorkspaceController struct {
	crud.CrudResource[UserClientWorkspace, CreateUserClientWorkspaceDto, UpdateUserClientWorkspaceDto]
	entity *UserClientWorkspaceEntity `inject:""`
}

func (c *UserClientWorkspaceController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[UserClientWorkspace, CreateUserClientWorkspaceDto, UpdateUserClientWorkspaceDto]{
		Name:       "IdentityUserClientWorkspace",
		Searchable: []string{"user_id", "client_id", "workspace_id"},
	})
}
