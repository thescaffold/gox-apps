package permission

import "github.com/thescaffold/gox-packages-core/crud"

type PermissionController struct {
	crud.CrudResource[Permission, CreatePermissionDto, UpdatePermissionDto]
	entity *PermissionEntity `inject:""`
}

func (c *PermissionController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Permission, CreatePermissionDto, UpdatePermissionDto]{
		Name:       "IdentityPermission",
		Searchable: []string{"user_id", "client_id", "workspace_id", "role_id", "key"},
	})
}
