package user

import "github.com/thescaffold/gox-packages/libs/core/crud"

type UserController struct {
	crud.CrudResource[User, CreateUserDto, UpdateUserDto]
	entity *UserEntity `inject:""`
}

func (c *UserController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[User, CreateUserDto, UpdateUserDto]{
		Name:       "IdentityUser",
		Searchable: []string{"email", "phone", "name"},
	})
}
