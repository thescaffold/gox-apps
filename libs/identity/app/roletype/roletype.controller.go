package roletype

import "github.com/thescaffold/gox-packages/libs/core/crud"

type RoleTypeController struct {
	crud.CrudResource[RoleType, CreateRoleTypeDto, UpdateRoleTypeDto]
	entity *RoleTypeEntity `inject:""`
}

func (c *RoleTypeController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[RoleType, CreateRoleTypeDto, UpdateRoleTypeDto]{
		Name:       "IdentityRoleType",
		Searchable: []string{"name"},
	})
}
