package role

import "github.com/thescaffold/gox-packages-core/crud"

type RoleController struct {
	crud.CrudResource[Role, CreateRoleDto, UpdateRoleDto]
	entity *RoleEntity `inject:""`
}

func (c *RoleController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Role, CreateRoleDto, UpdateRoleDto]{
		Name:       "IdentityRole",
		Searchable: []string{"client_id", "name", "type"},
	})
}
