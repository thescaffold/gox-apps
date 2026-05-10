package permissiontype

import "github.com/thescaffold/gox-packages/libs/core/crud"

type PermissionTypeController struct {
	crud.CrudResource[PermissionType, CreatePermissionTypeDto, UpdatePermissionTypeDto]
	entity *PermissionTypeEntity `inject:""`
}

func (c *PermissionTypeController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[PermissionType, CreatePermissionTypeDto, UpdatePermissionTypeDto]{
		Name:       "IdentityPermissionType",
		Searchable: []string{"key", "name", "resource"},
	})
}
